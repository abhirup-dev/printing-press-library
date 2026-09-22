// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// cli-printing-press: novel-scaffold-test
// Novel command scaffold tests. Keep the wiring smoke test and add behavior cases as needed.

package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"tickertape-pp-cli/internal/client"
	"tickertape-pp-cli/internal/cliutil/testenv"
	"tickertape-pp-cli/internal/config"
)

// TestNovelCompanyBriefHelpWires smoke-tests that the company brief command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelCompanyBriefHelpWires(t *testing.T) {
	testenv.Isolate(t)
	cmd := RootCmd()
	cmd.SetArgs([]string{"company", "brief", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("company brief --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "brief"} {
		if !strings.Contains(help, want) {
			t.Fatalf("company brief --help missing %q in output:\n%s", want, help)
		}
	}
}

func TestCompanyBriefAISummaryUsesDirectAuthenticatedRoute(t *testing.T) {
	testenv.Isolate(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BaseURL: server.URL,
		Headers: map[string]string{
			"Cookie":        "jwt=test-session",
			"x-csrf-token":  "test-csrf",
			"x-device-type": "web",
		},
	}
	c := client.New(cfg, 5e9, client.RateLimitAuto)
	c.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() != "analyze.api.tickertape.in" {
			t.Fatalf("AI summary used host %q, want analyze.api.tickertape.in", req.URL.Hostname())
		}
		if got := req.Header.Get("Accept-Version"); got != "8.14.0" {
			t.Fatalf("AI summary Accept-Version = %q, want 8.14.0", got)
		}
		if got := req.Header.Get("x-csrf-token"); got != "test-csrf" {
			t.Fatalf("AI summary lost authenticated CSRF header: %q", got)
		}
		req.URL.Scheme = serverURL.Scheme
		req.URL.Host = serverURL.Host
		return http.DefaultTransport.RoundTrip(req)
	})

	result := fetchResearchEndpoint(context.Background(), c, "ai_summary", companyBriefPaths("RELI")["ai_summary"], nil)
	if got := result["access_state"]; got != "ok" {
		t.Fatalf("company brief AI summary access_state = %v, want ok (result=%#v)", got, result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
