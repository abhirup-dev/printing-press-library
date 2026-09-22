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
	"sync/atomic"
	"testing"
	"time"

	"tijori-finance-pp-cli/internal/cliutil/testenv"
)

// TestNovelResearchResultPackHelpWires smoke-tests that the research result-pack command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelResearchResultPackHelpWires(t *testing.T) {
	testenv.Isolate(t)
	cmd := RootCmd()
	cmd.SetArgs([]string{"research", "result-pack", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("research result-pack --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "result-pack"} {
		if !strings.Contains(help, want) {
			t.Fatalf("research result-pack --help missing %q in output:\n%s", want, help)
		}
	}
}

func TestResearchResultPackUsesOneCommandWideTimeout(t *testing.T) {
	testenv.Isolate(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(2 * time.Second):
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html><html><head><title>Late</title></head><body></body></html>`))
		}
	}))
	defer server.Close()
	t.Setenv("TIJORI_FINANCE_BASE_URL", server.URL)

	cmd := RootCmd()
	cmd.SetArgs([]string{"research", "result-pack", "tata-steel-limited", "--period", "Jun-26", "--agent", "--timeout", "50ms"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	started := time.Now()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("timed-out result-pack should return a partial envelope, got %v; stderr=%s", err, errOut.String())
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("result-pack ignored command-wide timeout: elapsed=%s", elapsed)
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("expected partial JSON envelope, got %q: %v", out.String(), err)
	}
	results, _ := envelope["results"].(map[string]any)
	company, _ := results["company_snapshot"].(map[string]any)
	if company["status"] != "error" || company["availability"] != "upstream_error" {
		t.Fatalf("company timeout lacks explicit availability: %#v", company)
	}
	sources, _ := results["sources"].(map[string]any)
	for _, name := range []string{"reports", "concalls", "company_documents"} {
		source, _ := sources[name].(map[string]any)
		if source["status"] != "error" || source["availability"] != "upstream_error" {
			t.Fatalf("%s timeout lacks explicit availability: %#v", name, source)
		}
	}
}

func TestResearchResultPackParsesHTMLSources(t *testing.T) {
	testenv.Isolate(t)
	var inFlight, maxInFlight atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			maximum := maxInFlight.Load()
			if current <= maximum || maxInFlight.CompareAndSwap(maximum, current) {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><title>Tijori</title><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{}}}</script></head><body><a href="/company/tata-steel-limited/">Company</a><a href="/results/quarterly-results/">Results</a></body></html>`))
	}))
	defer server.Close()
	t.Setenv("TIJORI_FINANCE_BASE_URL", server.URL)

	cmd := RootCmd()
	cmd.SetArgs([]string{"research", "result-pack", "tata-steel-limited", "--period", "Jun-26", "--agent"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("result-pack HTML fixture error = %v; stderr=%s", err, errOut.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("expected JSON envelope, got %q: %v", out.String(), err)
	}
	results, ok := envelope["results"].(map[string]any)
	if !ok {
		t.Fatalf("missing results envelope: %#v", envelope)
	}
	sources, ok := results["sources"].(map[string]any)
	if !ok {
		t.Fatalf("missing source results: %#v", results)
	}
	for _, name := range []string{"reports", "concalls", "company_documents"} {
		source, ok := sources[name].(map[string]any)
		if !ok {
			t.Fatalf("missing %s source result: %#v", name, sources[name])
		}
		if source["status"] != "ok" || source["availability"] != "available" || source["stability"] != "observed" {
			t.Fatalf("%s lacks actionable HTML metadata: %#v", name, source)
		}
	}
	if got := maxInFlight.Load(); got < 3 {
		t.Fatalf("result-pack max concurrent requests = %d, want at least 3", got)
	}
}
