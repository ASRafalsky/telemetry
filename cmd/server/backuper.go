package main

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/pkg/services/backup"
)

func backupRepo(
	ctx context.Context, data map[string]backup.Repository, interval time.Duration, path string, l log.Logger,
) {
	l.Info("Backuping repository started with interval: "+interval.String(), "path:", path)
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := backup.DumpRepoToFile(path, data, 0o644); err != nil {
				l.Error("Failed to dump data to file", path, err.Error())
			}
		}
	}
}

func restoreRepo(path string, repos map[string]backup.Repository) error {
	return backup.RestoreRepoFromFile(path, repos, false)
}
