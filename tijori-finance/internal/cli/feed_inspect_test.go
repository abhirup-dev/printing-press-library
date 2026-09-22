// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// cli-printing-press: novel-scaffold-test
// Novel command scaffold tests. Keep the wiring smoke test and add behavior cases as needed.

package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tijori-finance-pp-cli/internal/cliutil/testenv"
)

// TestNovelFeedInspectHelpWires smoke-tests that the feed inspect command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelFeedInspectHelpWires(t *testing.T) {
	testenv.Isolate(t)
	cmd := RootCmd()
	cmd.SetArgs([]string{"feed", "inspect", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("feed inspect --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "inspect"} {
		if !strings.Contains(help, want) {
			t.Fatalf("feed inspect --help missing %q in output:\n%s", want, help)
		}
	}
}

func TestFeedInspectJSONOnEmptyTimelineBody(t *testing.T) {
	testenv.Isolate(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	}))
	defer server.Close()
	t.Setenv("TIJORI_FINANCE_BASE_URL", server.URL)

	flags := &rootFlags{}
	cmd := newRootCmd(flags)
	cmd.SetArgs([]string{"feed", "inspect", "--company", "tata-steel-limited", "--json"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("empty timeline should be a structured non-error availability state: %v", err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("expected JSON error envelope, got %q: %v", out.String(), err)
	}
	results, ok := envelope["results"].(map[string]any)
	if !ok {
		t.Fatalf("missing results envelope: %#v", envelope)
	}
	for _, field := range []string{"availability", "endpoint", "observed_at", "stability", "cause"} {
		value, _ := results[field].(string)
		if strings.TrimSpace(value) == "" {
			t.Fatalf("missing structured field %q: %#v", field, results)
		}
	}
	if results["endpoint"] != "/in/timeline/self/more" {
		t.Fatalf("unexpected endpoint: %#v", results["endpoint"])
	}
}
