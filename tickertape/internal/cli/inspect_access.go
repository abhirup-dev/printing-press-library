package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelInspectAccessCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "access <domain> <route> [value]",
		Short:       "Explain live source, observation time, identity, lock state, and entitlement errors.",
		Example:     "  tickertape-pp-cli inspect access company scorecard RELI --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:novel-feature": "true", "pp:happy-args": "domain=company;route=scorecard;value=RELI"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return usageErr(fmt.Errorf("domain and route are required\nUsage: %s <domain> <route> [value]", cmd.CommandPath()))
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "inspect access")
			}
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			domain := strings.ToLower(args[0])
			route := strings.ToLower(args[1])
			path, params, err := accessInspectionTarget(domain, route, args[2:])
			if err != nil {
				return usageErr(err)
			}
			result := fetchResearchEndpoint(ctx, c, domain+"_"+route, path, params)
			host := c.BaseURL
			if parsed, parseErr := url.Parse(path); parseErr == nil && parsed.Host != "" {
				host = parsed.Scheme + "://" + parsed.Host
			} else if parsed, parseErr := url.Parse(c.BaseURL); parseErr == nil {
				host = parsed.Scheme + "://" + parsed.Host
			}
			return printResearchResult(cmd, flags, "access_inspection", map[string]any{
				"target":       map[string]any{"domain": domain, "route": route, "path": path, "params": params},
				"source_host":  host,
				"access_state": result["access_state"],
				"observed_at":  result["observed_at"],
				"result":       result,
			})
		},
	}
	return cmd
}

func accessInspectionTarget(domain, route string, args []string) (string, map[string]string, error) {
	if len(args) == 0 && domain == "market" && route == "mood" {
		return "/mmi/now", nil, nil
	}
	if len(args) == 0 && domain == "search" && route == "suggest" {
		return "", nil, fmt.Errorf("search suggest requires a query")
	}
	if len(args) == 0 {
		return "", nil, fmt.Errorf("%s %s requires an identifier", domain, route)
	}
	sid := args[0]
	if domain == "search" && route == "suggest" {
		return "/search/suggest", map[string]string{"q": sid}, nil
	}
	if domain != "company" {
		return "", nil, fmt.Errorf("unsupported domain %q", domain)
	}
	paths := map[string]string{
		"info":             "/stocks/info/" + sid,
		"summary":          "/stocks/summary/" + sid,
		"scorecard":        "https://analyze.api.tickertape.in/stocks/scorecard/" + sid,
		"ratings":          "/stocks/ratings/" + sid,
		"forecast":         "/stocks/estimates/forecast/" + sid,
		"ai-summary":       "/stocks/aiSummary/" + sid,
		"aggregated-deals": "/stocks/aggregateddeals/" + sid,
	}
	path, ok := paths[route]
	if !ok {
		return "", nil, fmt.Errorf("unsupported company route %q", route)
	}
	return path, nil, nil
}
