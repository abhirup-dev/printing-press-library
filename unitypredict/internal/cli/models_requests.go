// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Route merge from
// endpoints.md (#20) — absent from the 25-endpoint sniffed spec.
// pp:data-source live

package cli

// models requests — per-model inference request history with each request's
// devLogUrl (presigned). Note: the engineId field in results carries the
// AppEngineDefinition- prefix; it is stripped in the rendered view.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUptModelsRequestsCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	var page int
	cmd := &cobra.Command{
		Use:   "requests <modelId>",
		Short: "Inference request history for a model",
		Long: "GET /api/models/requests/{modelId} — the per-model request history with status, timings,\n" +
			"cost, and each request's presigned devLogUrl.",
		Example: "  unitypredict-pp-cli models requests c6985b70-15e0-4ba9-8cb1-128200595a8f\n  unitypredict-pp-cli models requests c6985b70-15e0-4ba9-8cb1-128200595a8f --page 2 --json",
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
				return writeDryRun(cmd.OutOrStdout(), flags, "models requests")
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
			data, err := c.Get(ctx, "/api/models/requests/"+args[0], map[string]string{"pageNumber": fmt.Sprint(page)})
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var raw struct {
				Results []struct {
					RequestID    string  `json:"requestId"`
					Status       string  `json:"status"`
					CreatedTime  string  `json:"createdTime"`
					ComputeTime  string  `json:"computeTime"`
					TimeInQueue  string  `json:"timeInQueue"`
					ComputeCost  float64 `json:"computeCost"`
					EngineID     string  `json:"engineId"`
					HasDevLog    bool    `json:"-"`
				} `json:"results"`
				PageNumber int `json:"pageNumber"`
				TotalCount int `json:"totalCount"`
			}
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing requests response: %w", err)
			}
			type row struct {
				RequestID   string  `json:"requestId"`
				Status      string  `json:"status"`
				CreatedTime string  `json:"createdTime"`
				ComputeTime string  `json:"computeTime"`
				TimeInQueue string  `json:"timeInQueue"`
				ComputeCost float64 `json:"computeCost"`
				EngineID    string  `json:"engineId"`
			}
			rows := make([]row, 0, len(raw.Results))
			for _, r := range raw.Results {
				rows = append(rows, row{r.RequestID, r.Status, r.CreatedTime, r.ComputeTime, r.TimeInQueue, r.ComputeCost, stripEnginePrefix(r.EngineID)})
			}
			view := struct {
				ModelID    string `json:"model_id"`
				Env        string `json:"env"`
				Page       int    `json:"page"`
				TotalCount int    `json:"total_count"`
				Requests   []row  `json:"requests"`
			}{args[0], env, raw.PageNumber, raw.TotalCount, rows}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No requests recorded for this model.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d request(s) (page %d, total %d):\n\n", len(rows), view.Page, view.TotalCount)
			table := make([][]string, 0, len(rows))
			for _, r := range rows {
				table = append(table, []string{r.RequestID, r.Status, r.CreatedTime, r.ComputeTime, r.TimeInQueue, fmt.Sprintf("%.6f", r.ComputeCost)})
			}
			return flags.printTable(cmd, []string{"requestId", "status", "created", "compute", "queued", "cost"}, table)
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number to fetch")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if modelsCmd, _, err := root.Find([]string{"models"}); err == nil {
			addNovelCommandIfAbsent(modelsCmd, newUptModelsRequestsCmd(flags))
		}
	})
}
