package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"kite-zerodha-pp-cli/internal/config"
)

func newNovelSourceStatusCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "status",
		Short:       "Report which broker and deferred Console capabilities are safe to use.",
		Example:     "  kite-zerodha-pp-cli source status --agent --select capabilities",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "local"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "source status")
			}
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return err
			}
			authConfigured := cfg.CredentialConfigured()
			brokerStatus := "requires_authentication"
			if authConfigured {
				brokerStatus = "ready_for_authenticated_reads"
			}
			result := map[string]any{
				"as_of":  asOf(),
				"status": "ok",
				"source": "local_capability_report",
				"capabilities": map[string]any{
					"broker": map[string]any{
						"name":            "Kite Connect",
						"status":          brokerStatus,
						"base_url":        cfg.BaseURL,
						"auth_configured": authConfigured,
						"auth_boundary":   "KITE_CONNECT_TOKEN (api_key:access_token); value never emitted",
						"read_only":       true,
						"supported_live_reads": []string{
							"profile", "margins", "holdings", "positions", "orders", "order history", "trades", "order trades", "GTT list/detail", "instruments",
						},
					},
					"console": map[string]any{
						"status":                    "deferred",
						"historical_data_available": false,
						"validated_routes":          []string{"/reports/tradebook", "/reports/pnl", "/reports/taxpnl", "/reports/downloads"},
						"reason":                    "Console authentication replay and response schemas were not validated; no guessed historical adapter is enabled.",
						"read_only":                 true,
					},
				},
				"limitations": []string{
					"All broker data commands issue fresh official Kite Connect requests per invocation.",
					"No local response cache, sync database, historical ledger, browser session, or Console cookie replay is used.",
				},
			}
			// Keep the concise `--select sources` narrative form as an alias
			// for the explicit capability matrix.
			result["sources"] = result["capabilities"]
			raw, err := json.Marshal(result)
			if err != nil {
				return err
			}
			return printOutputWithFlagsMeta(cmd.OutOrStdout(), raw, flags, map[string]any{"source": "local"})
		},
	}
}
