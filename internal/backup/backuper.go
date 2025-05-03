package backup

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/pkg/log"
)

func BackupRepo(
	ctx context.Context, data dataDumper, interval int, path string, l log.Logger,
) {
	timeInt := 500 * time.Millisecond
	if interval > 0 {
		timeInt = time.Duration(interval) * time.Second
	}
	l.Info("Backuping repository started with interval:", timeInt.String(), "path:", path)
	ticker := time.NewTicker(timeInt)
	defer func() {
		l.Info("Backuping repository stopped, data stored to", path)
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			if err := DumpRepoToFile(path, data, 0o644); err != nil {
				l.Error("Failed to dump data to file", path, err.Error())
			}
			return
		case <-ticker.C:
			if !data.IsReady() {
				continue
			}
			if err := DumpRepoToFile(path, data, 0o644); err == nil {
				data.Ready()
			} else {
				l.Error("Failed to dump data to file", path, err.Error())
			}
		}
	}
}

func RestoreRepo(path string, repo dataRestorer) error {
	return RestoreRepoFromFile(path, repo, false)
}

type dataRestorer interface {
	Set(k string, v []byte)
	Size() int
}

type dataDumper interface {
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	IsReady() bool
	Ready()
}
