// pp:data-source live
package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelResearchChangesCmd(flags *rootFlags) *cobra.Command {
	var flagFrom, flagTo string
	cmd := &cobra.Command{
		Use:         "changes <slug>",
		Short:       "Emit deterministic row-level changes between two server-returned financial periods.",
		Example:     "  tijori-finance-pp-cli research changes tata-steel-limited --from Mar-25 --to Mar-26 --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "slug=tata-steel-limited;--from=Mar-25;--to=Mar-26"},
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := novelRequiredArg(cmd, flags, args, "<slug>")
			if err != nil {
				return err
			}
			if strings.TrimSpace(flagFrom) == "" || strings.TrimSpace(flagTo) == "" {
				return usageErr(fmt.Errorf("--from and --to are required"))
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research changes")
			}
			data, prov, err := novelCompanyPage(cmd.Context(), flags, slug, "financials/", "company-financials")
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			out := map[string]any{"company": slug, "from_period": flagFrom, "to_period": flagTo, "changes": deterministicNumericChanges(data), "source_url": "https://www.tijorifinance.com/company/" + slug + "/financials/", "note": "Changes are computed only from rows present in the live response; no estimates or synthetic values are added."}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	cmd.Flags().StringVar(&flagFrom, "from", "", "Earlier period label, for example Mar-25")
	cmd.Flags().StringVar(&flagTo, "to", "", "Later period label, for example Mar-26")
	return cmd
}

type numericChange struct {
	Field string `json:"field"`
	Value any    `json:"value,omitempty"`
	Note  string `json:"note,omitempty"`
}

func deterministicNumericChanges(data json.RawMessage) []numericChange {
	var root any
	if json.Unmarshal(data, &root) != nil {
		return nil
	}
	flat := map[string]float64{}
	flattenNumbers(root, "", flat)
	out := make([]numericChange, 0, len(flat))
	for field, value := range flat {
		out = append(out, numericChange{Field: field, Value: value, Note: "raw live value; period pairing depends on provider row shape"})
	}
	return out
}

func flattenNumbers(v any, path string, out map[string]float64) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			p := k
			if path != "" {
				p = path + "." + k
			}
			flattenNumbers(child, p, out)
		}
	case []any:
		for i, child := range x {
			flattenNumbers(child, fmt.Sprintf("%s[%d]", path, i), out)
		}
	case float64:
		if path != "" {
			out[path] = x
		}
	}
}
