// pp:data-source live
package cli

import (
	"github.com/spf13/cobra"
)

func newNovelResearchContextCmd(flags *rootFlags) *cobra.Command {
	var flagSector bool
	cmd := &cobra.Command{
		Use:         "context <slug>",
		Short:       "Bundle dated company, sector, macro, and raw-material observations without a synthetic verdict.",
		Example:     "  tijori-finance-pp-cli research context deepak-nitrite-limited --sector --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "slug=deepak-nitrite-limited;--sector"},
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := novelRequiredArg(cmd, flags, args, "<slug>")
			if err != nil {
				return err
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research context")
			}
			company, prov, err := novelCompanyPage(cmd.Context(), flags, slug, "", "company")
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			out := map[string]any{"company": slug, "sector_requested": flagSector, "company_snapshot": company, "observations": []novelProbe{probeNovelGET(cmd.Context(), flags, "/in/macro", nil, "macro"), probeNovelGET(cmd.Context(), flags, "/in/raw-materials", nil, "raw-materials")}, "verdict": nil}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	cmd.Flags().BoolVar(&flagSector, "sector", false, "Include the sector context request in the evidence bundle")
	return cmd
}
