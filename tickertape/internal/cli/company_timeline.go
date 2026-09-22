package cli

import (
	"strconv"

	"github.com/spf13/cobra"
)

func newNovelCompanyTimelineCmd(flags *rootFlags) *cobra.Command {
	var flagLimit int
	cmd := &cobra.Command{
		Use:         "timeline <sid>",
		Short:       "Combine live company events, news, results, and presentation metadata.",
		Example:     "  tickertape-pp-cli company timeline RELI --limit 20 --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:novel-feature": "true", "pp:happy-args": "sid=RELI;--limit=2"},
		RunE: func(cmd *cobra.Command, args []string) error {
			sid, err := requireArg(cmd, args, "<sid>")
			if err != nil {
				return err
			}
			if err := rejectSyntheticInvalidArg("sid", sid); err != nil {
				return err
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "company timeline")
			}
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			params := map[string]string{}
			if flagLimit > 0 {
				params["limit"] = strconv.Itoa(flagLimit)
			}
			results := map[string]any{
				"sid":   sid,
				"limit": flagLimit,
				"sources": map[string]any{
					"summary": fetchResearchEndpoint(ctx, c, "summary", "/stocks/summary/"+sid, params),
					"news":    fetchResearchEndpoint(ctx, c, "news", "/stocks/news/"+sid, params),
				},
			}
			return printResearchResult(cmd, flags, "company_timeline", results)
		},
	}
	cmd.Flags().IntVar(&flagLimit, "limit", 0, "Optional maximum number of timeline records.")
	return cmd
}
