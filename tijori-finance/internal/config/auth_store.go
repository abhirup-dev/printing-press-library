package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tijori-finance-pp-cli/internal/cliutil"
)

const (
	keychainService  = "tijori-finance-pp-cli/auth-header"
	keychainAccount  = "tijori-finance-pp-cli"
	fallbackAuthName = "auth-header"
)

var ErrAuthNotFound = errors.New("no stored Tijori authentication material")

// SessionMaterial is held only in memory while constructing a client. Values
// are never included in status reports, receipts, errors, or command output.
type SessionMaterial struct {
	AuthHeader string
	Headers    map[string]string
}

type storedSessionMaterial struct {
	AuthHeader    string            `json:"auth_header,omitempty"`
	Authorization string            `json:"authorization,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

// AuthStoreStatus is deliberately metadata-only: it never contains a
// credential value or a fingerprint that could be used to recover it.
type AuthStoreStatus struct {
	Backend string `json:"backend"`
	Present bool   `json:"present"`
	Path    string `json:"path,omitempty"`
}

func ReadStoredAuthMaterial() (SessionMaterial, string, error) {
	var raw string
	backend := ""
	if value, keychainErr := keychainRead(); keychainErr == nil && value != "" {
		raw, backend = value, "keychain"
	}
	if raw == "" {
		path, pathErr := fallbackAuthPath()
		if pathErr != nil {
			return SessionMaterial{}, "", pathErr
		}
		info, statErr := os.Stat(path)
		if statErr == nil && info.Mode().Perm()&0077 != 0 {
			return SessionMaterial{}, "", fmt.Errorf("stored auth fallback %q has unsafe permissions", path)
		}
		data, readErr := os.ReadFile(path) // #nosec G304 -- app-derived private auth path.
		if errors.Is(readErr, os.ErrNotExist) {
			return SessionMaterial{}, "", ErrAuthNotFound
		}
		if readErr != nil {
			return SessionMaterial{}, "", fmt.Errorf("read stored auth fallback: %w", readErr)
		}
		raw, backend = strings.TrimSpace(string(data)), "file"
	}
	material, err := decodeSessionMaterial(raw)
	if err != nil {
		return SessionMaterial{}, "", err
	}
	return material, backend, nil
}

func ReadStoredAuthHeader() (value, backend string, err error) {
	material, backend, err := ReadStoredAuthMaterial()
	return material.AuthHeader, backend, err
}

func StoreAuthHeader(value string) (backend string, err error) {
	return StoreAuthMaterial(SessionMaterial{AuthHeader: value})
}

func StoreAuthMaterial(material SessionMaterial) (backend string, err error) {
	if err := validateSessionMaterial(material); err != nil {
		return "", err
	}
	payload, err := json.Marshal(storedSessionMaterial{AuthHeader: material.AuthHeader, Headers: material.Headers})
	if err != nil {
		return "", fmt.Errorf("encode auth material: %w", err)
	}
	if err := keychainWrite(string(payload)); err == nil {
		return "keychain", nil
	}
	path, pathErr := fallbackAuthPath()
	if pathErr != nil {
		return "", pathErr
	}
	if err := cliutil.AtomicWritePrivateFile(path, append(payload, '\n'), 0600, 0700); err != nil {
		return "", err
	}
	return "file", nil
}

func ClearStoredAuthHeader() (backend string, err error) {
	var keychainErr error
	keychainErr = keychainDelete()
	path, pathErr := fallbackAuthPath()
	if pathErr != nil {
		return "", pathErr
	}
	fileErr := os.Remove(path)
	if errors.Is(fileErr, os.ErrNotExist) {
		fileErr = nil
	}
	if keychainErr != nil && fileErr != nil {
		return "", fmt.Errorf("clear keychain and file auth: %v; %w", keychainErr, fileErr)
	}
	if keychainErr == nil {
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
	if !strings.HasPrefix(raw, "{") {
		if strings.ContainsAny(raw, "\r\n") {
			return SessionMaterial{}, errors.New("stored authentication header contains multiple lines")
		}
		return SessionMaterial{AuthHeader: raw}, nil
	}
	var stored storedSessionMaterial
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return SessionMaterial{}, errors.New("stored authentication material is not valid JSON")
	}
	if stored.AuthHeader == "" {
		stored.AuthHeader = stored.Authorization
	}
	material := SessionMaterial{AuthHeader: strings.TrimSpace(stored.AuthHeader), Headers: stored.Headers}
	if err := validateSessionMaterial(material); err != nil {
		return SessionMaterial{}, err
	}
	return material, nil
}

func validateSessionMaterial(material SessionMaterial) error {
	if strings.TrimSpace(material.AuthHeader) == "" && len(material.Headers) == 0 {
		return errors.New("authentication material must include an auth header or request headers")
	}
	if strings.ContainsAny(material.AuthHeader, "\r\n") {
		return errors.New("authentication header must be one non-empty line")
	}
	for name, value := range material.Headers {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return errors.New("authentication request headers must have valid names and one-line values")
		}
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
