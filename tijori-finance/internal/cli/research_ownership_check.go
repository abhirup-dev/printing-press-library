// pp:data-source live
package cli

import (
	"github.com/spf13/cobra"
	"strings"
)

func newNovelResearchOwnershipCheckCmd(flags *rootFlags) *cobra.Command {
	var flagPeriod string
	cmd := &cobra.Command{
		Use:         "ownership-check <slug>",
		Short:       "Align same-period ownership movements with financial and cash-flow evidence.",
		Example:     "  tijori-finance-pp-cli research ownership-check tata-steel-limited --period Jun-26 --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "slug=tata-steel-limited;--period=Jun-26"},
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := novelRequiredArg(cmd, flags, args, "<slug>")
			if err != nil {
				return err
			}
			if strings.TrimSpace(flagPeriod) == "" {
				flagPeriod = "latest"
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research ownership-check")
			}
			ownership, prov, err := novelCompanyPage(cmd.Context(), flags, slug, "shareholding/", "shareholding")
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			financials, _, financialErr := novelCompanyPage(cmd.Context(), flags, slug, "financials/", "company-financials")
			out := map[string]any{"company": slug, "period": flagPeriod, "ownership": ownership, "financials": financials, "financials_error": nil, "interpretation": nil}
			if financialErr != nil {
				out["financials_error"] = sanitizeError(financialErr)
			}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	cmd.Flags().StringVar(&flagPeriod, "period", "", "Period label to align across ownership and financial tables")
	return cmd
}
