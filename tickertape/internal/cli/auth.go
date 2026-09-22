package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tickertape-pp-cli/internal/auth"
	"tickertape-pp-cli/internal/client"
	"tickertape-pp-cli/internal/config"
)

const (
	tickertapeAuthUserEndpoint = "https://auth.api.tickertape.in/auth/user/v3"
	tickertapeCreditEndpoint   = "https://auth.api.tickertape.in/user/credit/combined/v4"
	tickertapePortfolioStatus  = "https://ecosystem.api.tickertape.in/portfolio/v4/holdings/status"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage secure Tickertape authentication",
	}

	var stdin bool
	importCmd := &cobra.Command{
		Use:   "import --stdin",
		Short: "Import auth JSON from stdin into the macOS Keychain",
		Example: `  cat credentials.json | tickertape-pp-cli auth import --stdin
  ego-browser nodejs <<'EOF'
  // pipe {auth_header,headers} to auth import --stdin without printing it
  EOF`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !stdin {
				return usageErr(errors.New("auth import requires --stdin; credentials are never accepted as arguments"))
			}
			raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 64*1024))
			if err != nil {
				return fmt.Errorf("read credentials from stdin: %w", err)
			}
			value := strings.TrimSpace(string(raw))
			if value == "" {
				return usageErr(errors.New("stdin must contain session-material JSON or one auth header"))
			}
			var creds auth.Credentials
			if strings.HasPrefix(value, "{") {
				creds, err = auth.Parse([]byte(value))
			} else {
				if strings.ContainsAny(value, "\r\n") {
					return usageErr(errors.New("stdin auth header must be exactly one line"))
				}
				creds = auth.Credentials{AuthHeader: value}
				err = auth.Validate(creds)
			}
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()
			if err := auth.WriteKeychain(ctx, creds); err != nil {
				return err
			}
			return flags.printJSON(cmd, map[string]any{
				"stored":                     true,
				"store":                      "macos_keychain",
				"credential_type":            credentialType(creds),
				"credential_values_included": false,
			})
		},
	}
	importCmd.Flags().BoolVar(&stdin, "stdin", false, "Read credential JSON from stdin (required)")

	statusCmd := &cobra.Command{
		Use:     "status",
		Short:   "Probe the live authenticated identity and entitlement state",
		Example: `  tickertape-pp-cli auth status --agent`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthStatus(cmd, flags)
		},
	}

	logoutCmd := &cobra.Command{
		Use:     "logout",
		Short:   "Remove Tickertape credentials from the macOS Keychain",
		Example: `  tickertape-pp-cli auth logout --agent`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()
			if err := auth.DeleteKeychain(ctx); err != nil {
				return err
			}
			return flags.printJSON(cmd, map[string]any{"logged_out": true, "store": "macos_keychain"})
		},
	}

	cmd.AddCommand(importCmd, statusCmd, logoutCmd)
	return cmd
}

func credentialType(c auth.Credentials) string {
	switch {
	case c.HasAuthHeader() && c.HasCookie():
		return "authorization_and_cookie"
	case c.HasAuthHeader():
		return "authorization"
	case c.HasCookie():
		return "cookie"
	default:
		return "headers"
	}
}

func runAuthStatus(cmd *cobra.Command, flags *rootFlags) error {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return configErr(err)
	}
	result := map[string]any{
		"configured":       cfg.CredentialConfigured(),
		"status":           "not_configured",
		"source":           cfg.AuthSource,
		"credential_types": credentialTypesFromConfig(cfg),
		"observed_at":      time.Now().UTC().Format(time.RFC3339Nano),
		"probes":           map[string]any{},
	}
	if !cfg.CredentialConfigured() {
		result["probes"] = map[string]any{"identity": map[string]any{"state": "not_run", "reason": "no credentials configured"}}
		return flags.printJSON(cmd, result)
	}

	c := client.New(cfg, flags.timeout, flags.rateLimit)
	identity := sanitizedAuthProbe(cmd.Context(), c, tickertapeAuthUserEndpoint, nil)
	credit := sanitizedAuthProbe(cmd.Context(), c, tickertapeCreditEndpoint, map[string]string{"products": "LAS"})
	result["probes"] = map[string]any{"identity": identity, "entitlement": credit}
	result["status"] = classifyAuthStatus(identity, credit)
	return flags.printJSON(cmd, result)
}

func credentialTypesFromConfig(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	var out []string
	if cfg.AuthHeader() != "" {
		out = append(out, "authorization")
	}
	for key, value := range cfg.Headers {
		if strings.EqualFold(key, "Cookie") && strings.TrimSpace(value) != "" {
			out = append(out, "cookie")
		}
	}
	if len(out) == 0 {
		return []string{}
	}
	return out
}

func sanitizedAuthProbe(ctx context.Context, c *client.Client, endpoint string, params map[string]string) map[string]any {
	started := time.Now().UTC().Format(time.RFC3339Nano)
	data, err := c.GetWithHeaders(ctx, endpoint, params, nil)
	out := map[string]any{"endpoint": endpoint, "observed_at": started}
	if err != nil {
		out["state"] = "error"
		if apiErr := new(client.APIError); errors.As(err, &apiErr) {
			out["http_status"] = apiErr.StatusCode
			if apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden {
				out["access_state"] = "unauthorized"
			} else {
				out["access_state"] = "provider_error"
			}
		} else {
			out["access_state"] = "transport_error"
		}
		out["error"] = sanitizeAuthError(err)
		return out
	}
	out["state"] = "ok"
	out["http_status"] = http.StatusOK
	out["identity"] = sanitizeIdentity(data)
	out["entitlements"] = sanitizeEntitlementKeys(data)
	return out
}

func classifyAuthStatus(identity, entitlement map[string]any) string {
	identityState, _ := identity["access_state"].(string)
	identityHTTP, _ := identity["http_status"].(int)
	if identityState == "unauthorized" || identityHTTP == http.StatusUnauthorized {
		return "expired"
	}
	if identityState == "transport_error" || identityState == "provider_error" {
		return "upstream_error"
	}
	if identityState != "" {
		// A non-empty access state on the identity probe is an explicit
		// failure/access classification; keep it conservative.
		if identityState == "blocked" {
			return "blocked"
		}
		return "shape_changed"
	}
	if entitlementState, _ := entitlement["access_state"].(string); entitlementState == "unauthorized" {
		return "entitlement_denied"
	}
	if entitlementState, _ := entitlement["access_state"].(string); entitlementState == "transport_error" || entitlementState == "provider_error" {
		return "upstream_error"
	}
	if _, ok := identity["identity"]; !ok {
		return "shape_changed"
	}
	return "authenticated"
}

func sanitizeAuthError(err error) string {
	if err == nil {
		return ""
	}
	// Error strings from the HTTP client contain method/path/status but not
	// request headers. Keep the message bounded and never include response bodies
	// from an auth endpoint, which can contain account-specific fields.
	msg := err.Error()
	if i := strings.Index(msg, ": "); i >= 0 {
		msg = msg[:i]
	}
	if len(msg) > 240 {
		msg = msg[:240]
	}
	return msg
}

func sanitizeIdentity(data []byte) map[string]any {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return map[string]any{"present": false}
	}
	found := map[string]any{"present": true, "user_id_present": false, "name_present": false, "login_state": "unknown"}
	walkSanitized(value, func(key string, val any) {
		n := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", ""), "_", ""))
		switch n {
		case "id", "userid", "useridvalue", "profileid", "_id":
			found["user_id_present"] = true
		case "name", "displayname", "username", "handle", "email":
			found["name_present"] = true
		case "authenticated", "isloggedin", "loggedin":
			if b, ok := val.(bool); ok && b {
				found["login_state"] = "authenticated"
			}
		case "status":
			if s, ok := val.(string); ok && s != "" {
				found["login_state"] = "present"
			}
		}
	})
	return found
}

func sanitizeEntitlementKeys(data []byte) []string {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return []string{}
	}
	seen := map[string]bool{}
	walkSanitized(value, func(key string, _ any) {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "entitlement") || strings.Contains(lower, "membership") || strings.Contains(lower, "subscription") || strings.Contains(lower, "premium") || strings.Contains(lower, "credit") || strings.Contains(lower, "plan") {
			seen[key] = true
		}
	})
	out := make([]string, 0, len(seen))
	for key := range seen {
		out = append(out, key)
	}
	// Stable output helps agents compare status calls without persisting data.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func walkSanitized(value any, visit func(string, any)) {
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			visit(key, child)
			walkSanitized(child, visit)
		}
	case []any:
		for _, child := range node {
			walkSanitized(child, visit)
		}
	}
}

func newUserCmd(flags *rootFlags) *cobra.Command {
	group := &cobra.Command{Use: "user", Short: "Read-only authenticated user state"}
	group.AddCommand(newAuthenticatedResearchLeaf(flags, "status", "Live authenticated user status", tickertapeAuthUserEndpoint, nil, "user_status"))
	return group
}

func newCreditCmd(flags *rootFlags) *cobra.Command {
	group := &cobra.Command{Use: "credit", Short: "Read-only authenticated entitlement state"}
	group.AddCommand(newAuthenticatedResearchLeaf(flags, "summary", "Live credit and entitlement summary", tickertapeCreditEndpoint, map[string]string{"products": "LAS"}, "credit_summary"))
	return group
}

func newPortfolioCmd(flags *rootFlags) *cobra.Command {
	group := &cobra.Command{Use: "portfolio", Short: "Read-only authenticated portfolio status"}
	group.AddCommand(newAuthenticatedResearchLeaf(flags, "status", "Live holdings synchronization status", tickertapePortfolioStatus, nil, "portfolio_status"))
	return group
}

func newAuthenticatedResearchLeaf(flags *rootFlags, name, short, endpoint string, params map[string]string, resource string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel, c, err := newResearchRequest(cmd, flags)
			if err != nil {
				return err
			}
			defer cancel()
			result := fetchResearchEndpoint(ctx, c, resource, endpoint, params)
			return printResearchResult(cmd, flags, resource, result)
		},
	}
}
