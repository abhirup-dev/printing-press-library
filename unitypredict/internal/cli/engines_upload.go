// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Route merge from
// pp:data-source live
// endpoints.md (#25) — absent from the 25-endpoint sniffed spec.

package cli

// engines upload — presign a source-file upload for an engine and PUT the bytes.
// GET /api/engines/upload/{engineId}/{fileType} returns a presigned S3 PUT; the
// PUT must carry NO Authorization header or S3 400s and echoes the credential
// back in the error body. Mutating: refuses on prod without --i-know.

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newUptEnginesUploadCmd(flags *rootFlags) *cobra.Command {
	var envFlag string
	var filePath string
	var fileType string
	var iKnow bool
	cmd := &cobra.Command{
		Use:   "upload <engineId>",
		Short: "Upload a source file to an engine (presign + S3 PUT)",
		Long: "GET /api/engines/upload/{engineId}/{fileType} returns a presigned S3 PUT URL, then the file\n" +
			"bytes are PUT with no Authorization header (S3 rejects a riding credential). fileType must be\n" +
			"SourceFile for engine sources (EntryPoint.py, requirements.txt, AdditionalDockerCommands.txt).",
		Example: "  unitypredict-pp-cli engines upload b5132be0-a373-4f74-bedf-fabb17dc0342 --file EntryPoint.py\n  unitypredict-pp-cli engines upload b5132be0-a373-4f74-bedf-fabb17dc0342 --file requirements.txt --dry-run",
		Annotations: map[string]string{
			"pp:data-source": "live",
			"pp:happy-args":  "engineId=b5132be0-a373-4f74-bedf-fabb17dc0342;--file=EntryPoint.py;--dry-run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "engines upload")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("engineId is required\nUsage: %s <engineId>", cmd.CommandPath()))
			}
			if filePath == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--file is required (path to the source file to upload)"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			engineID := stripEnginePrefix(args[0])
			env := resolveUptEnv(envFlag)
			if err := guardUptProdMutation(env, "upload an engine source file", iKnow); err != nil {
				return err
			}
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			// Filesystem read stays after dryRunOK per the RunE contract.
			body, err := os.ReadFile(filePath) /* #nosec G304 -- user-supplied --file path is the feature; local read only */
			if err != nil {
				return fmt.Errorf("reading --file %s: %w", filePath, err)
			}
			data, err := c.Get(ctx, "/api/engines/upload/"+engineID+"/"+fileType, nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			uploadURL := uptExtractURL(data)
			if uploadURL == "" {
				return fmt.Errorf("no presigned URL in response for %s upload; raw: %.200s", fileType, data)
			}
			if _, status, err := uptFetchPresigned(ctx, "PUT", uploadURL, body); err != nil || status >= 300 {
				return fmt.Errorf("S3 PUT failed (HTTP %d): %w", status, err)
			}
			view := struct {
				EngineID  string `json:"engine_id"`
				Env       string `json:"env"`
				FileType  string `json:"file_type"`
				File      string `json:"file"`
				Bytes     int    `json:"bytes"`
				Uploaded  bool   `json:"uploaded"`
			}{engineID, env, fileType, filePath, len(body), true}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "uploaded %s (%d bytes) to engine %s as %s\n", filePath, len(body), engineID, fileType)
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to the file to upload (required)")
	cmd.Flags().StringVar(&fileType, "file-type", "SourceFile", "Upload file type; the API accepts SourceFile")
	cmd.Flags().BoolVar(&iKnow, "i-know", false, "Allow the mutation against prod (off by default: prod is read-only)")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if enginesCmd, _, err := root.Find([]string{"engines"}); err == nil {
			addNovelCommandIfAbsent(enginesCmd, newUptEnginesUploadCmd(flags))
		}
	})
}
