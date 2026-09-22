// pp:data-source live
package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelScreenerRunCmd(flags *rootFlags) *cobra.Command {
	var flagQuery string
	var flagExplain bool
	cmd := &cobra.Command{
		Use:         "run",
		Short:       "Run a native Tijori screener query and optionally explain the request.",
		Example:     "  tijori-finance-pp-cli screener run --query '( ROE > 15 )' --explain --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "--query=( ROE > 15 );--explain"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(flagQuery) == "" {
				return usageErr(fmt.Errorf("--query is required"))
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "screener run")
			}
			raw, prov, err := novelGET(cmd.Context(), flags, "/filter/", map[string]string{"fq": flagQuery}, map[string]string{"X-Printing-Press-HTML-Response": "true"}, "screener", false)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			result := map[string]any{
				"query":          flagQuery,
				"explain":        flagExplain,
				"matched_fields": screenerQueryFields(flagQuery),
				"request":        map[string]any{"method": "GET", "path": "/filter/", "query": map[string]string{"fq": flagQuery}, "source_url": "https://www.tijorifinance.com/filter/?fq=" + url.QueryEscape(flagQuery)},
			}
			if json.Valid(raw) {
				result["results"] = json.RawMessage(raw)
			} else {
				result["results"] = []any{}
				result["response"] = map[string]any{"content": "html", "note": "Tijori returned a page response; results are not parsed from HTML."}
			}
			return printNovelResult(cmd, flags, jsonObject(result), prov)
		},
	}
	cmd.Flags().StringVar(&flagQuery, "query", "", "Native Tijori financial/business-data filter expression")
	cmd.Flags().BoolVar(&flagExplain, "explain", false, "Include the submitted request, matched fields, and response shape")
	return cmd
}

func screenerQueryFields(query string) []string {
	known := []string{"roe", "roce", "revenue", "profit", "opm", "ebitda", "debt", "market capitalization", "market cap", "pe", "dividend yield"}
	lower := strings.ToLower(query)
	out := make([]string, 0, len(known))
	for _, field := range known {
		if strings.Contains(lower, field) {
			out = append(out, field)
		}
	}
	return out
}
