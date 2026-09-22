// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T1.
// pp:data-source live

package cli

// logs get — one command, every log granularity for a request:
//   devlog  : per-request DevLog.txt (presigned URL from /api/models/requests)
//   runtime : per-request Log.txt   (presigned URL from /api/predict/status/{id}/logs)
//   build   : engine lastbuild.log  (presigned URL from /api/engines/buildlogs/{engineId},
//             the same log the first-party CLI writes as DebugLogs.txt)
// All three presigned hops are fetched with NO Authorization header — the S3
// signature lives in the query string and a riding credential both breaks the
// request (400) and gets echoed back in the error body.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const uptMaxLogBytes = 1 << 20 // 1 MiB print cap per log

type uptLogPart struct {
	Kind    string `json:"kind"`
	URL     string `json:"url"`
	Status  int    `json:"http_status"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

func newUptLogsGetCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	var only string
	var all bool
	var maxScanPages int
	cmd := &cobra.Command{
		Use:   "get <requestId>",
		Short: "Fetch a request's devLog, runtime Log.txt, and engine build log",
		Long: "Fans out to the three log granularities for one predict request:\n" +
			"  devlog   per-request DevLog.txt  (/api/models/requests -> devLogUrl)\n" +
			"  runtime  per-request Log.txt     (/api/predict/status/{id}/logs)\n" +
			"  build    engine lastbuild.log    (/api/engines/buildlogs/{engineId}; the first-party\n" +
			"           CLI surfaces this as DebugLogs.txt)\n" +
			"Presigned S3 hops carry no Authorization header (a riding credential 400s and is echoed back).",
		Example: "  unitypredict-pp-cli logs get b5a43cbf-dac1-45a9-a27b-6bbb0e4098f8 --all\n  unitypredict-pp-cli logs get b5a43cbf-dac1-45a9-a27b-6bbb0e4098f8 --only devlog,runtime --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
			"pp:happy-args":  "requestId=b5a43cbf-dac1-45a9-a27b-6bbb0e4098f8;--all",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "logs get")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("requestId is required\nUsage: %s <requestId>", cmd.CommandPath()))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			requestID := args[0]
			env := resolveUptEnv(envFlag)
			want := map[string]bool{"devlog": true, "runtime": true, "build": true}
			if !all && strings.TrimSpace(only) != "" {
				want = map[string]bool{}
				for _, k := range strings.Split(only, ",") {
					k = strings.ToLower(strings.TrimSpace(k))
					switch k {
					case "devlog", "runtime", "build":
						want[k] = true
					default:
						return usageErr(fmt.Errorf("--only: unknown log kind %q (want devlog, runtime, build)", k))
					}
				}
			}
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			// The status record gives us modelId (and request state) without
			// any scan.
			statusData, err := c.GetNoCache(ctx, "/api/predict/status/"+requestID, nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var statusRec struct {
				ModelID string `json:"modelId"`
				Status  string `json:"status"`
			}
			_ = json.Unmarshal(statusData, &statusRec)

			fetchPart := func(kind string, presignJSON func() ([]byte, error)) uptLogPart {
				part := uptLogPart{Kind: kind}
				data, err := presignJSON()
				if err != nil {
					part.Error = fmt.Sprintf("presign step: %v", err)
					return part
				}
				part.URL = uptExtractURL(data)
				if part.URL == "" {
					part.Error = "no presigned URL in response"
					return part
				}
				content, code, err := uptFetchPresigned(ctx, "GET", part.URL, nil)
				part.Status = code
				if err != nil {
					part.Error = err.Error()
					return part
				}
				if len(content) > uptMaxLogBytes {
					content = append(content[:uptMaxLogBytes], []byte("\n...[truncated]")...)
				}
				part.Content = string(content)
				return part
			}

			parts := make([]uptLogPart, 0, 3)
			if want["runtime"] {
				parts = append(parts, fetchPart("runtime", func() ([]byte, error) {
					return c.Get(ctx, "/api/predict/status/"+requestID+"/logs", nil)
				}))
			}
			if want["devlog"] && statusRec.ModelID != "" {
				parts = append(parts, fetchPart("devlog", func() ([]byte, error) {
					// Scan-and-filter: match this requestId inside the model's
					// request history. Bounded by --max-scan-pages.
					pages := maxScanPages
					for page := 1; page <= pages; page++ {
						data, err := c.Get(ctx, "/api/models/requests/"+statusRec.ModelID, map[string]string{"pageNumber": fmt.Sprint(page)})
						if err != nil {
							return nil, err
						}
						var pageRec struct {
							Results []struct {
								RequestID string `json:"requestId"`
								DevLogURL string `json:"devLogUrl"`
							} `json:"results"`
						}
						if err := json.Unmarshal(data, &pageRec); err != nil {
							return nil, err
						}
						for _, r := range pageRec.Results {
							if r.RequestID == requestID && r.DevLogURL != "" {
								return []byte(`{"url":"` + r.DevLogURL + `"}`), nil
							}
						}
						if len(pageRec.Results) == 0 {
							break
						}
					}
					return nil, fmt.Errorf("request %s not found in model %s history within %d page(s)", requestID, statusRec.ModelID, pages)
				}))
			}
			if want["build"] && statusRec.ModelID != "" {
				parts = append(parts, fetchPart("build", func() ([]byte, error) {
					m, err := c.Get(ctx, "/api/models/"+statusRec.ModelID, nil)
					if err != nil {
						return nil, err
					}
					var modelRec struct {
						AppEngineID string `json:"appEngineId"`
					}
					if err := json.Unmarshal(m, &modelRec); err != nil {
						return nil, err
					}
					if modelRec.AppEngineID == "" {
						return nil, fmt.Errorf("model %s has no engine", statusRec.ModelID)
					}
					return c.Get(ctx, "/api/engines/buildlogs/"+stripEnginePrefix(modelRec.AppEngineID), nil)
				}))
			}

			failures := make([]uptLogPart, 0)
			ok := 0
			for _, p := range parts {
				if p.Error != "" {
					failures = append(failures, p)
				} else {
					ok++
				}
			}
			if len(failures) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %d of %d log fetches failed\n", len(failures), len(parts))
			}
			view := struct {
				RequestID string       `json:"request_id"`
				ModelID   string       `json:"model_id,omitempty"`
				Env       string       `json:"env"`
				Status    string       `json:"status,omitempty"`
				Logs      []uptLogPart `json:"logs"`
				ScannedPages int       `json:"scanned_request_pages"`
				FetchFailures []uptLogPart `json:"fetch_failures,omitempty"`
			}{RequestID: requestID, ModelID: statusRec.ModelID, Env: env, Status: statusRec.Status, Logs: parts, ScannedPages: maxScanPages}
			view.FetchFailures = failures
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			for _, p := range parts {
				fmt.Fprintf(cmd.OutOrStdout(), "===== %s =====\n", strings.ToUpper(p.Kind))
				if p.Error != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "[unavailable: %s]\n\n", p.Error)
					continue
				}
				fmt.Fprintln(cmd.OutOrStdout(), p.Content)
				fmt.Fprintln(cmd.OutOrStdout())
			}
			if len(parts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no log kinds selected")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().StringVar(&only, "only", "", "Comma list: devlog,runtime,build (default all)")
	cmd.Flags().BoolVar(&all, "all", true, "Fetch all three log kinds (default)")
	cmd.Flags().IntVar(&maxScanPages, "max-scan-pages", 3, "Request-history pages to scan when locating the devLog URL")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if logsParent, _, err := root.Find([]string{"logs"}); err == nil && logsParent.Name() == "logs" && !logsParent.Runnable() {
			addNovelCommandIfAbsent(logsParent, newUptLogsGetCmd(flags))
			return
		}
		logsParent := &cobra.Command{Use: "logs", Short: "Fetch UnityPredict logs at every granularity"}
		logsParent.AddCommand(newUptLogsGetCmd(flags))
		addNovelCommandIfAbsent(root, logsParent)
	})
}
