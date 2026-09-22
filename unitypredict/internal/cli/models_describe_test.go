// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored test (printing-press preserved file).

package cli

// The defect this file pins: `models describe get` used to read the record's
// modelLongDescription URL (an /api/public/... mirror) and best-effort fetch
// it — which 401s for every PRIVATE model even when authenticated, so the
// command reported content_bytes: 0 with a fetch_error and exited clean. The
// authenticated filekey route (302 -> presigned S3, no Authorization on the
// S3 hop) serves private and public models alike and is the only primary.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDescribeGetServesPrivateModelViaFileKeyRoute(t *testing.T) {
	const modelID = "c6985b70-15e0-4ba9-8cb1-128200595a8f" // private scratch model shape
	var mu sync.Mutex
	var paths []string
	authFileKey, authS3 := "", "UNSET"

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		switch r.URL.Path {
		case "/api/models/" + modelID + "/filekey/MODELDESCRIPTION":
			authFileKey = r.Header.Get("Authorization")
			http.Redirect(w, r, srv.URL+"/s3-presigned/longdescription.md", http.StatusFound)
		case "/s3-presigned/longdescription.md":
			authS3 = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# UnityPredict CLI\n\nprivate model long description body\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// Point dev credentials at the test server (same fixture pattern as
	// TestUptEnvClientKeepsBracketedKey).
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".unitypredict"), 0o700); err != nil {
		t.Fatal(err)
	}
	creds := `{"dev":{"UPT_API_KEY":"testkey-not-a-real-secret","API_URL":"` + srv.URL + `"}}`
	if err := os.WriteFile(filepath.Join(tmp, ".unitypredict", "credentials"), []byte(creds), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmp)

	f := &rootFlags{timeout: 5 * time.Second}
	c, err := uptEnvClient(f, "dev")
	if err != nil {
		t.Fatalf("env client build failed: %v", err)
	}
	data, code, err := uptGetAuthedToPresigned(context.Background(), f, c, "/api/models/"+modelID+"/filekey/MODELDESCRIPTION")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	// The defect's observable: a private model must yield content, not a
	// fetch_error. content_bytes == len(data) in the command's view.
	if len(data) == 0 {
		t.Fatal("content_bytes = 0 for a private model — describe get is broken again")
	}
	if !strings.Contains(string(data), "private model long description body") {
		t.Fatalf("unexpected body: %q", string(data)[:min(80, len(data))])
	}
	if code != 200 {
		t.Fatalf("final status = %d, want 200", code)
	}

	// Route discipline: filekey requested, /api/public/ never touched.
	mu.Lock()
	defer mu.Unlock()
	for _, p := range paths {
		if strings.Contains(p, "/api/public/") {
			t.Fatalf("public mirror route requested: %s (401s for private models)", p)
		}
	}
	if !containsPath(paths, "/api/models/"+modelID+"/filekey/MODELDESCRIPTION") {
		t.Fatalf("filekey route never requested; paths=%v", paths)
	}

	// Auth discipline: first hop authenticated, presigned hop clean.
	if !strings.HasPrefix(authFileKey, "Bearer APIKEY@") {
		t.Fatalf("filekey route missing durable-key auth header: %q", authFileKey)
	}
	if authS3 != "" {
		t.Fatalf("Authorization leaked on presigned S3 hop: %q", authS3)
	}
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}
