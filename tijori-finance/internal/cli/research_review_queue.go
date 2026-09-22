// pp:data-source live
package cli

import "github.com/spf13/cobra"

func newNovelResearchReviewQueueCmd(flags *rootFlags) *cobra.Command {
	var flagPortfolio bool
	cmd := &cobra.Command{
		Use:         "review-queue",
		Short:       "Join current holdings or watchlist names with upcoming results, events, reports, and alerts.",
		Example:     "  tijori-finance-pp-cli research review-queue --portfolio --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "--portfolio"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research review-queue")
			}
			out := map[string]any{"portfolio_requested": flagPortfolio, "queue": []any{}, "sources": map[string]any{"upcoming": probeNovelGET(cmd.Context(), flags, "/in/timeline/upcoming/self", nil, "timeline-upcoming"), "timeline": probeNovelGET(cmd.Context(), flags, "/in/timeline/self/more", nil, "timeline"), "reports": probeNovelGET(cmd.Context(), flags, "/financials/", nil, "reports"), "alerts": probeNovelGET(cmd.Context(), flags, "/api/alerts/", nil, "alerts")}}
			if flagPortfolio {
				out["holdings"] = probeNovelGET(cmd.Context(), flags, "/api/portfolio/company_summary/", nil, "portfolio-company-summary")
			} else {
				out["holdings"] = map[string]any{"status": "not-requested"}
			}
			return printNovelResult(cmd, flags, jsonObject(out), DataProvenance{Source: "live", ResourceType: "review-queue"})
		},
	}
	cmd.Flags().BoolVar(&flagPortfolio, "portfolio", false, "Include the authenticated portfolio company summary when available")
	return cmd
}
