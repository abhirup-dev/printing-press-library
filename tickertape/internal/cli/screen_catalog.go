package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newNovelScreenCatalogCmd(flags *rootFlags) *cobra.Command {
	var flagAssetClass string
	cmd := &cobra.Command{
		Use:         "catalog",
		Short:       "Organize live screener filters, universes, and prebuilt screens with premium markers.",
		Example:     "  tickertape-pp-cli screen catalog --asset-class equity --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:novel-feature": "true", "pp:happy-args": "--asset-class=equity"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "screen catalog")
			}
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			assetClass := strings.ToLower(strings.TrimSpace(flagAssetClass))
			if assetClass == "" {
				assetClass = "all"
			}
			groups := map[string]map[string]string{
				"equity": {
					"filters": "/screener/filters", "prebuilt": "/screener/prebuilt", "universes": "/screener/universes",
				},
				"mutual-fund": {
					"filters": "/mf-screener/filters", "prebuilt": "/mf-screener/prebuilt", "universes": "/mf-screener/universes",
				},
			}
			if assetClass != "all" && assetClass != "equity" && assetClass != "mutual-fund" && assetClass != "mutualfund" {
				return usageErr(fmt.Errorf("unsupported --asset-class %q (use equity, mutual-fund, or all)", flagAssetClass))
			}
			if assetClass == "mutualfund" {
				assetClass = "mutual-fund"
			}
			endpoints := map[string]any{}
			for group, paths := range groups {
				if assetClass != "all" && assetClass != group {
					continue
				}
				for name, path := range paths {
					endpoints[group+"_"+name] = fetchResearchEndpoint(ctx, c, group+"_"+name, path, nil)
				}
			}
			return printResearchResult(cmd, flags, "screen_catalog", map[string]any{
				"asset_class":     assetClass,
				"endpoints":       endpoints,
				"premium_markers": "preserved in each provider response",
			})
		},
	}
	cmd.Flags().StringVar(&flagAssetClass, "asset-class", "all", "Asset class: equity, mutual-fund, or all.")
	return cmd
}
