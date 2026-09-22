//go:build darwin && !cgo

package auth

import "errors"

// The release build uses Security.framework with cgo. This compile-only
// fallback keeps cross-target verification deterministic when cgo is disabled;
// it never pretends that credentials were stored.
func platformKeychainRead() ([]byte, error) { return nil, ErrNotFound }

func platformKeychainWrite(_ []byte) error {
	return errors.New("native macOS Keychain backend unavailable when cgo is disabled")
}

func platformKeychainDelete() error { return nil }
