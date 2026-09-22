// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T6.
// pp:data-source live

package cli

// engines concurrency renders stateInfo from GET /api/engines/{id}: how many
// engine nodes ran in parallel over time (historicalNodeCountStats), the
// current instance count, and the running-nodes table. This is the data behind
// the console's Engine Stats page.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

type uptNodeCountStats struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
	Sum   float64 `json:"sum"`
	Count int     `json:"count"`
}

type uptRunningNode struct {
	AgentInstanceID  string  `json:"agentInstanceId"`
	EngineLoadedTime string  `json:"engineLoadedTime"`
	GPUMemoryMB      float64 `json:"gpuMemory"`
	MemoryMB         float64 `json:"memory"`
	StorageMB        float64 `json:"storage"`
	BuildNumber      int     `json:"buildNumber"`
}

type uptStateInfo struct {
	NumRunningInstances    int                            `json:"numRunningInstances"`
	RunningNodes           []uptRunningNode               `json:"runningNodes"`
	HistoricalNodeCountStats map[string]uptNodeCountStats `json:"historicalNodeCountStats"`
}

// uptDayKeyLabel renders the stats map's day key. Observed keys carry a fixed
// prefix before a DDMMYYYY tail (e.g. "2118092026" -> 18 Sep 2026); when the
// tail does not parse, the raw key is shown rather than a fabricated date.
func uptDayKeyLabel(key string) string {
	if len(key) < 8 {
		return key
	}
	tail := key[len(key)-8:]
	dd, err1 := strconv.Atoi(tail[0:2])
	mm, err2 := strconv.Atoi(tail[2:4])
	yy, err3 := strconv.Atoi(tail[4:8])
	if err1 != nil || err2 != nil || err3 != nil || dd < 1 || dd > 31 || mm < 1 || mm > 12 {
		return key
	}
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	return fmt.Sprintf("%02d %s %d", dd, months[mm-1], yy)
}

func newUptEnginesConcurrencyCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	cmd := &cobra.Command{
		Use:   "concurrency <engineId>",
		Short: "Show engine node parallelism over time (stateInfo)",
		Long: "Renders the engine record's stateInfo: how many engine nodes ran in parallel per day\n" +
			"(historicalNodeCountStats min/avg/max), the current instance count, and each running node's\n" +
			"GPU memory, RAM, and storage. Same data as the console's Engine Stats page.",
		Example: "  unitypredict-pp-cli engines concurrency b5132be0-a373-4f74-bedf-fabb17dc0342\n  unitypredict-pp-cli engines concurrency b5132be0-a373-4f74-bedf-fabb17dc0342 --json",
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
				return writeDryRun(cmd.OutOrStdout(), flags, "engines concurrency")
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
			data, err := c.Get(ctx, "/api/engines/"+engineID, nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var record struct {
				StateInfo uptStateInfo `json:"stateInfo"`
				Status    string       `json:"status"`
			}
			if err := json.Unmarshal(data, &record); err != nil {
				return fmt.Errorf("parsing engine record: %w", err)
			}
			if record.StateInfo.RunningNodes == nil {
				record.StateInfo.RunningNodes = make([]uptRunningNode, 0)
			}
			days := make([]struct {
				Day  string           `json:"day"`
				Min  float64          `json:"min"`
				Avg  float64          `json:"avg"`
				Max  float64          `json:"max"`
				Samples int           `json:"samples"`
			}, 0, len(record.StateInfo.HistoricalNodeCountStats))
			for k, v := range record.StateInfo.HistoricalNodeCountStats {
				days = append(days, struct {
					Day  string  `json:"day"`
					Min  float64 `json:"min"`
					Avg  float64 `json:"avg"`
					Max  float64 `json:"max"`
					Samples int `json:"samples"`
				}{Day: uptDayKeyLabel(k), Min: v.Min, Avg: v.Mean, Max: v.Max, Samples: v.Count})
			}
			sort.Slice(days, func(i, j int) bool { return days[i].Day > days[j].Day })
			view := struct {
				EngineID            string            `json:"engine_id"`
				Env                 string            `json:"env"`
				EngineStatus        string            `json:"engine_status"`
				NumRunningInstances int               `json:"num_running_instances"`
				RunningNodes        []uptRunningNode  `json:"running_nodes"`
				HistoricalDayStats  []struct {
					Day     string  `json:"day"`
					Min     float64 `json:"min"`
					Avg     float64 `json:"avg"`
					Max     float64 `json:"max"`
					Samples int     `json:"samples"`
				} `json:"historical_day_stats"`
			}{
				EngineID:            engineID,
				Env:                 env,
				EngineStatus:        record.Status,
				NumRunningInstances: record.StateInfo.NumRunningInstances,
				RunningNodes:        record.StateInfo.RunningNodes,
				HistoricalDayStats:  days,
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "engine %s (%s, status %s): running on %d node(s)\n\n",
				engineID, view.Env, view.EngineStatus, view.NumRunningInstances)
			if len(days) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Node count by day (min / avg / max / samples):")
				for _, d := range days {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-14s %.0f / %.1f / %.0f / %d\n", d.Day, d.Min, d.Avg, d.Max, d.Samples)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			if len(view.RunningNodes) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Running nodes:")
				rows := make([][]string, 0, len(view.RunningNodes))
				for _, n := range view.RunningNodes {
					rows = append(rows, []string{
						n.AgentInstanceID,
						n.EngineLoadedTime,
						fmt.Sprintf("%.0f MB", n.GPUMemoryMB),
						fmt.Sprintf("%.0f MB", n.MemoryMB),
						fmt.Sprintf("%.0f MB", n.StorageMB),
					})
				}
				return flags.printTable(cmd, []string{"nodeId", "loadedAt", "gpuMem", "ram", "storage"}, rows)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if enginesCmd, _, err := root.Find([]string{"engines"}); err == nil {
			addNovelCommandIfAbsent(enginesCmd, newUptEnginesConcurrencyCmd(flags))
		}
	})
}
