package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newNovelActivitySellsCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "sells",
		Short:       "Summarize current-day sell executions with quantities, prices, and turnover.",
		Example:     "  kite-zerodha-pp-cli activity sells --agent --select trades",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "activity sells")
			}
			cmd, cancel := bindNovelCommand(cmd, flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			tradesRaw, err := fetchNovelLive(cmd, flags, c, "trades", "/trades", true)
			if err != nil {
				return err
			}
			allTrades, err := rawRecords(tradesRaw)
			if err != nil {
				return err
			}
			sells := make([]map[string]any, 0, len(allTrades))
			var quantity, turnoverTotal float64
			for _, trade := range allTrades {
				if strings.ToUpper(stringValue(trade, "transaction_type")) != "SELL" {
					continue
				}
				row := cloneRow(trade)
				rowTurnover := turnover(trade)
				row["turnover"] = rowTurnover
				row["side"] = "SELL"
				quantity += numberValue(trade, "quantity")
				turnoverTotal += rowTurnover
				sells = append(sells, row)
			}
			result := map[string]any{
				"as_of":  asOf(),
				"source": "official_kite_connect_live",
				"scope":  "current_day_trades",
				"summary": map[string]any{
					"sell_fill_count": len(sells),
					"sell_quantity":   quantity,
					"sell_turnover":   turnoverTotal,
				},
				"trades": sells,
				"limitations": []string{
					"Kite Connect /trades exposes current-day executions only.",
					"Historical sells require a separately validated Console adapter and are intentionally unavailable.",
				},
			}
			return printNovelLive(cmd, flags, result)
		},
	}
}
