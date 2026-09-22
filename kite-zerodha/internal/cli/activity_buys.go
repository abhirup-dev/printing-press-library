package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newNovelActivityBuysCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "buys",
		Short:       "Summarize current-day buy executions with quantities, prices, and turnover.",
		Example:     "  kite-zerodha-pp-cli activity buys --agent --select trades",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "activity buys")
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
			buys := make([]map[string]any, 0, len(allTrades))
			var quantity, turnoverTotal float64
			for _, trade := range allTrades {
				if strings.ToUpper(stringValue(trade, "transaction_type")) != "BUY" {
					continue
				}
				row := cloneRow(trade)
				rowTurnover := turnover(trade)
				row["turnover"] = rowTurnover
				row["side"] = "BUY"
				quantity += numberValue(trade, "quantity")
				turnoverTotal += rowTurnover
				buys = append(buys, row)
			}
			result := map[string]any{
				"as_of":  asOf(),
				"source": "official_kite_connect_live",
				"scope":  "current_day_trades",
				"summary": map[string]any{
					"buy_fill_count": len(buys),
					"buy_quantity":   quantity,
					"buy_turnover":   turnoverTotal,
				},
				"trades": buys,
				"limitations": []string{
					"Kite Connect /trades exposes current-day executions only.",
					"Historical buys require a separately validated Console adapter and are intentionally unavailable.",
				},
			}
			return printNovelLive(cmd, flags, result)
		},
	}
}
