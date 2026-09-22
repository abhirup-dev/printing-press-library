// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T7.
// pp:data-source live

package cli

// predict run — invoke, optionally wait to completion, optionally download
// File outcomes. Text inputs via --input k=v (repeatable); File inputs via
// --input-file k=path (repeatable), which runs the file session chain:
// POST /api/predict/initialize -> GET /api/predict/upload/{rid}/{name}
// (presigned) -> PUT bytes (no auth) -> POST /api/predict/{model}/{rid}.
// File outcomes download via /api/predict/download/{rid}/{name}, which 302s to
// presigned S3; the redirect hop drops Authorization (credential-echo vector).

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"unitypredict-pp-cli/internal/client"
)

func newUptPredictRunCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	var inputs []string
	var inputFiles []string
	var desired []string
	var wait bool
	var downloadDir string
	var pollInterval time.Duration
	var waitTimeout time.Duration
	cmd := &cobra.Command{
		Use:   "run <modelId>",
		Short: "Invoke a model (text or file inputs), wait, and download File outcomes",
		Long: "Invokes a MODEL id (not the engine id — predict 404s on engine ids).\n" +
			"Text inputs: --input Key=Value (repeatable). File inputs: --input-file Key=path (repeatable),\n" +
			"which runs the full file session chain (initialize -> presign -> S3 PUT -> predict).\n" +
			"--wait polls until Completed/Error (GPU cold starts can take minutes; default budget 15m).\n" +
			"--download DIR writes every File outcome into DIR; String outcomes print to stdout.",
		Example: "  unitypredict-pp-cli predict run <modelId> --input TargetLanguage=Hindi --wait\n" +
			"  unitypredict-pp-cli predict run <modelId> --input-file SubtitleFile=sub.srt --wait --download ./out",
		Annotations: map[string]string{
			"pp:data-source": "live",
			"pp:happy-args":  "modelId=c6985b70-15e0-4ba9-8cb1-128200595a8f;--input=InputMessage=hello;--dry-run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "predict run")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("modelId is required (the MODEL id, not the engine id)\nUsage: %s <modelId>", cmd.CommandPath()))
			}
			if len(inputs) == 0 && len(inputFiles) == 0 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("at least one --input Key=Value or --input-file Key=path is required"))
			}
			// Filesystem reads (input files) come after dryRunOK per contract.
			for _, in := range inputs {
				if !strings.Contains(in, "=") {
					return usageErr(fmt.Errorf("--input %q must be Key=Value", in))
				}
			}
			fileBytes := map[string][]byte{}
			for _, in := range inputFiles {
				kv := strings.SplitN(in, "=", 2)
				if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
					return usageErr(fmt.Errorf("--input-file %q must be Key=path", in))
				}
				b, err := os.ReadFile(kv[1])
				if err != nil {
					return fmt.Errorf("reading --input-file %s: %w", kv[1], err)
				}
				fileBytes[kv[0]] = b
			}
			ctx, cancel := contextWithWaitBudget(cmd.Context(), flags, waitTimeout)
			defer cancel()
			modelID := args[0]
			env := resolveUptEnv(envFlag)
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}

			inputValues := map[string]any{}
			for _, in := range inputs {
				kv := strings.SplitN(in, "=", 2)
				inputValues[kv[0]] = kv[1]
			}
			desiredOutcomes := make([]string, 0, len(desired))
			for _, d := range desired {
				desiredOutcomes = append(desiredOutcomes, strings.TrimSpace(d))
			}

			var requestID, contextID string
			if len(fileBytes) > 0 {
				// File session: initialize -> presign per file -> PUT -> session POST.
				initData, _, err := c.Post(ctx, "/api/predict/initialize/"+modelID, nil)
				if err != nil {
					return classifyAPIError(cmd.OutOrStdout(), err, flags)
				}
				var initRec struct {
					RequestID string `json:"requestId"`
					ContextID string `json:"contextId"`
				}
				if err := json.Unmarshal(initData, &initRec); err != nil || initRec.RequestID == "" {
					return fmt.Errorf("predict initialize response missing requestId/contextId: %.200s", initData)
				}
				requestID, contextID = initRec.RequestID, initRec.ContextID
				for key, b := range fileBytes {
					// The presign GET quotes the local filename; the canonical
					// name comes back in fileName and is what InputValues needs.
					base := filepath.Base(localNameForKey(inputFiles, key))
					presign, err := c.Get(ctx, "/api/predict/upload/"+requestID+"/"+url.PathEscape(base), nil)
					if err != nil {
						return fmt.Errorf("presigning upload for %s: %w", key, err)
					}
					uploadURL := uptExtractURL(presign)
					if uploadURL == "" {
						return fmt.Errorf("no presigned URL for %s; raw: %.200s", key, presign)
					}
					var canonical struct {
						FileName string `json:"fileName"`
					}
					_ = json.Unmarshal(presign, &canonical)
					if _, status, err := uptFetchPresigned(ctx, "PUT", uploadURL, b); err != nil || status >= 300 {
						return fmt.Errorf("S3 PUT for %s failed (HTTP %d): %w", key, status, err)
					}
					inputValues[key] = canonical.FileName
				}
				body := map[string]any{
					"InputValues":     inputValues,
					"DesiredOutcomes": desiredOutcomes,
					"contextId":       contextID,
					"RequestId":       requestID,
					"interfaceType":   "Form",
				}
				startData, _, err := c.Post(ctx, "/api/predict/"+modelID+"/"+requestID, body)
				if err != nil {
					return classifyAPIError(cmd.OutOrStdout(), err, flags)
				}
				var startRec struct {
					RequestID string `json:"requestId"`
				}
				_ = json.Unmarshal(startData, &startRec)
				if startRec.RequestID != "" {
					requestID = startRec.RequestID
				}
			} else {
				body := map[string]any{
					"InputValues":      inputValues,
					"DesiredOutcomes":  desiredOutcomes,
					"outputFolderPath": nil,
				}
				startData, _, err := c.Post(ctx, "/api/predict/"+modelID, body)
				if err != nil {
					return classifyAPIError(cmd.OutOrStdout(), err, flags)
				}
				var startRec struct {
					RequestID string `json:"requestId"`
					Status    string `json:"status"`
				}
				if err := json.Unmarshal(startData, &startRec); err != nil || startRec.RequestID == "" {
					return fmt.Errorf("predict response missing requestId: %.200s", startData)
				}
				requestID = startRec.RequestID
			}

			// Poll.
			statusRec := uptPollStatus{RequestID: requestID}
			if wait {
				ticker := time.NewTicker(pollInterval)
				defer ticker.Stop()
				for {
					rec, err := uptFetchStatus(ctx, c, requestID)
					if err != nil {
						return fmt.Errorf("polling status: %w", err)
					}
					statusRec = rec
					if rec.Status != "Processing" {
						break
					}
					select {
					case <-ctx.Done():
						return fmt.Errorf("wait budget exhausted while status=%s (request %s still running server-side)", rec.Status, requestID)
					case <-ticker.C:
					}
				}
			} else {
				rec, err := uptFetchStatus(ctx, c, requestID)
				if err == nil {
					statusRec = rec
				}
			}

			downloads := make([]map[string]any, 0)
			if downloadDir != "" && statusRec.Status == "Completed" {
				if err := os.MkdirAll(downloadDir, 0o755); err != nil { /* #nosec G301 -- user-openable download dir; no secrets land there */
					return fmt.Errorf("creating --download dir: %w", err)
				}
				authHdr := c.Config.AuthHeader() // covers creds-file and env-var sources
				for name, values := range statusRec.Outcomes {
					for _, ov := range values {
						if ov.DataType != "File" || ov.Value == "" {
							continue
						}
						dst := filepath.Join(downloadDir, fmt.Sprint(ov.Value))
						n, err := uptDownloadAuthedThenPresigned(ctx, flags, c.RequestBaseURL(), authHdr, "/api/predict/download/"+requestID+"/"+url.PathEscape(fmt.Sprint(ov.Value)), dst)
						entry := map[string]any{"outcome": name, "file": fmt.Sprint(ov.Value), "path": dst, "bytes": n}
						if err != nil {
							entry["error"] = err.Error()
						}
						downloads = append(downloads, entry)
					}
				}
			}

			view := struct {
				ModelID   string                   `json:"model_id"`
				Env       string                   `json:"env"`
				RequestID string                   `json:"request_id"`
				Status    string                   `json:"status"`
				Waited    bool                     `json:"waited"`
				Outcomes  map[string][]uptOutcome  `json:"outcomes,omitempty"`
				ErrorMessages string              `json:"error_messages,omitempty"`
				Downloads []map[string]any         `json:"downloads,omitempty"`
			}{modelID, env, requestID, statusRec.Status, wait, statusRec.Outcomes, statusRec.ErrorMessages, downloads}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "request %s: %s\n", requestID, statusRec.Status)
			if statusRec.ErrorMessages != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "engine errors: %s\n", statusRec.ErrorMessages)
			}
			for name, values := range statusRec.Outcomes {
				for _, ov := range values {
					fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", name, truncateForDisplay(fmt.Sprint(ov.Value), 300))
				}
			}
			for _, d := range downloads {
				if e, ok := d["error"]; ok {
					fmt.Fprintf(cmd.OutOrStderr(), "download failed for %v: %v\n", d["outcome"], e)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "saved %s -> %s (%v bytes)\n", d["file"], d["path"], d["bytes"])
				}
			}
			if statusRec.Status == "Error" {
				return apiErr(fmt.Errorf("request %s ended in Error", requestID))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().StringArrayVar(&inputs, "input", nil, "Text input Key=Value (repeatable)")
	cmd.Flags().StringArrayVar(&inputFiles, "input-file", nil, "File input Key=path (repeatable; runs the initialize/upload session chain)")
	cmd.Flags().StringArrayVar(&desired, "desired-outcomes", nil, "Outcome names to request (default: all)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Poll until Completed or Error (GPU cold starts can take minutes)")
	cmd.Flags().StringVar(&downloadDir, "download", "", "Directory to write File outcomes into")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", 3*time.Second, "Poll interval for --wait")
	cmd.Flags().DurationVar(&waitTimeout, "wait-timeout", uptDefaultWait, "Budget for --wait (overrides the 60s root default; an explicit --timeout still wins)")
	return cmd
}

type uptOutcome struct {
	Probability float64 `json:"probability"`
	Value       any     `json:"value"`
	DataType    string  `json:"dataType"`
}

type uptPollStatus struct {
	RequestID     string                  `json:"requestId"`
	Status        string                  `json:"status"`
	ErrorMessages string                  `json:"errorMessages"`
	Outcomes      map[string][]uptOutcome `json:"outcomes"`
}

func uptFetchStatus(ctx context.Context, c *client.Client, requestID string) (uptPollStatus, error) {
	data, err := c.GetNoCache(ctx, "/api/predict/status/"+requestID, nil)
	if err != nil {
		return uptPollStatus{}, err
	}
	var rec uptPollStatus
	if err := json.Unmarshal(data, &rec); err != nil {
		return uptPollStatus{}, fmt.Errorf("parsing status response: %w", err)
	}
	rec.RequestID = requestID
	return rec, nil
}

// contextWithWaitBudget applies the predict wait budget: an explicit root
// --timeout wins (the user asked for it); otherwise --wait-timeout (default
// 15m) replaces the 60s root default so minutes-long GPU cold starts survive.
func contextWithWaitBudget(parent context.Context, flags *rootFlags, waitTimeout time.Duration) (context.Context, context.CancelFunc) {
	if flags.timeoutExplicit {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, waitTimeout)
}

func truncateForDisplay(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func localNameForKey(inputFiles []string, key string) string {
	for _, in := range inputFiles {
		kv := strings.SplitN(in, "=", 2)
		if len(kv) == 2 && kv[0] == key {
			return kv[1]
		}
	}
	return key
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if predictCmd, _, err := root.Find([]string{"predict"}); err == nil {
			addNovelCommandIfAbsent(predictCmd, newUptPredictRunCmd(flags))
		}
	})
}
