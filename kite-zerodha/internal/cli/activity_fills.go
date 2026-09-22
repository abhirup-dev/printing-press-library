package cli

import (
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelActivityFillsCmd(flags *rootFlags) *cobra.Command {
	var orderID string
	cmd := &cobra.Command{
		Use:         "fills",
		Short:       "Show each order with the executions and partial fills it spawned.",
		Example:     "  kite-zerodha-pp-cli activity fills --order-id example-order-id --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "activity fills")
			}
			cmd, cancel := bindNovelCommand(cmd, flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			tradesPath := "/trades"
			if strings.TrimSpace(orderID) != "" {
				tradesPath = "/orders/" + url.PathEscape(strings.TrimSpace(orderID)) + "/trades"
			}
			tradesRaw, err := fetchNovelLive(cmd, flags, c, "trades", tradesPath, true)
			if err != nil {
				return err
			}
			trades, err := rawRecords(tradesRaw)
			if err != nil {
				return err
			}
			ordersRaw, err := fetchNovelLive(cmd, flags, c, "orders", "/orders", true)
			if err != nil {
				return err
			}
			orders, err := rawRecords(ordersRaw)
			if err != nil {
				return err
			}

			orderByID := make(map[string]map[string]any, len(orders))
			for _, order := range orders {
				id := stringValue(order, "order_id")
				if id != "" {
					orderByID[id] = order
				}
			}
			groupsByID := make(map[string][]map[string]any)
			for _, trade := range trades {
				id := stringValue(trade, "order_id")
				if id == "" {
					id = "unknown"
				}
				row := cloneRow(trade)
				row["turnover"] = turnover(trade)
				groupsByID[id] = append(groupsByID[id], row)
			}

			ids := make([]string, 0, len(groupsByID))
			for id := range groupsByID {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			groups := make([]map[string]any, 0, len(ids))
			var totalFills, totalQuantity int
			var totalTurnover float64
			for _, id := range ids {
				fills := groupsByID[id]
				var quantity, value float64
				for _, fill := range fills {
					quantity += numberValue(fill, "quantity")
					value += numberValue(fill, "turnover")
				}
				group := map[string]any{
					"order_id":        id,
					"fill_count":      len(fills),
					"filled_quantity": quantity,
					"turnover":        value,
					"partial_fill":    len(fills) > 1,
					"fills":           fills,
				}
				if order, ok := orderByID[id]; ok {
					orderQuantity := numberValue(order, "quantity")
					group["order"] = order
					group["order_quantity"] = orderQuantity
					group["quantity_shortfall"] = orderQuantity > quantity && orderQuantity > 0
					if orderQuantity > quantity && orderQuantity > 0 {
						group["partial_fill"] = true
					}
				}
				groups = append(groups, group)
				totalFills += len(fills)
				totalQuantity += int(quantity)
				totalTurnover += value
			}

			result := map[string]any{
				"as_of":  asOf(),
				"source": "official_kite_connect_live",
				"scope":  "current_day_trades",
				"filter": map[string]any{"order_id": strings.TrimSpace(orderID)},
				"summary": map[string]any{
					"order_count":     len(groups),
					"fill_count":      totalFills,
					"filled_quantity": totalQuantity,
					"turnover":        totalTurnover,
				},
				"orders": groups,
				"limitations": []string{
					"The broker endpoint exposes current-day executions; it is not a historical tradebook.",
					"partial_fill is inferred from multiple fills or a live order-quantity shortfall.",
				},
			}
			return printNovelLive(cmd, flags, result)
		},
	}
	cmd.Flags().StringVar(&orderID, "order-id", "", "Limit fills to one current-day order ID")
	return cmd
}
