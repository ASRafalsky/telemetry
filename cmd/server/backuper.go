package main

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/pkg/services/backup"
)

func backupRepo(
	ctx context.Context, data *extendedRepository, interval int, path string, l log.Logger,
) {
	timeInt := 500 * time.Millisecond
	if interval > 0 {
		timeInt = time.Duration(interval) * time.Second
	}
	l.Info("Backuping repository started with interval: "+timeInt.String(), "path:", path)
	ticker := time.NewTicker(timeInt)
	defer func() {
		l.Info("Backuping repository stopped, data stored to", path)
		ticker.Stop()
	}()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			if err := backup.DumpRepoToFile(path, data, 0o644); err != nil {
				l.Error("Failed to dump data to file", path, err.Error())
			}
			return
		case <-ticker.C:
			if !data.newData.Load() {
				continue
			}
			if err := backup.DumpRepoToFile(path, data, 0o644); err == nil {
				data.newData.Store(false)
			} else {
				l.Error("Failed to dump data to file", path, err.Error())
			}
		}
	}
}

func restoreRepo(path string, repo *extendedRepository) error {
	return backup.RestoreRepoFromFile(path, repo, false)
}
