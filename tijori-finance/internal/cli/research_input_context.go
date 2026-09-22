// pp:data-source live
package cli

import "github.com/spf13/cobra"

func newNovelResearchInputContextCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "input-context <slug>",
		Short:       "Place current raw-material observations beside live operational and revenue-mix data.",
		Example:     "  tijori-finance-pp-cli research input-context deepak-nitrite-limited --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "slug=deepak-nitrite-limited"},
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := novelRequiredArg(cmd, flags, args, "<slug>")
			if err != nil {
				return err
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research input-context")
			}
			company, prov, err := novelCompanyPage(cmd.Context(), flags, slug, "", "company")
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			out := map[string]any{"company": slug, "company_snapshot": company, "inputs": map[string]any{"raw_materials": probeNovelGET(cmd.Context(), flags, "/in/raw-materials", nil, "raw-materials"), "operational_metrics": probeNovelGET(cmd.Context(), flags, "/company/"+slug+"/", nil, "operational-metrics"), "revenue_mix": probeNovelGET(cmd.Context(), flags, "/company/"+slug+"/", nil, "revenue-mix")}, "verdict": nil}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	return cmd
}
