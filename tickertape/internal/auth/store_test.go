package auth

import (
	"context"
	"errors"
	"runtime"
	"testing"
)

func TestParseCookieCredentials(t *testing.T) {
	creds, err := Parse([]byte(`{"headers":{"Cookie":"jwt=redacted","User-Agent":"test"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !creds.HasCookie() || creds.HasAuthHeader() {
		t.Fatalf("unexpected credential classification: %#v", creds)
	}
}

func TestValidateRejectsFramingHeaders(t *testing.T) {
	if err := Validate(Credentials{Headers: map[string]string{"Host": "evil.example"}}); err == nil {
		t.Fatal("expected Host to be rejected")
	}
	if err := Validate(Credentials{AuthHeader: "Bearer bad\nvalue"}); err == nil {
		t.Fatal("expected newline to be rejected")
	}
}

func TestEnvironmentCredentials(t *testing.T) {
	t.Setenv("TICKERTAPE_AUTH_HEADER", "Bearer test-token")
	t.Setenv("TICKERTAPE_COOKIE", "jwt=test-cookie")
	t.Setenv("TICKERTAPE_USER_AGENT", "test-agent")
	creds, source, err := EnvironmentCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if source != "environment" || creds.AuthHeader != "Bearer test-token" || creds.Headers["Cookie"] != "jwt=test-cookie" {
		t.Fatalf("unexpected environment credentials: source=%q creds=%#v", source, creds)
	}
}

func TestEnvironmentCredentialsAbsent(t *testing.T) {
	t.Setenv("TICKERTAPE_AUTH_HEADER", "")
	t.Setenv("TICKERTAPE_COOKIE", "")
	t.Setenv("TICKERTAPE_USER_AGENT", "")
	_, _, err := EnvironmentCredentials()
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReadKeychainNonDarwinDoesNotShellOut(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("platform-specific keychain behavior")
	}
	_, err := ReadKeychain(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
