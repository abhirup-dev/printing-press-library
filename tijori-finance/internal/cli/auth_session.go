// pp:data-source live
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"tijori-finance-pp-cli/internal/config"
)

func newNovelAuthSessionCommands(flags *rootFlags) []*cobra.Command {
	status := &cobra.Command{
		Use:         "status",
		Short:       "Show sanitized auth-store state and probe live capabilities.",
		Example:     "  tijori-finance-pp-cli auth status --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": ""},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "auth status")
			}
			capabilities, err := authCapabilityReport(cmd.Context(), flags)
			if err != nil {
				return err
			}
			capabilities["store"] = config.AuthStoreMetadata()
			capabilities["environment_override"] = os.Getenv("TIJORI_FINANCE_AUTH_HEADER") != ""
			return printNovelResult(cmd, flags, jsonObject(capabilities), DataProvenance{Source: "live", ResourceType: "auth-status"})
		},
	}

	var fromStdin bool
	importCommand := &cobra.Command{
		Use:         "import",
		Short:       "Read one auth-header line from stdin and store it securely.",
		Example:     "  <approved-browser-session-pipe> | tijori-finance-pp-cli auth import --stdin --json",
		Annotations: map[string]string{"mcp:read-only": "false", "pp:data-source": "computed", "pp:happy-args": "--stdin"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !fromStdin {
				return usageErr(errors.New("auth import requires --stdin; pipe a browser-established auth header without printing it"))
			}
			data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 8192))
			if err != nil {
				return fmt.Errorf("read auth header from stdin: %w", err)
			}
			value := strings.TrimSpace(string(data))
			if value == "" {
				return usageErr(errors.New("stdin must contain one auth header or session-material JSON object"))
			}
			material := config.SessionMaterial{}
			if strings.HasPrefix(value, "{") {
				var payload struct {
					AuthHeader    string            `json:"auth_header"`
					Authorization string            `json:"authorization"`
					Headers       map[string]string `json:"headers"`
				}
				if err := json.Unmarshal([]byte(value), &payload); err != nil {
					return usageErr(errors.New("stdin session material is not valid JSON"))
				}
				material.AuthHeader = payload.AuthHeader
				if material.AuthHeader == "" {
					material.AuthHeader = payload.Authorization
				}
				material.Headers = payload.Headers
			} else {
				if strings.ContainsAny(value, "\r\n") {
					return usageErr(errors.New("stdin auth header must be exactly one line"))
				}
				material.AuthHeader = value
			}
			backend, err := config.StoreAuthMaterial(material)
			if err != nil {
				return err
			}
			return flags.printJSON(cmd, map[string]any{"stored": true, "backend": backend, "credential_values_included": false})
		},
	}
	importCommand.Flags().BoolVar(&fromStdin, "stdin", false, "Read the auth header from stdin")

	logout := &cobra.Command{
		Use:         "logout",
		Short:       "Remove the stored auth header from the secure store.",
		Example:     "  tijori-finance-pp-cli auth logout --json",
		Annotations: map[string]string{"mcp:read-only": "false", "pp:data-source": "computed", "pp:happy-args": ""},
		RunE: func(cmd *cobra.Command, args []string) error {
			backend, err := config.ClearStoredAuthHeader()
			if err != nil {
				return err
			}
			return flags.printJSON(cmd, map[string]any{"cleared": true, "backend": backend, "environment_override_active": os.Getenv("TIJORI_FINANCE_AUTH_HEADER") != ""})
		},
	}
	return []*cobra.Command{status, importCommand, logout}
}
