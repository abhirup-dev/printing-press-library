package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tijori-finance-pp-cli/internal/cliutil/testenv"
)

func TestFeedCompanyUsesProviderCompanyIDParameter(t *testing.T) {
	testenv.Isolate(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("company_id"); got != "338" {
			t.Errorf("company_id = %q, want 338", got)
		}
		if got := r.URL.Query().Get("id"); got != "" {
			t.Errorf("legacy id parameter must be absent, got %q", got)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><title>Tata Steel events</title></head><body></body></html>`))
	}))
	defer server.Close()
	t.Setenv("TIJORI_FINANCE_BASE_URL", server.URL)

	cmd := RootCmd()
	cmd.SetArgs([]string{"feed", "company", "--id", "338", "--agent"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("feed company error = %v; stderr=%s", err, errOut.String())
	}
}

func TestCompanyFinancialsRejectsInvalidSlugBeforeHTTP(t *testing.T) {
	testenv.Isolate(t)
	cmd := RootCmd()
	cmd.SetArgs([]string{"company", "financials", "__printing_press_invalid__"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("invalid company slug unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "invalid company slug") {
		t.Fatalf("unexpected error: %v", err)
	}
}
