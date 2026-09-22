// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T4.
// pp:data-source live

package cli

// engines models — reverse lookup: which models are built on this engine.
// The engineId= query param on /api/models/usermodels and /api/models/search is
// silently ignored (returns the unfiltered list), so the dedicated
// GET /api/engines/{engineId}/models route is the only correct answer.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUptEnginesModelsCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	cmd := &cobra.Command{
		Use:   "models <engineId>",
		Short: "List every model built on an engine (reverse lookup)",
		Long: "Reverse lookup: given an engine, which models use it. The engineId= filter on list\n" +
			"routes is silently ignored, so before redeploying or deleting an engine this is the only\n" +
			"way to see its blast radius. Accepts the bare uuid or an AppEngineDefinition-prefixed id.",
		Example: "  unitypredict-pp-cli engines models b5132be0-a373-4f74-bedf-fabb17dc0342\n  unitypredict-pp-cli engines models b5132be0-a373-4f74-bedf-fabb17dc0342 --env prod --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
			"pp:happy-args":  "engineId=b5132be0-a373-4f74-bedf-fabb17dc0342",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "engines models")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("engineId is required\nUsage: %s <engineId>", cmd.CommandPath()))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			engineID := stripEnginePrefix(args[0])
			env := resolveUptEnv(envFlag)
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			data, err := c.Get(ctx, "/api/engines/"+engineID+"/models", nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var items []map[string]any
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("parsing models-for-engine response: %w", err)
			}
			view := struct {
				EngineID string           `json:"engine_id"`
				Env      string           `json:"env"`
				Count    int              `json:"count"`
				Models   []map[string]any `json:"models"`
			}{EngineID: engineID, Env: env, Count: len(items), Models: items}
			if items == nil {
				view.Models = make([]map[string]any, 0)
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			rows := make([][2]string, 0, len(items))
			for _, m := range items {
				rows = append(rows, [2]string{fmt.Sprint(m["modelId"]), fmt.Sprint(m["modelName"])})
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No models use this engine.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d model(s) use engine %s:\n\n", len(rows), engineID)
			return flags.printTable(cmd, []string{"modelId", "modelName"}, flattenRows(rows))
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	return cmd
}

func flattenRows(rows [][2]string) [][]string {
	out := make([][]string, len(rows))
	for i, r := range rows {
		out[i] = []string{r[0], r[1]}
	}
	return out
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if enginesCmd, _, err := root.Find([]string{"engines"}); err == nil {
			addNovelCommandIfAbsent(enginesCmd, newUptEnginesModelsCmd(flags))
		}
	})
}
