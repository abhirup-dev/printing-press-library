// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T5.
// pp:data-source live

package cli

// auth check validates credentials the only honest way on this API: list routes
// (/api/models/usermodels, /api/engines/userengines, /api/repository/search)
// return HTTP 200 for a garbage bearer, so they can never prove auth.
// GET /api/engines/{bare-uuid} separates the cases: 200 valid+owned, 401 bad
// key, 404 valid key + unknown id.

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"unitypredict-pp-cli/internal/client"
)

func newUptAuthCheckCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate credentials via GET /api/engines/{bare-uuid}",
		Long: "Validate the durable API key for an environment using the one route that cannot lie about auth.\n" +
			"List endpoints answer 200 for a garbage bearer; only GET /api/engines/{bare-uuid} separates a bad\n" +
			"key (401) from an unknown id (404). Exits 0 when authenticated, 4 when not.",
		Example: "  unitypredict-pp-cli auth check\n  unitypredict-pp-cli auth check --env prod",
		Annotations: map[string]string{
			"mcp:read-only":   "true",
			"pp:data-source":  "live",
			"pp:happy-args":   "--env=dev",
			"pp:typed-exit-codes": "0,4",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// No required inputs: bare invocation runs the check (prints help
			// only via -h). A help-only gate here would make "auth check" a
			// no-op without a flag — the exact trap this command exists to
			// remove.
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "auth check")
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			env := resolveUptEnv(envFlag)
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			// Pick a probe engine id owned by this account (list route is fine
			// for discovery — never for validation).
			probe := ""
			if data, err := c.Get(ctx, "/api/engines/userengines", map[string]string{"page": "1", "pagecount": "5"}); err == nil {
				var items []map[string]any
				_ = json.Unmarshal(data, &items)
				for _, it := range items {
					if id, ok := it["engineId"].(string); ok && id != "" {
						probe = stripEnginePrefix(id)
						break
					}
					if id, ok := it["id"].(string); ok && id != "" {
						probe = stripEnginePrefix(id)
						break
					}
				}
			}
			probeKind := "owned engine"
			if probe == "" {
				// Valid-format unknown uuid: 404 still proves the key authenticated
				// (a bad key would 401 before the lookup).
				probe = "00000000-0000-0000-0000-000000000000"
				probeKind = "unknown-id probe"
			}
			type probeResult struct {
				EngineID   string `json:"engine_id"`
				ProbeKind  string `json:"probe_kind"`
				HTTPStatus int    `json:"http_status"`
			}
			view := struct {
				Env              string      `json:"env"`
				BaseURL          string      `json:"base_url"`
				CredentialSource string      `json:"credential_source"`
				Authenticated    bool        `json:"authenticated"`
				Verdict          string      `json:"verdict"`
				Probe            probeResult `json:"probe"`
			}{
				Env:              env,
				CredentialSource: "~/.unitypredict/credentials[" + env + "]",
			}
			view.BaseURL = c.RequestBaseURL()
			view.Probe = probeResult{EngineID: probe, ProbeKind: probeKind}
			// APIError carries the HTTP status; success is implicitly 200.
			status := 200
			if _, err := c.Get(ctx, "/api/engines/"+probe, nil); err != nil {
				var apiErr *client.APIError
				if errors.As(err, &apiErr) {
					status = apiErr.StatusCode
				} else {
					return fmt.Errorf("probing /api/engines/%s: %w", probe, err)
				}
			}
			view.Probe.HTTPStatus = status
			switch {
			case status == 200:
				view.Authenticated = true
				view.Verdict = "authenticated (probe engine owned)"
			case status == 404:
				view.Authenticated = true
				view.Verdict = "authenticated (key valid; probe id not found — expected for the unknown-id probe)"
			case status == 401:
				view.Authenticated = false
				view.Verdict = "NOT authenticated: key rejected (401). Check UPT_API_KEY for this env in ~/.unitypredict/credentials"
			default:
				view.Verdict = fmt.Sprintf("inconclusive: HTTP %d", status)
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				if err := printJSONFiltered(cmd.OutOrStdout(), view, flags); err != nil {
					return err
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "env: %s  base: %s\n", view.Env, view.BaseURL)
				fmt.Fprintf(cmd.OutOrStdout(), "probe: %s (%s) -> HTTP %d\n", view.Probe.EngineID, view.Probe.ProbeKind, view.Probe.HTTPStatus)
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", view.Verdict)
			}
			if !view.Authenticated {
				return authErr(fmt.Errorf("%s", view.Verdict))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant to validate: dev or prod (default $UNITYPREDICT_ENV or dev)")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if authCmd, _, err := root.Find([]string{"auth"}); err == nil {
			addNovelCommandIfAbsent(authCmd, newUptAuthCheckCmd(flags))
		}
	})
}
