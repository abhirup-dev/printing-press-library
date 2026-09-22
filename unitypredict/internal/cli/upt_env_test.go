// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored tests for the upt novel helpers (printing-press preserved file).

package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestUptExtractURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"uploadLink field", `{"uploadLink":"https://s3/presigned?X-Amz-Signature=abc","fileName":"f.md"}`, "https://s3/presigned?X-Amz-Signature=abc"},
		{"bare url body", "\"https://s3/x\"\n", "https://s3/x"},
		{"nested logUrl", `{"logs":{"logUrl":"https://s3/log.txt"}}`, "https://s3/log.txt"},
		{"no url", `{"error":"none"}`, ""},
		{"non-json", `boom`, ""},
	}
	for _, tc := range cases {
		if got := uptExtractURL([]byte(tc.in)); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestStripEnginePrefix(t *testing.T) {
	if got := stripEnginePrefix("AppEngineDefinition-1d0e3201-a88e-4ed8-af73-84abd7d3e474"); got != "1d0e3201-a88e-4ed8-af73-84abd7d3e474" {
		t.Errorf("prefixed id not stripped: %q", got)
	}
	if got := stripEnginePrefix("1d0e3201-a88e-4ed8-af73-84abd7d3e474"); got != "1d0e3201-a88e-4ed8-af73-84abd7d3e474" {
		t.Errorf("bare id mangled: %q", got)
	}
}

func TestResolveUptEnv(t *testing.T) {
	t.Setenv("UNITYPREDICT_ENV", "")
	if got := resolveUptEnv(""); got != "dev" {
		t.Errorf("default env = %q, want dev", got)
	}
	if got := resolveUptEnv("PROD"); got != "prod" {
		t.Errorf("flag prod (case) = %q, want prod", got)
	}
	t.Setenv("UNITYPREDICT_ENV", "prod")
	if got := resolveUptEnv(""); got != "prod" {
		t.Errorf("env var prod = %q, want prod", got)
	}
	if got := resolveUptEnv("dev"); got != "dev" {
		t.Errorf("flag wins over env = %q, want dev", got)
	}
}

func TestUptIsProdHost(t *testing.T) {
	if !uptIsProdHost("https://api.prod.unitypredict.com") {
		t.Error("prod host not detected")
	}
	if uptIsProdHost("https://api.dev.unitypredict.net") {
		t.Error("dev host misclassified as prod")
	}
}

// The presigned hop must never carry an Authorization header: S3 400s and echoes
// the credential back. Pin it.
func TestUptFetchPresignedSendsNoAuth(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	defer srv.Close()
	data, code, err := uptFetchPresigned(context.Background(), http.MethodGet, srv.URL+"/signed", nil)
	if err != nil || code != 200 || string(data) != "content" {
		t.Fatalf("fetch failed: code=%d err=%v data=%q", code, err, data)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization leaked on presigned fetch: %q", sawAuth)
	}
}

func TestGuardUptProdMutation(t *testing.T) {
	if err := guardUptProdMutation("dev", "create a model", false); err != nil {
		t.Errorf("dev guard fired: %v", err)
	}
	err := guardUptProdMutation("prod", "create a model", false)
	if err == nil || !strings.Contains(err.Error(), "--i-know") {
		t.Fatalf("prod guard missing --i-know hint: %v", err)
	}
	if err := guardUptProdMutation("prod", "create a model", true); err != nil {
		t.Errorf("--i-know override refused: %v", err)
	}
	t.Setenv("PRINTING_PRESS_VERIFY", "1")
	if err := guardUptProdMutation("prod", "create a model", true); err == nil {
		t.Fatal("harness must hard-refuse prod mutation even with --i-know")
	}
	_ = os.Unsetenv("PRINTING_PRESS_VERIFY")
}

// The regression the user flagged: a prod key containing { [ & * ) must survive
// auth-header construction at the client layer too (config layer is covered by
// internal/config/authfmt_test.go).
func TestUptEnvClientKeepsBracketedKey(t *testing.T) {
	f := &rootFlags{timeout: 5 * time.Second}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	// ponytail: writes a throwaway HOME with a fixture creds file; restores after.
	orig := home
	tmp := t.TempDir()
	if err := os.MkdirAll(tmp+"/.unitypredict", 0o700); err != nil {
		t.Fatal(err)
	}
	key := "k{y[z&w*x)v"
	if err := os.WriteFile(tmp+"/.unitypredict/credentials", []byte(`{"dev":{"UPT_API_KEY":"`+key+`","API_URL":"http://127.0.0.1:1"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmp)
	_ = orig
	c, err := uptEnvClient(f, "dev")
	if err != nil {
		t.Fatalf("env client build failed: %v", err)
	}
	got := c.Config.AuthHeader()
	want := "Bearer APIKEY@" + key
	if got != want {
		t.Fatalf("auth header mangled for bracketed key\n got: %q\nwant: %q", got, want)
	}
}
