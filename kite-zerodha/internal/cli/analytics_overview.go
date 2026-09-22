package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelAnalyticsOverviewCmd(flags *rootFlags) *cobra.Command {
	var by string
	cmd := &cobra.Command{
		Use:         "overview",
		Short:       "Compute current P&L, concentration, turnover, and unrealized win/loss summaries from live rows.",
		Example:     "  kite-zerodha-pp-cli analytics overview --by symbol --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE: func(cmd *cobra.Command, args []string) error {
			groupBy := normalizeGroup(by)
			if groupBy != "symbol" && groupBy != "exchange" && groupBy != "product" {
				return usageErrf("invalid --by %q; use symbol, exchange, or product", by)
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "analytics overview")
			}
			cmd, cancel := bindNovelCommand(cmd, flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			holdingsRaw, err := fetchNovelLive(cmd, flags, c, "holdings", "/portfolio/holdings", true)
			if err != nil {
				return err
			}
			positionsRaw, err := fetchNovelLive(cmd, flags, c, "positions", "/portfolio/positions", false)
			if err != nil {
				return err
			}
			tradesRaw, err := fetchNovelLive(cmd, flags, c, "trades", "/trades", true)
			if err != nil {
				return err
			}
			holdings, err := rawRecords(holdingsRaw)
			if err != nil {
				return err
			}
			positions, err := rawObject(positionsRaw)
			if err != nil {
				return err
			}
			trades, err := rawRecords(tradesRaw)
			if err != nil {
				return err
			}
			groups := map[string]map[string]any{}
			getGroup := func(row map[string]any) map[string]any {
				key := analyticsGroupKey(row, groupBy)
				group := groups[key]
				if group == nil {
					group = map[string]any{
						"key": key, "holdings_count": 0, "position_count": 0,
						"trade_count": 0, "quantity": 0.0, "market_value": 0.0,
						"unrealized_pnl": 0.0, "turnover": 0.0,
						"buy_quantity": 0.0, "sell_quantity": 0.0,
						"buy_turnover": 0.0, "sell_turnover": 0.0,
					}
					groups[key] = group
				}
				return group
			}

			var holdingsPnl, positionsPnl, buyTurnover, sellTurnover, currentTurnover float64
			var winning, losing, flat int
			for _, row := range holdings {
				group := getGroup(row)
				pnl := pnlValue(row)
				value := rowMarketValue(row)
				group["holdings_count"] = intValue(group["holdings_count"]) + 1
				group["quantity"] = numberValue(group, "quantity") + rowQuantity(row)
				group["market_value"] = numberValue(group, "market_value") + value
				group["unrealized_pnl"] = numberValue(group, "unrealized_pnl") + pnl
				holdingsPnl += pnl
				if pnl > 0 {
					winning++
				} else if pnl < 0 {
					losing++
				} else {
					flat++
				}
			}
			for _, row := range rawArrayField(positions, "net") {
				group := getGroup(row)
				pnl := pnlValue(row)
				group["position_count"] = intValue(group["position_count"]) + 1
				group["quantity"] = numberValue(group, "quantity") + rowQuantity(row)
				group["market_value"] = numberValue(group, "market_value") + rowMarketValue(row)
				group["unrealized_pnl"] = numberValue(group, "unrealized_pnl") + pnl
				positionsPnl += pnl
				if pnl > 0 {
					winning++
				} else if pnl < 0 {
					losing++
				} else {
					flat++
				}
			}
			for _, row := range trades {
				group := getGroup(row)
				value := turnover(row)
				quantity := numberValue(row, "quantity")
				group["trade_count"] = intValue(group["trade_count"]) + 1
				group["turnover"] = numberValue(group, "turnover") + value
				group["quantity"] = numberValue(group, "quantity") + quantity
				currentTurnover += value
				if strings.EqualFold(stringValue(row, "transaction_type"), "BUY") {
					group["buy_quantity"] = numberValue(group, "buy_quantity") + quantity
					group["buy_turnover"] = numberValue(group, "buy_turnover") + value
					buyTurnover += value
				} else if strings.EqualFold(stringValue(row, "transaction_type"), "SELL") {
					group["sell_quantity"] = numberValue(group, "sell_quantity") + quantity
					group["sell_turnover"] = numberValue(group, "sell_turnover") + value
					sellTurnover += value
				}
			}

			keys := make([]string, 0, len(groups))
			for key := range groups {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			groupRows := make([]map[string]any, 0, len(keys))
			var concentrationBase float64
			for _, key := range keys {
				concentrationBase += numberValue(groups[key], "market_value")
			}
			if concentrationBase < 0 {
				concentrationBase = -concentrationBase
			}
			for _, key := range keys {
				row := groups[key]
				value := numberValue(row, "market_value")
				if concentrationBase > 0 {
					row["concentration_pct"] = value / concentrationBase * 100
				} else {
					row["concentration_pct"] = 0.0
				}
				groupRows = append(groupRows, row)
			}

			result := map[string]any{
				"as_of": asOf(), "source": "official_kite_connect_live", "scope": "current_live_state",
				"group_by": groupBy,
				"metrics": map[string]any{
					"holdings_unrealized_pnl":  holdingsPnl,
					"positions_unrealized_pnl": positionsPnl,
					"total_unrealized_pnl":     holdingsPnl + positionsPnl,
					"current_day_turnover":     currentTurnover,
					"buy_turnover":             buyTurnover, "sell_turnover": sellTurnover,
					"winning_unrealized_positions": winning,
					"losing_unrealized_positions":  losing,
					"flat_unrealized_positions":    flat,
				},
				"groups": groupRows,
				"limitations": []string{
					"P&L and win/loss figures are unrealized/current-state aggregates, not realized historical results.",
					"Turnover is limited to current-day executions returned by Kite Connect.",
					"Historical statements, charges, taxes, and corporate actions are intentionally excluded pending Console validation.",
				},
			}
			return printNovelLive(cmd, flags, result)
		},
	}
	cmd.Flags().StringVar(&by, "by", "", "Group by symbol, exchange, or product (default: symbol)")
	return cmd
}

func analyticsGroupKey(row map[string]any, by string) string {
	var value string
	switch by {
	case "exchange":
		value = stringValue(row, "exchange")
	case "product":
		value = stringValue(row, "product")
	default:
		value = stringValue(row, "tradingsymbol", "symbol")
	}
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func intValue(value any) int {
	switch n := value.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func usageErrf(format string, args ...any) error { return usageErr(fmt.Errorf(format, args...)) }
