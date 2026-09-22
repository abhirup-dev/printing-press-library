package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelLookupCmd(flags *rootFlags) *cobra.Command {
	var flagMarket string
	cmd := &cobra.Command{
		Use:         "lookup <query>",
		Short:       "Look up Indian or US stocks, funds, ETFs, and indices with live source metadata.",
		Example:     "  tickertape-pp-cli lookup VOO --market US --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:novel-feature": "true", "pp:happy-args": "query=VOO;--market=US"},
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := requireArg(cmd, args, "<query>")
			if err != nil {
				return err
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "lookup")
			}
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			market := strings.ToUpper(strings.TrimSpace(flagMarket))
			if market == "" {
				market = "IN"
			}
			results := map[string]any{"query": query, "market": market}
			switch market {
			case "IN", "INDIA", "INDIAN":
				results["market"] = "IN"
				results["result"] = fetchResearchEndpoint(ctx, c, "search_suggest", "/search/suggest", map[string]string{"q": query})
			case "US", "USA":
				results["market"] = "US"
				results["result"] = fetchResearchEndpoint(ctx, c, "us_security_info", "https://gms-api.tickertape.in/US/securities/info", map[string]string{"ticker": query})
			default:
				return usageErr(fmt.Errorf("unsupported --market %q (use IN or US)", flagMarket))
			}
			return printResearchResult(cmd, flags, "lookup", results)
		},
	}
	cmd.Flags().StringVar(&flagMarket, "market", "IN", "Market to search: IN or US.")
	return cmd
}
