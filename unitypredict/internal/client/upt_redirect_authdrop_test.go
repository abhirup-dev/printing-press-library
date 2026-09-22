// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored regression test (printing-press preserved file).

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"unitypredict-pp-cli/internal/config"
)

// UnityPredict presigned routes (filekey, devLog, download, upload) 302 to S3.
// S3 rejects a riding Authorization header with 400 AND echoes the credential
// back in the error body. The redirect handler must therefore drop the header
// on any host change. This pins that contract: the origin request is authed,
// the foreign hop must not be.
func TestRedirectDropsAuthorizationOnHostChange(t *testing.T) {
	var foreignSawAuth string
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		foreignSawAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer foreign.Close()

	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("origin request was not authenticated; test setup is broken")
		}
		http.Redirect(w, r, foreign.URL+"/object", http.StatusFound)
	}))
	defer origin.Close()

	cfg := &config.Config{BaseURL: origin.URL, AuthHeaderVal: "Bearer APIKEY@<test-key>"}
	c := New(cfg, 0, 0)
	// The generated transport forces h2 prior-knowledge, which httptest's
	// plain servers reject; swap in the TLS test client but KEEP the
	// CheckRedirect policy under test.
	tc := origin.Client()
	tc.CheckRedirect = c.HTTPClient.CheckRedirect
	c.HTTPClient = tc
	c.NoCache = true
	if _, err := c.Get(t.Context(), "/api/predict/download/x/y", nil); err != nil {
		t.Fatalf("get through redirect: %v", err)
	}
	if foreignSawAuth != "" {
		t.Fatalf("Authorization header leaked to cross-host redirect: %q", foreignSawAuth)
	}
}
