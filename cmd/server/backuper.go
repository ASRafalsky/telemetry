package main

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/pkg/services/backup"
)

func backupRepo(
	ctx context.Context, data repository, interval int, path string, l log.Logger,
) {
	var timer *time.Timer
	if interval > 0 {
		timeInt := time.Duration(interval) * time.Second
		l.Info("Backuping repository started with interval: "+timeInt.String(), "path:", path)
		timer = time.NewTimer(timeInt)
		defer timer.Stop()
	}

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			if err := backup.DumpRepoToFile(path, data, 0o644); err != nil {
				l.Error("Failed to dump data to file", path, err.Error())
			}
			return
		case <-timer.C:
			if err := backup.DumpRepoToFile(path, data, 0o644); err != nil {
				l.Error("Failed to dump data to file", path, err.Error())
			}
		}
	}
}

func restoreRepo(path string, repo repository) error {
	return backup.RestoreRepoFromFile(path, repo, false)
}
