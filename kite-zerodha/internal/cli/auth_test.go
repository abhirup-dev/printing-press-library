// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestExchangeKiteRequestToken(t *testing.T) {
	const apiKey = "test-api-key"
	const apiSecret = "test-api-secret"
	const requestToken = "test-request-token"
	const accessToken = "test-access-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("X-Kite-Version"); got != "3" {
			t.Errorf("X-Kite-Version = %q, want 3", got)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/x-www-form-urlencoded") {
			t.Errorf("Content-Type = %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("api_key"); got != apiKey {
			t.Errorf("api_key = %q, want test value", got)
		}
		if got := r.Form.Get("request_token"); got != requestToken {
			t.Errorf("request_token = %q, want test value", got)
		}
		sum := sha256.Sum256([]byte(apiKey + requestToken + apiSecret))
		if got := r.Form.Get("checksum"); got != hex.EncodeToString(sum[:]) {
			t.Errorf("checksum mismatch")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"access_token":"` + accessToken + `"}}`))
	}))
	defer server.Close()

	got, err := exchangeKiteRequestToken(t.Context(), server.URL, apiKey, apiSecret, requestToken)
	if err != nil {
		t.Fatalf("exchangeKiteRequestToken: %v", err)
	}
	if got != accessToken {
		t.Errorf("access token mismatch")
	}
}

func TestValidateKiteRedirectURI(t *testing.T) {
	valid := []string{
		"http://127.0.0.1:8765/callback",
		"http://localhost:8765/callback",
		"http://[::1]:8765/callback",
	}
	for _, raw := range valid {
		if err := validateKiteRedirectURI(raw); err != nil {
			t.Errorf("validateKiteRedirectURI(%q): %v", raw, err)
		}
	}
	invalid := []string{
		"https://127.0.0.1:8765/callback",
		"http://example.com:8765/callback",
		"http://127.0.0.1/callback",
		"http://127.0.0.1:8765/callback?x=1",
	}
	for _, raw := range invalid {
		if err := validateKiteRedirectURI(raw); err == nil {
			t.Errorf("validateKiteRedirectURI(%q) accepted invalid URI", raw)
		}
	}
}

func TestKiteLoginURL(t *testing.T) {
	got, err := kiteLoginURL("test-api-key")
	if err != nil {
		t.Fatalf("kiteLoginURL: %v", err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse login URL: %v", err)
	}
	if u.Host != "kite.zerodha.com" || u.Path != "/connect/login" {
		t.Fatalf("unexpected login URL route")
	}
	if u.Query().Get("v") != "3" || u.Query().Get("api_key") != "test-api-key" {
		t.Fatalf("unexpected login URL parameters")
	}
}
