// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Route merge from
// pp:data-source live
// endpoints.md (#19) — absent from the 25-endpoint sniffed spec.

package cli

// models usage — GET /api/models/{modelId}/usageStats: earnings and inference
// stats by year-month. Response shape varies; passthrough JSON plus a best
// effort table when rows are recognizable.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUptModelsUsageCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	cmd := &cobra.Command{
		Use:   "usage <modelId>",
		Short: "Earnings and inference stats for a model (usageStats)",
		Example: "  unitypredict-pp-cli models usage c6985b70-15e0-4ba9-8cb1-128200595a8f\n  unitypredict-pp-cli models usage c6985b70-15e0-4ba9-8cb1-128200595a8f --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
			"pp:happy-args":  "modelId=c6985b70-15e0-4ba9-8cb1-128200595a8f",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "models usage")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("modelId is required\nUsage: %s <modelId>", cmd.CommandPath()))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			env := resolveUptEnv(envFlag)
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			data, err := c.Get(ctx, "/api/models/"+args[0]+"/usageStats", nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			view := struct {
				ModelID string          `json:"model_id"`
				Env     string          `json:"env"`
				Stats   json.RawMessage `json:"stats"`
			}{args[0], env, data}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "usage stats for %s (%s):\n", args[0], env)
			var pretty any
			if json.Unmarshal(data, &pretty) == nil {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(pretty)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if modelsCmd, _, err := root.Find([]string{"models"}); err == nil {
			addNovelCommandIfAbsent(modelsCmd, newUptModelsUsageCmd(flags))
		}
	})
}
