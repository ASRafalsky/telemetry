package repository

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/internal/backup"
	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

func (r *ExtendedRepository) Maintain(ctx context.Context, cfg config.Server, l log.Logger) {
	if cfg.Restore {
		if err := backup.RestoreRepo(cfg.DumpPath, r); err != nil {
			l.Error("Failed to restore from the dump file:", cfg.DumpPath, err.Error())
		}
	}

	switch {
	case r.db != nil && cfg.DB.DSN != "":
		go func() {
			r.syncer(ctx, l)

		}()
	case cfg.DumpPath != config.DefaultDumpPath && cfg.StorePeriod != config.DefaultDumpInterval:
		go backup.BackupRepo(ctx, r, cfg.StorePeriod, cfg.DumpPath, l)
	}
}

func (r *ExtendedRepository) syncer(ctx context.Context, l log.Logger) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer func() {
		l.Info("Syncing repository stopped")
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			syncOnStopCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			if err := r.Sync(syncOnStopCtx); err != nil {
				l.Error("Failed to sync data with db on the stop", err.Error())
			}
			cancel()
			return
		case <-ticker.C:
			if !r.IsReady() {
				continue
			}
			if err := r.Sync(ctx); err != nil {
				l.Error("Failed to sync data with db", err.Error())
			}
		}
	}
}
