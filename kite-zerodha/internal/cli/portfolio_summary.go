package cli

import "github.com/spf13/cobra"

func newNovelPortfolioSummaryCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "summary",
		Short:       "Combine live holdings, positions, and margin state into one current portfolio view.",
		Example:     "  kite-zerodha-pp-cli portfolio summary --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "portfolio summary")
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
			marginsRaw, err := fetchNovelLive(cmd, flags, c, "margins", "/user/margins", false)
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
			margins, err := rawObject(marginsRaw)
			if err != nil {
				return err
			}

			var marketValue, cost, holdingsPnl float64
			for i, row := range holdings {
				value := rowMarketValue(row)
				rowCostValue := rowCost(row)
				pnl := pnlValue(row)
				marketValue += value
				cost += rowCostValue
				holdingsPnl += pnl
				row["market_value"] = value
				row["cost_value"] = rowCostValue
				row["unrealized_pnl"] = pnl
				holdings[i] = row
			}

			netPositions := rawArrayField(positions, "net")
			dayPositions := rawArrayField(positions, "day")
			var positionsPnl float64
			for _, row := range netPositions {
				positionsPnl += pnlValue(row)
			}
			for i, row := range netPositions {
				row["unrealized_pnl"] = pnlValue(row)
				netPositions[i] = row
			}

			result := map[string]any{
				"as_of":  asOf(),
				"source": "official_kite_connect_live",
				"scope":  "current_live_state",
				"summary": map[string]any{
					"holdings_count":              len(holdings),
					"net_positions_count":         len(netPositions),
					"day_positions_count":         len(dayPositions),
					"holdings_market_value":       marketValue,
					"holdings_cost_value":         cost,
					"holdings_unrealized_pnl":     holdingsPnl,
					"positions_unrealized_pnl":    positionsPnl,
					"total_unrealized_pnl":        holdingsPnl + positionsPnl,
					"portfolio_value_is_holdings": true,
				},
				"holdings":  holdings,
				"positions": map[string]any{"net": netPositions, "day": dayPositions},
				"margins":   margins,
				"limitations": []string{
					"This is a live snapshot, not a historical ledger.",
					"Console tradebook, P&L, tax P&L, charges, and corporate actions are not included.",
				},
			}
			return printNovelLive(cmd, flags, result)
		},
	}
}
