// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

// Package auth owns the CLI's credential boundary. It deliberately stores only
// authentication material in the platform keychain; research responses never
// pass through this package or get persisted locally.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	Service = "tickertape-pp-cli"
	Account = "default"
)

var ErrNotFound = errors.New("no Tickertape credentials in keychain")

// Credentials is intentionally small and JSON-compatible so browser bootstrap
// can pipe it to `auth import --stdin` without putting secrets in argv.
type Credentials struct {
	AuthHeader string            `json:"auth_header,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

func (c Credentials) Empty() bool {
	return strings.TrimSpace(c.AuthHeader) == "" && len(c.Headers) == 0
}

func (c Credentials) HasCookie() bool {
	for k, v := range c.Headers {
		if strings.EqualFold(k, "Cookie") && strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

func (c Credentials) HasAuthHeader() bool { return strings.TrimSpace(c.AuthHeader) != "" }

func Parse(raw []byte) (Credentials, error) {
	var c Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		return Credentials{}, fmt.Errorf("credentials must be a JSON object with auth_header and/or headers: %w", err)
	}
	if err := Validate(c); err != nil {
		return Credentials{}, err
	}
	return c, nil
}

func Validate(c Credentials) error {
	if c.Empty() {
		return errors.New("credentials are empty")
	}
	if strings.ContainsAny(c.AuthHeader, "\r\n") {
		return errors.New("auth_header contains an invalid newline")
	}
	for k, v := range c.Headers {
		key := strings.TrimSpace(k)
		if key == "" || strings.ContainsAny(key, "\r\n") || strings.ContainsAny(v, "\r\n") {
			return fmt.Errorf("header %q contains an invalid newline or empty name", k)
		}
		// These are request-framing headers and must not be imported from a
		// browser dump. Authorization/Cookie and provider-specific headers are
		// allowed; the HTTP client owns Host and Content-Length.
		switch strings.ToLower(key) {
		case "host", "content-length", "transfer-encoding", "connection":
			return fmt.Errorf("header %q is not permitted", key)
		}
	}
	return nil
}

func ReadKeychain(ctx context.Context) (Credentials, error) {
	_ = ctx
	out, err := platformKeychainRead()
	if err != nil {
		return Credentials{}, err
	}
	c, err := Parse(out)
	if err != nil {
		return Credentials{}, fmt.Errorf("invalid Tickertape keychain entry: %w", err)
	}
	return c, nil
}

func WriteKeychain(ctx context.Context, c Credentials) error {
	_ = ctx
	if err := Validate(c); err != nil {
		return err
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}
	return platformKeychainWrite(payload)
}

func DeleteKeychain(ctx context.Context) error {
	_ = ctx
	return platformKeychainDelete()
}

// EnvironmentCredentials is the explicit non-persistent fallback for CI and
// controlled shells. It never writes the values anywhere.
func EnvironmentCredentials() (Credentials, string, error) {
	c := Credentials{Headers: map[string]string{}}
	if v := strings.TrimSpace(os.Getenv("TICKERTAPE_AUTH_HEADER")); v != "" {
		c.AuthHeader = v
	}
	if v := strings.TrimSpace(os.Getenv("TICKERTAPE_COOKIE")); v != "" {
		c.Headers["Cookie"] = v
	}
	if v := strings.TrimSpace(os.Getenv("TICKERTAPE_USER_AGENT")); v != "" {
		c.Headers["User-Agent"] = v
	}
	if len(c.Headers) == 0 {
		c.Headers = nil
	}
	if c.Empty() {
		return Credentials{}, "", ErrNotFound
	}
	if err := Validate(c); err != nil {
		return Credentials{}, "", err
	}
	return c, "environment", nil
}
