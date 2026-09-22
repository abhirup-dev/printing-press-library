// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Route merge from
// endpoints.md (#3) — absent from the 25-endpoint sniffed spec.
// pp:data-source live

package cli

// organization list — GET /api/organization. Wrong-env tokens return an empty
// result here (not an error); pair with `auth check` when in doubt.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUptOrganizationCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Organizations visible to the caller",
		Example: "  unitypredict-pp-cli organization list\n  unitypredict-pp-cli organization list --env prod --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
			"pp:happy-args":  "",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "organization list")
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			env := resolveUptEnv(envFlag)
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			data, err := c.Get(ctx, "/api/organization", nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var items []map[string]any
			_ = json.Unmarshal(data, &items)
			if items == nil {
				items = make([]map[string]any, 0)
			}
			view := struct {
				Env   string           `json:"env"`
				Count int              `json:"count"`
				Note  string           `json:"note,omitempty"`
				Orgs  []map[string]any `json:"organizations"`
			}{Env: env, Count: len(items), Orgs: items}
			if len(items) == 0 {
				view.Note = "empty result — a wrong-env key returns empty here rather than an error; try auth check"
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No organizations visible.")
				return nil
			}
			return printAutoTable(cmd.OutOrStdout(), items)
		},
	}
	listCmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")

	parent := &cobra.Command{
		Use:   "organization",
		Short: "UnityPredict organizations",
		Example: "  unitypredict-pp-cli organization list",
	}
	parent.AddCommand(listCmd)
	return parent
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		addNovelCommandIfAbsent(root, newUptOrganizationCmd(flags))
	})
}
