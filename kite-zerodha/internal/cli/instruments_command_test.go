// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kite-zerodha-pp-cli/internal/cliutil"
)

func TestInstrumentsExchangeCommandParsesCSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/instruments/NSE" && r.URL.Path != "/instruments" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		_, _ = fmt.Fprint(w, "exchange,tradingsymbol,name,last_price\nNSE,ABC,Acme,123.45\n")
	}))
	defer server.Close()

	home := t.TempDir()
	restoreHome, err := cliutil.SetHomeOverride("")
	if err != nil {
		t.Fatal(err)
	}
	defer restoreHome()
	if err := os.MkdirAll(filepath.Join(home, "config"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "config", "config.toml"), []byte("base_url = \""+server.URL+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := RootCmd()
	cmd.SetArgs([]string{"--home", home, "instruments", "exchange", "NSE", "--agent", "--select", "exchange,tradingsymbol"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("exchange command: %v; stderr=%s", err, errOut.String())
	}
	if !strings.Contains(out.String(), `"tradingsymbol": "ABC"`) {
		t.Fatalf("JSON output missing parsed row: %s", out.String())
	}
	if !strings.Contains(out.String(), `"source": "live"`) || !strings.Contains(out.String(), `observed_at`) {
		t.Fatalf("JSON output missing provenance metadata: %s", out.String())
	}

	list := RootCmd()
	list.SetArgs([]string{"--home", home, "instruments", "list", "--agent", "--select", "exchange,tradingsymbol"})
	var listOut, listErrOut bytes.Buffer
	list.SetOut(&listOut)
	list.SetErr(&listErrOut)
	if err := list.Execute(); err != nil {
		t.Fatalf("list command: %v; stderr=%s", err, listErrOut.String())
	}
	if !strings.Contains(listOut.String(), `"tradingsymbol": "ABC"`) {
		t.Fatalf("list JSON output missing parsed row: %s", listOut.String())
	}
}
