// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T2.
// pp:data-source live

package cli

// models describe — the long description is a hosted FILE, not a text field.
// Writing prose into modelLongDescription is silently ignored; the console does
// presign -> S3 PUT -> full-record upsert. These commands collapse that.
//   get: fetch the rendered MODELDESCRIPTION.html content
//   set: upload a Markdown file and persist the reference (mutation; prod-guarded)

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newUptModelsDescribeCmd(flags *rootFlags) *cobra.Command {
	var envFlag, getEnv string
	var filePath string
	var iKnow bool

	getCmd := &cobra.Command{
		Use:   "get <modelId>",
		Short: "Fetch a model's long description (authenticated filekey route -> presigned S3)",
		Example: "  unitypredict-pp-cli models describe get c6985b70-15e0-4ba9-8cb1-128200595a8f",
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
				return writeDryRun(cmd.OutOrStdout(), flags, "models describe get")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("modelId is required\nUsage: %s <modelId>", cmd.CommandPath()))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			env := resolveUptEnv(getEnv)
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			// Primary (and only) path: the authenticated filekey route. It 302s
			// to a presigned S3 URL for private AND public models; the
			// /api/public/... mirror 401s for private models even when
			// authenticated, so it is not used at all.
			data, code, ferr := uptGetAuthedToPresigned(ctx, flags, c, "/api/models/"+args[0]+"/filekey/MODELDESCRIPTION")
			content := ""
			fetchErr := ""
			note := ""
			switch {
			case ferr != nil:
				fetchErr = ferr.Error()
			case code == 404:
				note = "no long description hosted for this model"
			case code != 200:
				fetchErr = fmt.Sprintf("HTTP %d fetching the hosted file", code)
			default:
				content = string(data)
			}
			view := struct {
				ModelID      string `json:"model_id"`
				Env          string `json:"env"`
				Source       string `json:"source"`
				Content      string `json:"content,omitempty"`
				ContentBytes int    `json:"content_bytes"`
				FetchError   string `json:"fetch_error,omitempty"`
				Note         string `json:"note,omitempty"`
			}{args[0], env, "filekey", content, len(content), fetchErr, note}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			if content != "" {
				fmt.Fprint(cmd.OutOrStdout(), content)
				return nil
			}
			if fetchErr != "" {
				return fmt.Errorf("%s", fetchErr)
			}
			fmt.Fprintln(cmd.OutOrStdout(), note)
			return nil
		},
	}
	getCmd.Flags().StringVar(&getEnv, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")

	setCmd := &cobra.Command{
		Use:   "set <modelId>",
		Short: "Upload a Markdown file as the model's long description",
		Long: "The long description is a hosted file, not a text field: presign\n" +
			"(/api/models/upload/{id}/MODELDESCRIPTION?fileName=longdescription.md), PUT the bytes to S3 with\n" +
			"no Authorization header, then POST the full model record back. Writing prose directly into\n" +
			"modelLongDescription is silently ignored.\n" +
			"Markdown limits: no fenced code blocks (the console parser rejects ``` fences); use lists and\n" +
			"inline `code`.",
		Example: "  unitypredict-pp-cli models describe set c6985b70-15e0-4ba9-8cb1-128200595a8f --file README.md\n  unitypredict-pp-cli models describe set c6985b70-15e0-4ba9-8cb1-128200595a8f --file README.md --dry-run",
		Annotations: map[string]string{
			"pp:data-source": "live",
			"pp:happy-args":  "modelId=c6985b70-15e0-4ba9-8cb1-128200595a8f;--file=README.md;--dry-run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "models describe set")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("modelId is required\nUsage: %s <modelId>", cmd.CommandPath()))
			}
			if filePath == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--file is required (Markdown file to upload)"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			modelID := args[0]
			env := resolveUptEnv(envFlag)
			if err := guardUptProdMutation(env, "set a model long description", iKnow); err != nil {
				return err
			}
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			body, err := os.ReadFile(filePath) /* #nosec G304 -- user-supplied --file path is the feature; local read only */
			if err != nil {
				return fmt.Errorf("reading --file %s: %w", filePath, err)
			}
			// 1. presign
			presign, err := c.Get(ctx, "/api/models/upload/"+modelID+"/MODELDESCRIPTION", map[string]string{"fileName": "longdescription.md"})
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			uploadURL := uptExtractURL(presign)
			if uploadURL == "" {
				return fmt.Errorf("no presigned URL in upload response; raw: %.200s", presign)
			}
			// 2. S3 PUT — no Authorization header, ever.
			if _, status, err := uptFetchPresigned(ctx, "PUT", uploadURL, body); err != nil || status >= 300 {
				return fmt.Errorf("S3 PUT failed (HTTP %d): %w", status, err)
			}
			// 3. full-record upsert (GET, POST unchanged record back; the file
			// reference is managed server-side).
			record, err := c.GetNoCache(ctx, "/api/models/"+modelID, nil)
			if err != nil {
				return fmt.Errorf("re-reading model record after upload: %w", err)
			}
			if _, _, err := c.Post(ctx, "/api/models", json.RawMessage(record)); err != nil {
				return fmt.Errorf("persisting model record: %w", err)
			}
			view := struct {
				ModelID   string `json:"model_id"`
				Env       string `json:"env"`
				File      string `json:"file"`
				Bytes     int    `json:"bytes"`
				Uploaded  bool   `json:"uploaded"`
			}{modelID, env, filePath, len(body), true}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "long description for %s set from %s (%d bytes)\n", modelID, filePath, len(body))
			return nil
		},
	}
	setCmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	setCmd.Flags().StringVar(&filePath, "file", "", "Markdown file to upload (required)")
	setCmd.Flags().BoolVar(&iKnow, "i-know", false, "Allow the mutation against prod (off by default: prod is read-only)")

	parent := &cobra.Command{
		Use:   "describe",
		Short: "Model long description (hosted file: presign + S3 PUT + upsert)",
		Example: "  unitypredict-pp-cli models describe get <modelId>\n  unitypredict-pp-cli models describe set <modelId> --file README.md",
	}
	parent.AddCommand(getCmd)
	parent.AddCommand(setCmd)
	return parent
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if modelsCmd, _, err := root.Find([]string{"models"}); err == nil {
			addNovelCommandIfAbsent(modelsCmd, newUptModelsDescribeCmd(flags))
		}
	})
}
