// pp:data-source live
package cli

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

func newNovelFeedInspectCmd(flags *rootFlags) *cobra.Command {
	var flagCompany string
	cmd := &cobra.Command{
		Use:         "inspect",
		Short:       "Resolve timeline items to normalized event, report, and document metadata while preserving cursors.",
		Example:     "  tijori-finance-pp-cli feed inspect --company tata-steel-limited --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "--company=tata-steel-limited"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(flagCompany) == "" {
				return usageErr(fmt.Errorf("--company is required"))
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "feed inspect")
			}
			params := map[string]string{"company": flagCompany, "timestamp": "0", "timelinetype": "all"}
			timeline, prov, timelineErr := novelGET(cmd.Context(), flags, "/in/timeline/self/more", params, map[string]string{"X-Printing-Press-HTML-Response": "true"}, "timeline", true)
			var timelineResult any = timeline
			if timelineErr != nil {
				prov = DataProvenance{Source: "live", ResourceType: "timeline"}
				timelineResult = probeNovelGET(cmd.Context(), flags, "/in/timeline/self/more", params, "timeline")
			}
			upcoming, _, upcomingErr := novelGET(cmd.Context(), flags, "/in/timeline/upcoming/self", map[string]string{"company": flagCompany}, map[string]string{"X-Printing-Press-HTML-Response": "true"}, "timeline-upcoming", true)
			var upcomingResult any = upcoming
			if upcomingErr != nil {
				upcomingResult = probeNovelGET(cmd.Context(), flags, "/in/timeline/upcoming/self", map[string]string{"company": flagCompany}, "timeline-upcoming")
			}
			out := map[string]any{"company": flagCompany, "timeline": timelineResult, "upcoming": upcomingResult, "upcoming_error": nil, "timeline_error": nil, "cursor": extractCursor(timeline), "normalization": "raw provider event metadata is preserved; documents are not downloaded implicitly"}
			if timelineErr != nil {
				out["timeline_error"] = sanitizeError(timelineErr)
			}
			if upcomingErr != nil {
				out["upcoming_error"] = sanitizeError(upcomingErr)
			}
			if timelineErr != nil || upcomingErr != nil {
				availability := "upstream_error"
				cause := timelineErr
				if cause == nil {
					cause = upcomingErr
				}
				causeText := strings.ToLower(sanitizeError(cause))
				if strings.Contains(causeText, "empty response body") ||
					(strings.Contains(causeText, "html instead of json") &&
						!strings.Contains(causeText, "not authenticated") &&
						!strings.Contains(causeText, "expired")) {
					availability = "empty"
				}
				out["availability"] = availability
				out["endpoint"] = "/in/timeline/self/more"
				out["observed_at"] = time.Now().UTC().Format(time.RFC3339)
				out["stability"] = "unstable"
				out["cause"] = sanitizeError(cause)
				envelope := map[string]any{
					"meta":    map[string]any{"source": "live", "resource": "timeline", "observed_at": out["observed_at"]},
					"results": out,
				}
				jsonMode := flags.asJSON || flags.agent
				if flag := cmd.Flags().Lookup("json"); flag != nil && flag.Changed {
					jsonMode = true
				}
				if flag := cmd.InheritedFlags().Lookup("json"); flag != nil && flag.Changed {
					jsonMode = true
				}
				if jsonMode {
					if err := json.NewEncoder(cmd.OutOrStdout()).Encode(envelope); err != nil {
						return err
					}
				}
				if availability == "empty" {
					return nil
				}
				return cause
			}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	cmd.Flags().StringVar(&flagCompany, "company", "", "Company slug or identifier used to scope timeline reads")
	return cmd
}

func extractCursor(raw json.RawMessage) any {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	var found any
	var walk func(any)
	walk = func(x any) {
		switch y := x.(type) {
		case map[string]any:
			for k, val := range y {
				if strings.Contains(strings.ToLower(k), "cursor") || strings.Contains(strings.ToLower(k), "next") {
					if found == nil {
						found = val
					}
				}
				walk(val)
			}
		case []any:
			for _, val := range y {
				walk(val)
			}
		}
	}
	walk(v)
	return found
}
