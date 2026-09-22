// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel extension (printing-press preserved file). DO NOT EDIT generated siblings.

package cli

// UnityPredict runs two isolated tenants (dev/prod) with per-env durable keys in
// ~/.unitypredict/credentials. The generated config models a single base URL; this
// file adds env awareness for novel commands: env-resolved client construction, the
// prod read-only guard (manifest T10), presigned-S3 fetches that must never carry
// Authorization, and small helpers shared by the hand-coded commands.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"unitypredict-pp-cli/internal/cliutil"
	"unitypredict-pp-cli/internal/client"
	"unitypredict-pp-cli/internal/config"
)

const (
	uptEnvDev  = "dev"
	uptEnvProd = "prod"

	// uptDefaultWait is the poll budget for predict --wait when the user has
	// not set an explicit --timeout. GPU cold starts run minutes-long; the
	// root default (60s) is a per-request ceiling, not a sane wait budget.
	uptDefaultWait = 15 * time.Minute
)

func uptDefaultBaseURL(env string) string {
	if env == uptEnvProd {
		return "https://api.prod.unitypredict.com"
	}
	return "https://api.dev.unitypredict.net"
}

type uptEnvCredential struct {
	UPTAPIKey string `json:"UPT_API_KEY"`
	APIURL    string `json:"API_URL"`
}

// loadUptEnvCredential reads one tenant's key from ~/.unitypredict/credentials.
// The key value is never logged or returned in error text.
func loadUptEnvCredential(env string) (uptEnvCredential, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return uptEnvCredential{}, fmt.Errorf("resolving home directory: %w", err)
	}
	raw, err := os.ReadFile(filepath.Join(home, ".unitypredict", "credentials"))
	if err != nil {
		return uptEnvCredential{}, fmt.Errorf("reading ~/.unitypredict/credentials: %w", err)
	}
	var byEnv map[string]uptEnvCredential
	if err := json.Unmarshal(raw, &byEnv); err != nil {
		return uptEnvCredential{}, fmt.Errorf("parsing ~/.unitypredict/credentials (expected JSON keyed dev/prod): %w", err)
	}
	cred, ok := byEnv[env]
	if !ok || cred.UPTAPIKey == "" {
		return uptEnvCredential{}, fmt.Errorf("no %s entry with a UPT_API_KEY in ~/.unitypredict/credentials", env)
	}
	if cred.APIURL == "" {
		cred.APIURL = uptDefaultBaseURL(env)
	}
	return cred, nil
}

// resolveUptEnv picks the tenant: explicit flag > UNITYPREDICT_ENV > dev.
func resolveUptEnv(flagVal string) string {
	if v := strings.ToLower(strings.TrimSpace(flagVal)); v == uptEnvDev || v == uptEnvProd {
		return v
	}
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("UNITYPREDICT_ENV"))); v == uptEnvDev || v == uptEnvProd {
		return v
	}
	return uptEnvDev
}

// uptEnvClient builds a generated transport pointed at one tenant, authenticated
// with that tenant's durable key (Bearer APIKEY@<key>). Reuses the generated
// client so redirects, retries, caching, and rate limiting behave identically.
func uptEnvClient(flags *rootFlags, env string) (*client.Client, error) {
	cred, err := loadUptEnvCredential(env)
	if err != nil {
		// Fallback for environments without ~/.unitypredict/credentials (CI,
		// sandboxes): per-tenant env vars, dev and prod alike (diff needs both).
		tok := ""
		if env == uptEnvDev {
			tok = os.Getenv("DEV_UNITYPREDICT_TOKEN")
		} else if env == uptEnvProd {
			tok = os.Getenv("PROD_UNITYPREDICT_TOKEN")
		}
		if tok != "" {
			cred = uptEnvCredential{UPTAPIKey: tok, APIURL: uptDefaultBaseURL(env)}
		} else {
			return nil, configErr(err)
		}
	}
	cfg := &config.Config{
		BaseURL:     cred.APIURL,
		AuthHeaderVal: "Bearer APIKEY@" + cred.UPTAPIKey,
	}
	if v := os.Getenv("UNITYPREDICT_BASE_URL"); v != "" {
		// Verify/mock harness override wins so tests can point this client anywhere.
		cfg.BaseURL = v
	}
	if flags.insecure {
		cfg.SetSkipTLSVerify(true)
	}
	c := client.New(cfg, flags.timeout, flags.rateLimit)
	if flags.timeoutExplicit {
		c.SetTimeoutExplicit(true)
	}
	c.DryRun = flags.dryRun
	c.NoCache = flags.noCache
	if err := ApplyClientHooks(c); err != nil {
		return nil, err
	}
	return c, nil
}

// uptIsProdHost reports whether a base URL points at the prod tenant.
func uptIsProdHost(baseURL string) bool {
	u := strings.ToLower(baseURL)
	return strings.Contains(u, "://api.prod.unitypredict.com") || strings.HasPrefix(u, "api.prod.unitypredict.com")
}

// uptRefusalErr is the typed exit (6) for the prod read-only guard.
func uptRefusalErr(err error) error { return &cliError{code: 6, err: err} }

// guardUptProdMutation implements manifest T10 for env-aware novel commands:
// mutating verbs refuse against prod unless --i-know. Under a verify/dogfood
// harness the refusal is absolute — a live matrix must never touch prod.
func guardUptProdMutation(env, action string, iKnow bool) error {
	if env != uptEnvProd {
		return nil
	}
	if cliutil.IsAnyHarness() {
		return uptRefusalErr(fmt.Errorf("refusing to %s on prod: mutations are disabled under the verify/dogfood harness", action))
	}
	if !iKnow {
		return uptRefusalErr(fmt.Errorf("refusing to %s on prod: prod is read-only by default; pass --i-know to override", action))
	}
	return nil
}

// installUptProdGuard wraps every generated mutating command (pp:method POST/PUT/
// PATCH/DELETE) so pointing the CLI at the prod host refuses the same way. The
// env-aware novel commands guard themselves with the tenant name; this wrapper
// covers spec-emitted mutations that reach prod via config base_url or
// UNITYPREDICT_BASE_URL. Override: UNITYPREDICT_I_KNOW=1. Dry-run passes through
// (no network is touched).
func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		wrapMutatingCommands(root, flags)
	})
}

func wrapMutatingCommands(cmd *cobra.Command, flags *rootFlags) {
	for _, child := range cmd.Commands() {
		wrapMutatingCommands(child, flags)
	}
	switch cmd.Annotations["pp:method"] {
	case "POST", "PUT", "PATCH", "DELETE":
	default:
		return
	}
	if cmd.RunE == nil || cmd.Annotations["pp:upt-prod-guard"] == "true" {
		return
	}
	cmd.Annotations["pp:upt-prod-guard"] = "true"
	orig := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if !flags.dryRun && !uptHarnessOverride() {
			if cfg, err := config.Load(flags.configPath); err == nil && uptIsProdHost(cfg.BaseURL) {
				return uptRefusalErr(fmt.Errorf("refusing %s on prod host %s: prod is read-only by default; set UNITYPREDICT_I_KNOW=1 to override", c.CommandPath(), cfg.BaseURL))
			}
		}
		return orig(c, args)
	}
}

func uptHarnessOverride() bool { return os.Getenv("UNITYPREDICT_I_KNOW") == "1" }

// stripEnginePrefix removes the AppEngineDefinition- prefix that some responses
// (models/requests results) carry; every route that takes an engine id wants the
// bare uuid and 404s misleadingly on the prefixed form.
func stripEnginePrefix(id string) string {
	return strings.TrimPrefix(id, "AppEngineDefinition-")
}

// uptUUIDRe matches the platform's canonical ids (bare uuid; engines may carry
// the AppEngineDefinition- prefix stripped before matching).
var uptUUIDRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func uptIsUUID(s string) bool { return uptUUIDRe.MatchString(s) }

// uptS3Limiter paces the direct presigned-S3 hops (upload PUTs, log GETs,
// outcome downloads). One-shot calls, but the per-source-rate-limiting rule
// applies to every hand-written outbound client.
var uptS3Limiter = cliutil.NewAdaptiveLimiterAuto(2.0)

const uptMaxPresignedBytes = 64 << 20 // 64 MiB cap on any single presigned body

// uptFetchPresigned performs the S3 hop for a presigned URL. The signature lives
// in the query string; an Authorization header breaks it (400) and the error body
// echoes the credential back — so no auth is ever attached here. This is the
// correctness requirement behind the whole presigned flow.
func uptFetchPresigned(ctx context.Context, method, rawURL string, body []byte) ([]byte, int, error) {
	if err := uptS3Limiter.Wait(ctx); err != nil {
		return nil, 0, fmt.Errorf("rate limiter: %w", err)
	}
	var rdr io.Reader
	if body != nil {
		rdr = strings.NewReader(string(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.ContentLength = int64(len(body))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, uptMaxPresignedBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	uptS3Limiter.OnSuccess()
	if resp.StatusCode == http.StatusTooManyRequests {
		uptS3Limiter.OnRateLimit()
		return nil, resp.StatusCode, &cliutil.RateLimitError{URL: rawURL, RetryAfter: 5 * time.Second, Body: string(data[:min(len(data), 512)])}
	}
	return data, resp.StatusCode, nil
}

// uptDownloadAuthedThenPresigned fetches an authed API URL that 302s to presigned
// S3 (File outcomes). The first request carries the durable key; the redirect hop
// drops Authorization (cross-host credential leak vector — the S3 error body
// echoes any header it rejects). Streams to dst; returns the bytes written.
// uptGetAuthedToPresigned fetches an API path with the durable key and follows
// the 302 to the presigned S3 URL, dropping Authorization on every redirect
// hop (the presigned hop must never see credentials). Returns the final body.
// This is how the filekey routes (GET /api/{entity}/{id}/filekey/<KEY>) serve
// hosted files for PRIVATE models — the /api/public/... mirror 401s for them
// even when authenticated, so the authed route is the only correct primary.
func uptGetAuthedToPresigned(ctx context.Context, flags *rootFlags, c *client.Client, path string) ([]byte, int, error) {
	if err := uptS3Limiter.Wait(ctx); err != nil {
		return nil, 0, fmt.Errorf("rate limiter: %w", err)
	}
	tc := &http.Client{Timeout: flags.timeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		req.Header.Del("Authorization")
		return nil
	}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.Config.BaseURL, "/")+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", c.Config.AuthHeader())
	resp, err := tc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, uptMaxPresignedBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	uptS3Limiter.OnSuccess()
	return data, resp.StatusCode, nil
}

func uptDownloadAuthedThenPresigned(ctx context.Context, flags *rootFlags, baseURL, authHeader, path, dstFile string) (int64, error) {
	if err := uptS3Limiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limiter: %w", err)
	}
	tc := &http.Client{Timeout: flags.timeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		req.Header.Del("Authorization")
		return nil
	}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := tc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return 0, fmt.Errorf("download %s: HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	uptS3Limiter.OnSuccess()
	f, err := os.Create(dstFile)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, io.LimitReader(resp.Body, uptMaxPresignedBytes))
}

// uptExtractURL finds the presigned URL inside a JSON response whose field name
// the API is not consistent about (uploadLink, logUrl, url, presignedUrl, or a
// bare string body). Returns "" when nothing URL-shaped is present.
func uptExtractURL(data []byte) string {
	trimmed := strings.Trim(strings.TrimSpace(string(data)), "\" \n\t")
	if strings.HasPrefix(trimmed, "http") {
		return trimmed
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	var best string
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			if strings.HasPrefix(t, "https://") && len(t) > len(best) {
				best = t
			}
		case map[string]any:
			for _, vv := range t {
				walk(vv)
			}
		}
	}
	walk(m)
	return best
}

// init registers the credentials-file bridge: when the generated config has no
// credential (fresh install, no auth set-token, no env var), fall back to the
// first-party ~/.unitypredict/credentials store so every command — generated
// and novel — works out of the box. The key is matched to the tenant the
// client's base URL points at; config/env credentials always win.
func init() {
	registerClientHook(func(c *client.Client) error {
		if c == nil || c.Config == nil || c.Config.AuthHeader() != "" {
			return nil
		}
		env := uptEnvDev
		if uptIsProdHost(c.Config.BaseURL) {
			env = uptEnvProd
		}
		cred, err := loadUptEnvCredential(env)
		if err != nil {
			return nil // no creds file or no entry: leave the standard auth flow in charge
		}
		c.Config.AuthHeaderVal = "Bearer APIKEY@" + cred.UPTAPIKey
		return nil
	})
}
