// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// Package gptapi is the session-aware ChatGPT web API client used by the
// hand-built novel commands. It layers three things over the generated HTTP
// client: a persistent browser-cookie jar, automatic bearer minting (and
// re-minting) via /api/auth/session, and detection of the API's masked
// "not really authenticated" responses (HTTP 200 with empty items, or 404
// conversation_inaccessible) that mean the bearer is missing or stale.
package gptapi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"

	"chatgpt-pp-cli/internal/client"
	"chatgpt-pp-cli/internal/cliutil"
	"chatgpt-pp-cli/internal/config"
)

// MaskedResponseError reports a response shape the ChatGPT API uses when the
// request carried cookies but no valid bearer: HTTP 200 with zero items, or
// 404 with code conversation_inaccessible. Callers treat it as "re-mint and
// retry once", never as "no data exists".
type MaskedResponseError struct {
	Path   string
	Status int
	Code   string
}

func (e *MaskedResponseError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("masked auth response for %s: HTTP %d (%s) — bearer missing or expired", e.Path, e.Status, e.Code)
	}
	return fmt.Sprintf("masked auth response for %s: HTTP 200 with empty result — bearer missing or expired", e.Path)
}

// Session holds minted bearer state. The token value never leaves this
// process except in the Authorization header.
type Session struct {
	mu          sync.Mutex
	accessToken string
	expires     time.Time
	mintedAt    time.Time
	source      string
	mintDead    bool // mint failed once (e.g. no cookie jar) — stop retrying per-call
}

// Client wraps the generated HTTP client with session behavior.
type Client struct {
	http    *client.Client
	cfg     *config.Config
	jarPath string
	session *Session
	limiter *cliutil.AdaptiveLimiter
}

// jarFile locations: alongside the config file, 0600.
func JarPath(cfg *config.Config) string {
	base, _ := cliutil.ConfigDir()
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config", "chatgpt-pp-cli")
	}
	return filepath.Join(base, "chatgpt-pp-cli", "cookies.jar")
}

// New builds a session-aware client.
func New(cfg *config.Config, c *client.Client) *Client {
	return &Client{
		http:    c,
		cfg:     cfg,
		jarPath: JarPath(cfg),
		session: &Session{},
		limiter: cliutil.NewAdaptiveLimiter(2.0), // start at 2 req/s, adapt
	}
}

// CookieJarExists reports whether a browser cookie jar has been imported.
func CookieJarExists() bool {
	_, err := os.Stat(JarPath(nil))
	return err == nil
}

// ImportCookieJar writes a Netscape-format cookie file to the private jar
// location. Values never appear in logs or output.
func ImportCookieJar(srcPath string) (int, error) {
	if fi, err := os.Stat(srcPath); err != nil || fi.IsDir() {
		return 0, fmt.Errorf("cookie file not readable: %s", srcPath)
	}
	// User-supplied cookie export path is read by design (their own session file).
	raw, err := os.ReadFile(srcPath) // #nosec G304 -- user-directed import of their own cookie export
	if err != nil {
		return 0, fmt.Errorf("reading cookie file: %w", err)
	}
	n := 0
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, "chatgpt.com") && !strings.Contains(line, "openai.com") {
			continue
		}
		n++
	}
	if n == 0 {
		return 0, errors.New("no chatgpt.com/openai.com cookies found in file")
	}
	dst := JarPath(nil)
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return n, err
	}
	// dst is a CLI-owned fixed path under the user config dir (no user input).
	if err := os.WriteFile(dst, raw, 0o600); err != nil { // #nosec G703 -- fixed CLI-owned path
		return n, err
	}
	return n, nil
}

// cookieHeader loads the jar and renders a Cookie header value (never logged).
func (c *Client) cookieHeader() string {
	raw, err := os.ReadFile(c.jarPath)
	if err != nil {
		return ""
	}
	var pairs []string
	for _, line := range strings.Split(string(raw), "\n") {
		f := strings.Split(strings.TrimSpace(line), "\t")
		if len(f) >= 7 && (strings.HasSuffix(f[0], "chatgpt.com")) {
			pairs = append(pairs, f[5]+"="+f[6])
		}
	}
	return strings.Join(pairs, "; ")
}

// bearer resolves the current bearer: minted token first, config/env token
// (CHATGPT_TOKEN) as fallback, optional ~/.codex/auth.json as second fallback.
func (c *Client) bearer(ctx context.Context) (string, error) {
	c.session.mu.Lock()
	tok := c.session.accessToken
	exp := c.session.expires
	c.session.mu.Unlock()
	if tok != "" && (exp.IsZero() || time.Now().Before(exp.Add(-time.Minute))) {
		return tok, nil
	}
	c.session.mu.Lock()
	dead := c.session.mintDead
	c.session.mu.Unlock()
	if !dead {
		if err := c.mint(ctx); err == nil {
			c.session.mu.Lock()
			tok = c.session.accessToken
			c.session.mu.Unlock()
			return tok, nil
		}
		// Do not retry the mint on every call once it has failed once
		// without a jar — fall straight to the configured token.
		c.session.mu.Lock()
		c.session.mintDead = true
		c.session.mu.Unlock()
	}
	// Fall back to configured token.
	if h := c.cfg.AuthHeader(); h != "" {
		return strings.TrimPrefix(h, "Bearer "), nil
	}
	if t := codexToken(); t != "" {
		return t, nil
	}
	return "", errors.New("no bearer available: run 'chatgpt-pp-cli auth import --cookies <file>' or set CHATGPT_TOKEN")
}

// codexToken reads ~/.codex/auth.json (tokens.access_token) if present.
// File contents are never logged.
func codexToken() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// Fixed well-known codex login location (no user-controlled component).
	raw, err := os.ReadFile(filepath.Join(home, ".codex", "auth.json")) // #nosec G304 -- fixed path
	if err != nil {
		return ""
	}
	var j struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if json.Unmarshal(raw, &j) != nil {
		return ""
	}
	if j.Tokens.AccessToken != "" {
		return j.Tokens.AccessToken
	}
	if j.AccessToken != "" {
		return j.AccessToken
	}
	return j.Token
}

// CodexAuthAvailable reports whether the codex login file exists.
func CodexAuthAvailable() bool { return codexToken() != "" }

// mint fetches /api/auth/session with the cookie jar and stores the token.
func (c *Client) mint(ctx context.Context) error {
	hdrs := map[string]string{}
	if ck := c.cookieHeader(); ck != "" {
		hdrs["Cookie"] = ck
	} else {
		return errors.New("no cookie jar imported; run 'chatgpt-pp-cli auth import --cookies <file>'")
	}
	raw, err := c.http.GetWithHeaders(ctx, "/api/auth/session", nil, hdrs)
	if err != nil {
		return fmt.Errorf("session mint: %w", err)
	}
	var j struct {
		AccessToken string `json:"accessToken"`
		Expires     string `json:"expires"`
	}
	if err := json.Unmarshal(raw, &j); err != nil {
		return fmt.Errorf("session mint parse: %w", err)
	}
	if j.AccessToken == "" {
		return errors.New("session mint returned no accessToken — cookies may be expired; re-import")
	}
	exp := jwtExpiry(j.AccessToken)
	c.session.mu.Lock()
	c.session.accessToken = j.AccessToken
	c.session.expires = exp
	c.session.mintedAt = time.Now()
	c.session.source = "cookie-mint"
	c.session.mu.Unlock()
	return nil
}

// jwtExpiry decodes the exp claim (numeric only). Token value stays local.
func jwtExpiry(tok string) time.Time {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return time.Time{}
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(payload, &claims) != nil || claims.Exp == 0 {
		return time.Time{}
	}
	return time.Unix(claims.Exp, 0)
}

// MintNow forces a fresh mint (used by `auth import`).
func (c *Client) MintNow(ctx context.Context) (time.Time, error) {
	if err := c.mint(ctx); err != nil {
		return time.Time{}, err
	}
	c.session.mu.Lock()
	defer c.session.mu.Unlock()
	return c.session.expires, nil
}

// SessionState describes the session without leaking the token.
type SessionState struct {
	HasCookieJar bool      `json:"has_cookie_jar"`
	CodexAuth    bool      `json:"codex_auth_available"`
	Minted       bool      `json:"minted"`
	BearerExpiry time.Time `json:"bearer_expiry,omitempty"`
	Source       string    `json:"source,omitempty"`
}

// State reports session status (no secrets).
func (c *Client) State() SessionState {
	c.session.mu.Lock()
	defer c.session.mu.Unlock()
	_, jarErr := os.Stat(c.jarPath)
	return SessionState{
		HasCookieJar: jarErr == nil,
		CodexAuth:    CodexAuthAvailable(),
		Minted:       c.session.accessToken != "",
		BearerExpiry: c.session.expires,
		Source:       c.session.source,
	}
}

func (c *Client) waitTurn(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

func (c *Client) noteResult(err error) {
	if err == nil {
		c.limiter.OnSuccess()
		return
	}
	var rl *cliutil.RateLimitError
	if errors.As(err, &rl) {
		c.limiter.OnRateLimit()
	}
}

// authedHeaders builds headers with bearer + cookies.
func (c *Client) authedHeaders(ctx context.Context, extra map[string]string) (map[string]string, error) {
	tok, err := c.bearer(ctx)
	if err != nil {
		return nil, err
	}
	h := map[string]string{"Authorization": "Bearer " + tok}
	if ck := c.cookieHeader(); ck != "" {
		h["Cookie"] = ck
	}
	for k, v := range extra {
		h[k] = v
	}
	return h, nil
}

// Get issues a session-authenticated GET and decodes JSON. Masked shapes
// return *MaskedResponseError after exactly one re-mint retry.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			c.session.mu.Lock()
			c.session.accessToken = "" // force re-mint
			c.session.mu.Unlock()
		}
		hdrs, err := c.authedHeaders(ctx, nil)
		if err != nil {
			return nil, err
		}
		if err := c.waitTurn(ctx); err != nil {
			return nil, err
		}
		raw, err := c.http.GetWithHeadersValues(ctx, path, params, hdrs)
		c.noteResult(err)
		if err != nil {
			var rl *cliutil.RateLimitError
			if errors.As(err, &rl) {
				return nil, err
			}
			if attempt == 0 && isAuthishError(err) {
				continue // re-mint and retry once
			}
			return nil, err
		}
		if masked, merr := detectMasked(path, raw, hdrs); masked {
			if attempt == 0 {
				continue
			}
			return nil, merr
		}
		return raw, nil
	}
	return nil, errors.New("unreachable")
}

// Post issues a session-authenticated POST with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any, extra map[string]string) (json.RawMessage, int, error) {
	hdrs, err := c.authedHeaders(ctx, extra)
	if err != nil {
		return nil, 0, err
	}
	hdrs["Content-Type"] = "application/json"
	if err := c.waitTurn(ctx); err != nil {
		return nil, 0, err
	}
	var contentType string
	if v, ok := hdrs["Content-Type"]; ok {
		contentType = v
	}
	_ = contentType
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	hdrs["Content-Type"] = "application/json"
	raw, status, err := c.http.PostWithHeaders(ctx, path, json.RawMessage(buf), hdrs)
	c.noteResult(err)
	return raw, status, err
}

// detectMasked inspects decoded bodies for the masked-auth shapes.
func detectMasked(path string, raw json.RawMessage, _ map[string]string) (bool, *MaskedResponseError) {
	s := string(raw)
	if strings.Contains(path, "/conversations") && !strings.Contains(path, "/search") {
		if strings.Contains(s, `"total":0`) && strings.Contains(s, `"items":[]`) && !strings.Contains(s, "accessToken") {
			return true, &MaskedResponseError{Path: path, Status: 200}
		}
		if strings.Contains(s, "conversation_inaccessible") {
			return true, &MaskedResponseError{Path: path, Status: 404, Code: "conversation_inaccessible"}
		}
	}
	return false, nil
}

func isAuthishError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "401") || strings.Contains(msg, "403") || strings.Contains(msg, "conversation_inaccessible")
}

// ---------- sentinel (chat-requirements + PoW) ----------

// Requirements is the chat-requirements response (token value never printed).
type Requirements struct {
	Persona     string `json:"persona"`
	Token       string `json:"token"`
	ExpireAfter int    `json:"expire_after"`
	ProofOfWork struct {
		Required   bool   `json:"required"`
		Seed       string `json:"seed"`
		Difficulty string `json:"difficulty"`
	} `json:"proofofwork"`
	Turnstile struct {
		Required bool `json:"required"`
	} `json:"turnstile"`
}

// fingerprint builds the `p` payload (base64 JSON array; format mirrors the
// public web client shape).
func fingerprint(ans string) string {
	cores := []int{8, 12, 16, 24}
	screens := []int{3000, 4000, 5000}
	var answer any
	if ans != "" {
		answer = ans
	}
	payload := []any{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36", cores, screens, answer, time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")}
	b, _ := json.Marshal(payload)
	return base64.StdEncoding.EncodeToString(b)
}

// SolvePoW finds a 16-hex-char answer whose sha3-512(seed+answer) hex prefix
// (difficulty-length) is <= difficulty. Difficulty ~0x06ab73 needs ~40 tries.
func SolvePoW(seed, difficulty string) (string, error) {
	if difficulty == "" {
		return "", errors.New("empty difficulty")
	}
	diffBig, success := new(big.Int).SetString(difficulty, 16)
	if !success || diffBig.Sign() <= 0 {
		return "", fmt.Errorf("bad difficulty %q", difficulty)
	}
	n := len(difficulty)
	const hexChars = "0123456789abcdef"
	for i := 0; i < 500000; i++ {
		cand := make([]byte, 16)
		if _, err := rand.Read(cand); err != nil {
			return "", fmt.Errorf("pow entropy: %w", err)
		}
		for j := range cand {
			cand[j] = hexChars[int(cand[j])%16]
		}
		h := sha3.New512()
		h.Write([]byte(seed))
		h.Write(cand)
		out := h.Sum(nil)
		hexStr := fmt.Sprintf("%x", out)
		val, ok := new(big.Int).SetString(hexStr[:n], 16)
		if ok && val.Cmp(diffBig) <= 0 {
			return string(cand), nil
		}
	}
	return "", errors.New("proof-of-work not solved within iteration budget")
}

// Requirements mints chat-requirements, solving PoW when required. Returns
// the requirements token plus the proof token ("" when PoW not required).
func (c *Client) Requirements(ctx context.Context) (req Requirements, proof string, err error) {
	hdrs := map[string]string{}
	// step 1: fingerprint, no answer
	if err := c.waitTurn(ctx); err != nil {
		return req, "", err
	}
	raw, status, err := c.Post(ctx, "/backend-api/sentinel/chat-requirements", map[string]string{"p": fingerprint("")}, hdrs)
	if err != nil || status != 200 {
		return req, "", fmt.Errorf("chat-requirements: status=%d err=%v", status, err)
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, "", err
	}
	if !req.ProofOfWork.Required {
		return req, "", nil
	}
	ans, err := SolvePoW(req.ProofOfWork.Seed, req.ProofOfWork.Difficulty)
	if err != nil {
		return req, "", err
	}
	if err := c.waitTurn(ctx); err != nil {
		return req, "", err
	}
	raw2, status2, err := c.Post(ctx, "/backend-api/sentinel/chat-requirements", map[string]string{"p": fingerprint(ans)}, hdrs)
	if err != nil || status2 != 200 {
		return req, "", fmt.Errorf("chat-requirements (pow): status=%d err=%v", status2, err)
	}
	if err := json.Unmarshal(raw2, &req); err != nil {
		return req, "", err
	}
	return req, fingerprint(ans), nil
}

// ConduitPrepare calls POST /f/conversation/prepare.
func (c *Client) ConduitPrepare(ctx context.Context) (string, error) {
	raw, status, err := c.Post(ctx, "/backend-api/f/conversation/prepare", map[string]string{}, nil)
	if err != nil || status != 200 {
		return "", fmt.Errorf("conduit prepare: status=%d err=%v", status, err)
	}
	var j struct {
		Status       string `json:"status"`
		ConduitToken string `json:"conduit_token"`
	}
	if err := json.Unmarshal(raw, &j); err != nil {
		return "", err
	}
	if j.ConduitToken == "" {
		return "", fmt.Errorf("conduit prepare returned no token (status=%s)", j.Status)
	}
	return j.ConduitToken, nil
}

// ---------- SSE ----------

// SSEEvent is one parsed data: line payload.
type SSEEvent struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

// StreamSend POSTs to /f/conversation and streams SSE events until the typed
// terminal (message_stream_complete) or ctx ends. It returns the parsed
// events, whether the terminal was observed, and the resume token when seen.
// Connection close or bare [DONE] without the typed terminal is reported as
// incomplete — never as success.
func (c *Client) StreamSend(ctx context.Context, body any, reqToken, proofToken, conduitToken string) (events []SSEEvent, terminal bool, resumeToken string, err error) {
	hdrs, err := c.authedHeaders(ctx, map[string]string{
		"Accept": "text/event-stream",
		"OpenAI-Sentinel-Chat-Requirements-Token": reqToken,
		"X-Conduit-Token":                         conduitToken,
		"OAI-Language":                            "en-US",
	})
	if err != nil {
		return nil, false, "", err
	}
	if proofToken != "" {
		hdrs["OpenAI-Sentinel-Proof-Token"] = proofToken
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, false, "", err
	}
	if err := c.waitTurn(ctx); err != nil {
		return nil, false, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.http.RequestBaseURL()+"/backend-api/f/conversation", bytes.NewReader(buf))
	if err != nil {
		return nil, false, "", err
	}
	for k, v := range hdrs {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	streamClient := client.StreamingHTTPClient(c.http.HTTPClient, 30*time.Second)
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, false, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, false, "", fmt.Errorf("send failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)
	events = make([]SSEEvent, 0)
	var lastType string
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			if strings.HasPrefix(line, "event:") {
				lastType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			}
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			// Advisory only: semantic completion is message_stream_complete.
			continue
		}
		var ev struct {
			Type  string `json:"type"`
			Token string `json:"token"`
		}
		_ = json.Unmarshal([]byte(data), &ev)
		t := ev.Type
		if t == "" {
			t = lastType
		}
		events = append(events, SSEEvent{Type: t, Raw: json.RawMessage(data)})
		if t == "resume_conversation_token" && ev.Token != "" {
			resumeToken = ev.Token
		}
		if t == "message_stream_complete" {
			terminal = true
		}
	}
	return events, terminal, resumeToken, sc.Err()
}
