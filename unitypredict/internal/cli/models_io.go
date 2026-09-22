// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T3.
// pp:data-source live

package cli

// models io set — declarative aiEngineConfig.inputs/outcomes from YAML.
// The upsert is a FULL-RECORD round-trip (partial bodies clobber the I/O
// contract), and each variable needs a matching variableDisplayOptions entry
// keyed <name_lower>_inputcon / <name_lower>_outputcon. This command owns the
// whole round-trip: GET record -> replace inputs/outcomes -> rebuild display
// options (preserving unrelated keys) -> POST back -> verify.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type uptIOVar struct {
	Name        string   `yaml:"name"          json:"name"`
	Type        string   `yaml:"type"          json:"-"`
	Description string   `yaml:"description"   json:"description"`
	Default     string   `yaml:"default"       json:"defaultValue"`
	Options     []string `yaml:"options"       json:"options"`
	Hidden      bool     `yaml:"hidden"        json:"-"`
	Multiline   bool     `yaml:"multiline"     json:"-"`
	Picture     bool     `yaml:"picture-viewer" json:"-"`
	Validation  bool     `yaml:"validation"    json:"-"`
	Chatbot     bool     `yaml:"chatbot"       json:"-"`
}

type uptIOSpec struct {
	Inputs   []uptIOVar `yaml:"inputs"`
	Outcomes []uptIOVar `yaml:"outcomes"`
}

// parseUptIOSpec parses and validates the declarative IO yaml.
func parseUptIOSpec(data []byte) (uptIOSpec, error) {
	var spec uptIOSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return spec, fmt.Errorf("parsing IO yaml: %w", err)
	}
	seen := map[string]bool{}
	validate := func(vars []uptIOVar, kind string) error {
		for i, v := range vars {
			if v.Name == "" {
				return fmt.Errorf("%s[%d]: name is required", kind, i)
			}
			if v.Type == "" {
				return fmt.Errorf("%s %q: type is required (e.g. String, File)", kind, v.Name)
			}
			key := strings.ToLower(v.Name) + "_" + kind
			if seen[key] {
				return fmt.Errorf("duplicate variable %q", v.Name)
			}
			seen[key] = true
			vars[i].Options = coalesceOptions(v.Options)
		}
		return nil
	}
	if err := validate(spec.Inputs, "inputcon"); err != nil {
		return spec, err
	}
	if err := validate(spec.Outcomes, "outputcon"); err != nil {
		return spec, err
	}
	return spec, nil
}

func coalesceOptions(o []string) []string {
	if o == nil {
		return []string{}
	}
	return o
}

// uptDisplayOptions builds the variableDisplayOptions value for one variable.
func uptDisplayOptions(v uptIOVar) map[string]any {
	return map[string]any{
		"enableMultiLineText":  v.Multiline,
		"enablePictureViewer":  v.Picture,
		"hiddenVariable":       v.Hidden,
		"enableValidation":     v.Validation,
		"useChatbotInterface":  v.Chatbot,
	}
}

func newUptModelsIOSetCmd(flags *rootFlags) *cobra.Command {
	var envFlag, fromFile string
	var iKnow, dryRunOnly bool
	cmd := &cobra.Command{
		Use:   "set <modelId>",
		Short: "Replace a model's inputs/outcomes from a YAML file",
		Long: "Declarative aiEngineConfig: replaces the model's inputs and outcomes from YAML, keeping\n" +
			"the rest of the record intact (the API upsert is full-record; partial bodies clobber the I/O\n" +
			"contract). Each variable's display options (hiddenVariable etc.) are rebuilt; unrelated\n" +
			"variableDisplayOptions keys are preserved.\n\n" +
			"YAML shape:\n" +
			"  inputs:\n" +
			"    - name: InputMessage\n" +
			"      type: String        # inputType: String, File, ...\n" +
			"      description: Free text in\n" +
			"      default: \"\"\n" +
			"      hidden: false       # developer-only variable\n" +
			"  outcomes:\n" +
			"    - name: OutputMessage\n" +
			"      type: String\n\n" +
			"Input names must exactly match the keys the predict body's InputValues accepts.",
		Example: "  unitypredict-pp-cli models io set c6985b70-15e0-4ba9-8cb1-128200595a8f --from io.yaml\n  unitypredict-pp-cli models io set c6985b70-15e0-4ba9-8cb1-128200595a8f --from io.yaml --dry-run",
		Annotations: map[string]string{
			"pp:data-source": "live",
			"pp:happy-args":  "modelId=c6985b70-15e0-4ba9-8cb1-128200595a8f;--from=io.yaml;--dry-run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "models io set")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("modelId is required\nUsage: %s <modelId>", cmd.CommandPath()))
			}
			if fromFile == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--from is required (YAML file defining inputs/outcomes)"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			modelID := args[0]
			env := resolveUptEnv(envFlag)
			if err := guardUptProdMutation(env, "set model inputs/outcomes", iKnow); err != nil {
				return err
			}
			raw, err := os.ReadFile(fromFile) /* #nosec G304 -- user-supplied --from path is the feature; local read only */
			if err != nil {
				return fmt.Errorf("reading --from %s: %w", fromFile, err)
			}
			spec, err := parseUptIOSpec(raw)
			if err != nil {
				return usageErr(err)
			}
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			recordJSON, err := c.GetNoCache(ctx, "/api/models/"+modelID, nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var record map[string]any
			if err := json.Unmarshal(recordJSON, &record); err != nil {
				return fmt.Errorf("parsing model record: %w", err)
			}
			engineCfg, _ := record["aiEngineConfig"].(map[string]any)
			if engineCfg == nil {
				engineCfg = map[string]any{}
			}
			var prevInputs, prevOutcomes []any
			if arr, ok := engineCfg["inputs"].([]any); ok {
				prevInputs = arr
			}
			if arr, ok := engineCfg["outcomes"].([]any); ok {
				prevOutcomes = arr
			}
			if dryRunOnly {
				view := struct {
					ModelID   string   `json:"model_id"`
					DryRun    bool     `json:"dry_run"`
					WouldSet  uptIOSpec `json:"would_set"`
				}{modelID, true, spec}
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}

			toVars := func(vars []uptIOVar) []any {
				out := make([]any, 0, len(vars))
				for _, v := range vars {
					out = append(out, map[string]any{
						"name":              v.Name,
						"inputDescription":  "", // set below per kind
						"inputType":         v.Type,
						"defaultValue":      v.Default,
						"options":           coalesceOptions(v.Options),
					})
				}
				return out
			}
			inputVars := toVars(spec.Inputs)
			for i, v := range spec.Inputs {
				inputVars[i].(map[string]any)["inputDescription"] = v.Description
			}
			outcomeVars := toVars(spec.Outcomes)
			for i, v := range spec.Outcomes {
				m := outcomeVars[i].(map[string]any)
				m["outcomeDescription"] = v.Description
				m["outcomeType"] = v.Type
				delete(m, "inputDescription")
				delete(m, "inputType")
			}
			engineCfg["inputs"] = inputVars
			engineCfg["outcomes"] = outcomeVars
			record["aiEngineConfig"] = engineCfg

			display, _ := record["variableDisplayOptions"].(map[string]any)
			if display == nil {
				display = map[string]any{}
			}
			for _, v := range spec.Inputs {
				display[strings.ToLower(v.Name)+"_inputcon"] = uptDisplayOptions(v)
			}
			for _, v := range spec.Outcomes {
				display[strings.ToLower(v.Name)+"_outputcon"] = uptDisplayOptions(v)
			}
			record["variableDisplayOptions"] = display

			payload, err := json.Marshal(record)
			if err != nil {
				return err
			}
			if _, _, err := c.Post(ctx, "/api/models", json.RawMessage(payload)); err != nil {
				return fmt.Errorf("upserting model record: %w", err)
			}
			// Read-after-write can serve stale data in a warm session; re-read
			// with the cache bypassed.
			verify, err := c.GetNoCache(ctx, "/api/models/"+modelID, nil)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: post-write verification read failed: %v\n", err)
			}
			view := struct {
				ModelID       string      `json:"model_id"`
				Env           string      `json:"env"`
				InputsBefore  int         `json:"inputs_before"`
				OutcomesBefore int        `json:"outcomes_before"`
				InputsAfter   int         `json:"inputs_after"`
				OutcomesAfter int         `json:"outcomes_after"`
				Applied       uptIOSpec   `json:"applied"`
				Verified      bool        `json:"verified"`
			}{
				ModelID: modelID, Env: env,
				InputsBefore: len(prevInputs), OutcomesBefore: len(prevOutcomes),
				InputsAfter: len(spec.Inputs), OutcomesAfter: len(spec.Outcomes),
				Applied: spec, Verified: err == nil && strings.Contains(string(verify), `"name":"`+firstVarName(spec)),
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "model %s: inputs %d -> %d, outcomes %d -> %d (verified: %v)\n",
				modelID, view.InputsBefore, view.InputsAfter, view.OutcomesBefore, view.OutcomesAfter, view.Verified)
			for _, v := range spec.Inputs {
				fmt.Fprintf(cmd.OutOrStdout(), "  in  %-20s %s%s\n", v.Name, v.Type, hiddenMark(v.Hidden))
			}
			for _, v := range spec.Outcomes {
				fmt.Fprintf(cmd.OutOrStdout(), "  out %-20s %s%s\n", v.Name, v.Type, hiddenMark(v.Hidden))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().StringVar(&fromFile, "from", "", "YAML file defining inputs/outcomes (required)")
	cmd.Flags().BoolVar(&iKnow, "i-know", false, "Allow the mutation against prod (off by default: prod is read-only)")
	cmd.Flags().BoolVar(&dryRunOnly, "plan", false, "Parse and show what would change without writing (complements --dry-run)")
	return cmd
}

func hiddenMark(hidden bool) string {
	if hidden {
		return "  [hidden]"
	}
	return ""
}

func firstVarName(spec uptIOSpec) string {
	if len(spec.Inputs) > 0 {
		return spec.Inputs[0].Name
	}
	if len(spec.Outcomes) > 0 {
		return spec.Outcomes[0].Name
	}
	return ""
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if modelsCmd, _, err := root.Find([]string{"models"}); err == nil {
			ioParent := &cobra.Command{
				Use:   "io",
				Short: "Model input/outcome configuration (aiEngineConfig)",
				Example: "  unitypredict-pp-cli models io set <modelId> --from io.yaml",
			}
			ioParent.AddCommand(newUptModelsIOSetCmd(flags))
			addNovelCommandIfAbsent(modelsCmd, ioParent)
		}
	})
}
