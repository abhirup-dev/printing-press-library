package cli

import "github.com/spf13/cobra"

func newNovelCompanyBriefCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "brief <sid>",
		Short:       "Compose live company identity, summary, scorecard, reports, and access state.",
		Example:     "  tickertape-pp-cli company brief RELI --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:novel-feature": "true", "pp:happy-args": "sid=RELI"},
		RunE: func(cmd *cobra.Command, args []string) error {
			sid, err := requireArg(cmd, args, "<sid>")
			if err != nil {
				return err
			}
			if err := rejectSyntheticInvalidArg("sid", sid); err != nil {
				return err
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "company brief")
			}
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			paths := companyBriefPaths(sid)
			results := map[string]any{"sid": sid, "market": "IN", "endpoints": map[string]any{}}
			endpoints := results["endpoints"].(map[string]any)
			for name, path := range paths {
				endpoints[name] = fetchResearchEndpoint(ctx, c, name, path, nil)
			}
			return printResearchResult(cmd, flags, "company_brief", results)
		},
	}
	return cmd
}

func companyBriefPaths(sid string) map[string]string {
	return map[string]string{
		"info":      "/stocks/info/" + sid,
		"summary":   "/stocks/summary/" + sid,
		"scorecard": "https://analyze.api.tickertape.in/stocks/scorecard/" + sid,
		"ratings":   "/stocks/ratings/" + sid,
		"forecast":  "/stocks/estimates/forecast/" + sid,
		// AI summary is served by the analyze host, just like scorecard. Keeping
		// this absolute preserves the same route-specific auth/version handling
		// as `company ai-summary` in the compound fanout.
		"ai_summary":       "https://analyze.api.tickertape.in/stocks/aiSummary/" + sid,
		"aggregated_deals": "/stocks/aggregateddeals/" + sid,
	}
}
