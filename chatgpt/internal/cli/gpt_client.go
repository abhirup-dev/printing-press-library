// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"chatgpt-pp-cli/internal/config"
	"chatgpt-pp-cli/internal/gptapi"
)

// newGptClient builds the session-aware wrapper used by hand-built commands.
func newGptClient(flags *rootFlags) (*gptapi.Client, error) {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return nil, configErr(err)
	}
	c, err := flags.newClient()
	if err != nil {
		return nil, err
	}
	return gptapi.New(cfg, c), nil
}
