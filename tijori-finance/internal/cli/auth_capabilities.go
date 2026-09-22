// pp:data-source live
package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"tijori-finance-pp-cli/internal/config"
)

func authCapabilityReport(ctx context.Context, flags *rootFlags) (map[string]any, error) {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return nil, err
	}
	endpoints := []struct {
		name, path string
		params     map[string]string
	}{
		{"company_financials", "/company/swiggy-ltd/financials/", nil},
		{"quarterly_results", "/results/quarterly-results/", nil},
		{"markets", "/in/markets", nil},
		{"portfolio_summary", "/api/portfolio/summary/", nil},
		{"portfolio_company_summary", "/api/portfolio/company_summary/", nil},
		{"alerts", "/api/alerts/", nil},
		{"timeline_upcoming", "/in/timeline/upcoming/self", nil},
	}
	probes := make([]novelProbe, 0, len(endpoints))
	for _, endpoint := range endpoints {
		probes = append(probes, probeNovelGET(ctx, flags, endpoint.path, endpoint.params, endpoint.name))
	}
	return map[string]any{
		"auth_configured":            cfg.CredentialConfigured() || os.Getenv("TIJORI_FINANCE_AUTH_HEADER") != "",
		"auth_source":                cfg.AuthSource,
		"credential_values_included": false,
		"endpoints":                  probes,
		"notes":                      []string{"Only GET probes were issued.", "HTML responses are classified as public page responses; JSON/authenticated state may still require a browser session."},
	}, nil
}

func newNovelAuthCapabilitiesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "capabilities",
		Short:       "Probe safe GET endpoints and report access state without exposing credentials.",
		Example:     "  tijori-finance-pp-cli auth capabilities --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": ""},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "auth capabilities")
			}
			out, err := authCapabilityReport(cmd.Context(), flags)
			if err != nil {
				return err
			}
			return printNovelResult(cmd, flags, jsonObject(out), DataProvenance{Source: "live", ResourceType: "auth-capabilities"})
		},
	}
	return cmd
}
