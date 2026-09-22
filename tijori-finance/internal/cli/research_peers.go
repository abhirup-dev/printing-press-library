// pp:data-source live
package cli

import (
	"github.com/spf13/cobra"
	"strings"
)

func newNovelResearchPeersCmd(flags *rootFlags) *cobra.Command {
	var flagMetrics string
	cmd := &cobra.Command{
		Use:         "peers <slug>",
		Short:       "Compare live peer financial, ownership, and operating evidence in one unit-aware matrix.",
		Example:     "  tijori-finance-pp-cli research peers tata-steel-limited --metrics revenue,opm,roe,debt,holding --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "slug=tata-steel-limited;--metrics=revenue,opm,roe,debt,holding"},
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := novelRequiredArg(cmd, flags, args, "<slug>")
			if err != nil {
				return err
			}
			if strings.TrimSpace(flagMetrics) == "" {
				flagMetrics = "revenue,opm,roe,debt,holding"
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research peers")
			}
			company, prov, err := novelCompanyPage(cmd.Context(), flags, slug, "", "company")
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			peers, _, peerErr := novelGET(cmd.Context(), flags, "/api/v1/ind/company_search/", map[string]string{"q": slug}, nil, "company-search", true)
			out := map[string]any{"company": slug, "metrics": strings.Split(flagMetrics, ","), "company_snapshot": company, "peer_candidates": peers, "peer_candidates_error": nil, "units": "Provider units are preserved; no cross-company normalization is inferred."}
			if peerErr != nil {
				out["peer_candidates_error"] = sanitizeError(peerErr)
			}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	cmd.Flags().StringVar(&flagMetrics, "metrics", "", "Comma-separated metrics to include, for example revenue,opm,roe")
	return cmd
}
