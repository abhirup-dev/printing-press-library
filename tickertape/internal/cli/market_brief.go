package cli

import "github.com/spf13/cobra"

func newNovelMarketBriefCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "brief",
		Short:       "Turn the live Market Mood Index into a timestamped component brief.",
		Example:     "  tickertape-pp-cli market brief --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:novel-feature": "true", "pp:happy-args": "--agent"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "market brief")
			}
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			result := fetchResearchEndpoint(ctx, c, "mmi_now", "/mmi/now", nil)
			return printResearchResult(cmd, flags, "market_brief", map[string]any{
				"index":      result,
				"components": []string{"fear_greed", "fii", "momentum", "volatility"},
			})
		},
	}
	return cmd
}
