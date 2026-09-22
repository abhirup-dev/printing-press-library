package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"moneycontrol-pp-cli/internal/cliutil"
)

const (
	keychainService  = "moneycontrol-pp-cli/auth-session"
	keychainAccount  = "moneycontrol-pp-cli"
	fallbackAuthName = "auth-session.json"
)

var ErrAuthNotFound = errors.New("no stored Moneycontrol authentication material")

// SessionMaterial is retained only in memory while constructing the client.
// Values must never appear in status output, receipts, errors, or argv.
type SessionMaterial struct {
	Cookie  string
	Headers map[string]string
	Domain  string
}

type storedSessionMaterial struct {
	Cookie  string            `json:"cookie"`
	Headers map[string]string `json:"headers,omitempty"`
	Domain  string            `json:"domain,omitempty"`
}

// AuthStoreStatus is intentionally metadata-only.
type AuthStoreStatus struct {
	Backend string `json:"backend"`
	Present bool   `json:"present"`
	Path    string `json:"path,omitempty"`
}

func ReadStoredAuthMaterial() (SessionMaterial, string, error) {
	if value := strings.TrimSpace(os.Getenv("MONEYCONTROL_AUTH_COOKIE")); value != "" {
		return SessionMaterial{Cookie: value, Domain: "moneycontrol.com"}, "environment", nil
	}
	var raw, backend string
	if runtime.GOOS == "darwin" {
		if value, err := readKeychainAuth(); err == nil && strings.TrimSpace(value) != "" {
			raw, backend = value, "keychain"
		}
	}
	if raw == "" {
		path, err := fallbackAuthPath()
		if err != nil {
			return SessionMaterial{}, "", err
		}
		if info, err := os.Stat(path); err == nil && info.Mode().Perm()&0077 != 0 {
			return SessionMaterial{}, "", fmt.Errorf("stored auth fallback %q has unsafe permissions", path)
		}
		data, err := os.ReadFile(path) // #nosec G304 -- app-derived private auth path.
		if errors.Is(err, os.ErrNotExist) {
			return SessionMaterial{}, "", ErrAuthNotFound
		}
		if err != nil {
			return SessionMaterial{}, "", fmt.Errorf("read stored auth fallback: %w", err)
		}
		raw, backend = string(data), "file"
	}
	material, err := decodeSessionMaterial(raw)
	if err != nil {
		return SessionMaterial{}, "", err
	}
	return material, backend, nil
}

func StoreAuthMaterial(material SessionMaterial) (string, error) {
	if err := validateSessionMaterial(material); err != nil {
		return "", err
	}
	payload, err := json.Marshal(storedSessionMaterial{Cookie: material.Cookie, Headers: material.Headers, Domain: material.Domain})
	if err != nil {
		return "", fmt.Errorf("encode auth material: %w", err)
	}
	if runtime.GOOS == "darwin" {
		if err := writeKeychainAuth(string(payload)); err == nil {
			return "keychain", nil
		}
	}
	path, err := fallbackAuthPath()
	if err != nil {
		return "", err
	}
	if err := cliutil.AtomicWritePrivateFile(path, append(payload, '\n'), 0600, 0700); err != nil {
		return "", err
	}
	return "file", nil
}

func ClearStoredAuthMaterial() (string, error) {
	var keychainErr error
	if runtime.GOOS == "darwin" {
		keychainErr = deleteKeychainAuth()
	}
	path, err := fallbackAuthPath()
	if err != nil {
		return "", err
	}
	fileErr := os.Remove(path)
	if errors.Is(fileErr, os.ErrNotExist) {
		fileErr = nil
	}
	if keychainErr != nil && fileErr != nil {
		return "", fmt.Errorf("clear keychain and file auth: %v; %w", keychainErr, fileErr)
	}
	if keychainErr == nil && runtime.GOOS == "darwin" {
		return "keychain", nil
	}
	return "file", fileErr
}

func AuthStoreMetadata() AuthStoreStatus {
	if _, backend, err := ReadStoredAuthMaterial(); err == nil {
		status := AuthStoreStatus{Backend: backend, Present: true}
		if backend == "file" {
			if path, pathErr := fallbackAuthPath(); pathErr == nil {
				status.Path = path
			}
		}
		return status
	}
	return AuthStoreStatus{Backend: "none", Present: false}
}

func decodeSessionMaterial(raw string) (SessionMaterial, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return SessionMaterial{}, ErrAuthNotFound
	}
	var stored storedSessionMaterial
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return SessionMaterial{}, errors.New("stored authentication material is not valid JSON")
	}
	material := SessionMaterial{Cookie: strings.TrimSpace(stored.Cookie), Headers: stored.Headers, Domain: strings.TrimSpace(stored.Domain)}
	if material.Domain == "" {
		material.Domain = "moneycontrol.com"
	}
	if err := validateSessionMaterial(material); err != nil {
		return SessionMaterial{}, err
	}
	return material, nil
}

func validateSessionMaterial(material SessionMaterial) error {
	if strings.TrimSpace(material.Cookie) == "" {
		return errors.New("authentication material must include a Cookie session")
	}
	if strings.ContainsAny(material.Cookie, "\r\n") {
		return errors.New("Cookie session must be one line")
	}
	for name, value := range material.Headers {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return errors.New("authentication request headers must have valid names and one-line values")
		}
	}
	if !strings.EqualFold(strings.TrimPrefix(strings.TrimSpace(material.Domain), "."), "moneycontrol.com") {
		return errors.New("authentication material domain must be moneycontrol.com")
	}
	return nil
}

func fallbackAuthPath() (string, error) {
	dir, err := cliutil.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fallbackAuthName), nil
}

func readKeychainAuth() (string, error) {
	cmd := exec.Command("/usr/bin/security", "find-generic-password", "-a", keychainAccount, "-s", keychainService, "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func writeKeychainAuth(value string) error {
	cmd := exec.Command("/usr/bin/security", "add-generic-password", "-U", "-a", keychainAccount, "-s", keychainService, "-w")
	cmd.Stdin = strings.NewReader(value + "\n")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return err
	}
	// Some security CLI versions report success while storing an empty
	// password when -w is fed from a non-terminal. Verify without exposing the
	// value; StoreAuthMaterial will fall back to the owner-only file if this
	// backend did not retain the payload.
	stored, err := readKeychainAuth()
	if err != nil || stored != value {
		return errors.New("keychain write verification failed")
	}
	return nil
}

func deleteKeychainAuth() error {
	cmd := exec.Command("/usr/bin/security", "delete-generic-password", "-a", keychainAccount, "-s", keychainService)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 44 {
		return nil
	}
	return err
}
