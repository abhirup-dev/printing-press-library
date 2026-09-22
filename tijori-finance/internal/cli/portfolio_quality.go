// pp:data-source live
package cli

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"strings"
)

func newNovelPortfolioQualityCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "quality",
		Short:       "Join live portfolio weights with public company quality evidence and explicit unavailable states.",
		Example:     "  tijori-finance-pp-cli portfolio quality --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": ""},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "portfolio quality")
			}
			holdings, prov, err := novelGET(cmd.Context(), flags, "/api/portfolio/company_summary/", nil, nil, "portfolio-company-summary", true)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			quality := map[string]any{}
			for _, slug := range candidateSlugs(holdings) {
				page, _, pageErr := novelCompanyPage(cmd.Context(), flags, slug, "", "company")
				if pageErr != nil {
					quality[slug] = map[string]any{"status": "unavailable", "error": sanitizeError(pageErr)}
				} else {
					quality[slug] = map[string]any{"status": "available", "evidence": page}
				}
			}
			out := map[string]any{"holdings": holdings, "quality_by_slug": quality, "unmatched_holdings": true, "note": "Only explicit slug matches are enriched; inaccessible or slug-less holdings remain visible rather than being dropped."}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	return cmd
}

func candidateSlugs(raw json.RawMessage) []string {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	seen := map[string]bool{}
	out := []string{}
	var walk func(any)
	walk = func(x any) {
		switch y := x.(type) {
		case map[string]any:
			for k, val := range y {
				lk := strings.ToLower(k)
				if lk == "slug" || lk == "company_slug" {
					if s, ok := val.(string); ok && strings.TrimSpace(s) != "" && !seen[s] {
						seen[s] = true
						out = append(out, s)
						if len(out) >= 20 {
							return
						}
					}
				}
				walk(val)
			}
		case []any:
			for _, val := range y {
				walk(val)
				if len(out) >= 20 {
					return
				}
			}
		}
	}
	walk(v)
	return out
}
