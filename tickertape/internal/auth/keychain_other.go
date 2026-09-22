//go:build !darwin

package auth

import "errors"

func platformKeychainRead() ([]byte, error) { return nil, ErrNotFound }

func platformKeychainWrite(_ []byte) error {
	return errors.New("keychain auth is supported only on macOS")
}

func platformKeychainDelete() error { return nil }
