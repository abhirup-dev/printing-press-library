// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// LegacyMigrationState reports artifact metadata only. The Tickertape CLI is
// live-only and never adopts, copies, backs up, or opens local domain stores.
type LegacyMigrationState struct {
	Status    string                   `json:"status"`
	Preserved bool                     `json:"preserved"`
	Reason    string                   `json:"reason"`
	Config    *LegacyMigrationArtifact `json:"config,omitempty"`
	Database  *LegacyMigrationArtifact `json:"database,omitempty"`
}

type LegacyMigrationArtifact struct {
	Path        string `json:"path"`
	Status      string `json:"status"`
	Copied      bool   `json:"copied"`
	Destination string `json:"destination,omitempty"`
	BackupPath  string `json:"backup_path,omitempty"`
}

// InspectLegacyMigrationPaths performs filesystem metadata checks only. It
// intentionally does not read local config or database contents.
func InspectLegacyMigrationPaths(configPath, databasePath string) (LegacyMigrationState, error) {
	state := LegacyMigrationState{
		Status:    "disabled",
		Preserved: true,
		Reason:    "live-only CLI does not adopt local domain data",
	}
	for kind, raw := range map[string]string{"config": configPath, "database": databasePath} {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return LegacyMigrationState{}, fmt.Errorf("inspect legacy %s: %w", kind, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return LegacyMigrationState{}, fmt.Errorf("legacy %s must be a regular non-symlink file", kind)
		}
		artifact := &LegacyMigrationArtifact{Path: path, Status: "not_adopted", Copied: false}
		if kind == "config" {
			state.Config = artifact
		} else {
			state.Database = artifact
		}
	}
	return state, nil
}

func AdoptVerifiedLegacyDatabase(_ context.Context, _ *Session, _ *LegacyMigrationState) error {
	return errors.New("legacy database adoption is disabled: domain data is live-only")
}

func BackupLegacyArtifacts(_ *LegacyMigrationState) error {
	return errors.New("legacy artifact backup is disabled: this CLI does not manage local domain-data files")
}
