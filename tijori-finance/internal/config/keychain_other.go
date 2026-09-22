//go:build !darwin || !cgo

package config

import "errors"

var errKeychainUnavailable = errors.New("macOS Keychain unavailable")

func keychainRead() (string, error) { return "", errKeychainUnavailable }
func keychainWrite(string) error    { return errKeychainUnavailable }
func keychainDelete() error         { return errKeychainUnavailable }
