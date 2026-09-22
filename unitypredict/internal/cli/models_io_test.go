// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored tests for the novel IO-spec parser (printing-press preserved file).

package cli

import (
	"strings"
	"testing"
)

func TestParseUptIOSpec(t *testing.T) {
	valid := `
inputs:
  - name: SubtitleFile
    type: File
    description: The subtitle file to translate
    default: ""
    options: []
    hidden: false
  - name: TargetLanguage
    type: String
    description: Language to translate into
    default: Hindi
    options: [English, Hindi, Bengali]
outcomes:
  - name: TranslatedText
    type: String
    description: Translated subtitle text
    hidden: true
`
	spec, err := parseUptIOSpec([]byte(valid))
	if err != nil {
		t.Fatalf("valid spec rejected: %v", err)
	}
	if len(spec.Inputs) != 2 || len(spec.Outcomes) != 1 {
		t.Fatalf("counts wrong: inputs=%d outcomes=%d", len(spec.Inputs), len(spec.Outcomes))
	}
	if spec.Inputs[0].Name != "SubtitleFile" || spec.Inputs[0].Type != "File" {
		t.Errorf("input[0] = %+v", spec.Inputs[0])
	}
	if spec.Inputs[1].Default != "Hindi" || len(spec.Inputs[1].Options) != 3 {
		t.Errorf("input[1] default/options not parsed: %+v", spec.Inputs[1])
	}
	if !spec.Outcomes[0].Hidden {
		t.Error("outcome hidden flag not parsed")
	}
	if spec.Inputs[0].Options == nil {
		t.Error("nil options must coerce to [] for JSON [] not null")
	}
}

func TestParseUptIOSpecErrors(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want string
	}{
		{"missing name", "inputs:\n  - type: String\n", "name is required"},
		{"missing type", "inputs:\n  - name: A\n", "type is required"},
		{"duplicate", "inputs:\n  - name: A\n    type: String\n  - name: A\n    type: File\n", "duplicate"},
		{"garbage", "inputs: [", "parsing IO yaml"},
	}
	for _, tc := range cases {
		_, err := parseUptIOSpec([]byte(tc.yaml))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err=%v want substring %q", tc.name, err, tc.want)
		}
	}
}

func TestUptDisplayOptionsShape(t *testing.T) {
	got := uptDisplayOptions(uptIOVar{Name: "x", Hidden: true, Multiline: true})
	wantKeys := []string{"enableMultiLineText", "enablePictureViewer", "hiddenVariable", "enableValidation", "useChatbotInterface"}
	m := map[string]bool{}
	for k := range got {
		m[k] = true
	}
	for _, k := range wantKeys {
		if !m[k] {
			t.Errorf("display options missing %s", k)
		}
	}
	if got["hiddenVariable"] != true || got["enableMultiLineText"] != true {
		t.Errorf("flags not mapped: %+v", got)
	}
}
