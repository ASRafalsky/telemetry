package main

import (
	"context"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/db/postgres"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

func initDB(ctx context.Context, cfg config.DB, l log.Logger) (postgres.DB, error) {
	if cfg.DSN == "" {
		return postgres.DB{}, errors.New("database not configured")
	}
	db, err := postgres.Open(cfg.DSN)
	if err == nil {
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(5 * time.Minute)
		l.Info("Opened database with dsn", "dsn", cfg.DSN)
		if err = db.WaitDBIsReady(ctx, 1, time.Second); err != nil {
			l.Error("failed to ping to database: ", err.Error())
		}
		if err = db.MigrateUp(cfg.MigrationsPath); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				l.Info("Nothing to migrate")
			} else {
				l.Error("failed to migrate database: ", err.Error())
				if err = db.MigrateRollback(cfg.MigrationsPath, 1); err != nil {
					l.Error("failed to rollback database: ", err.Error())
				}
			}
		}
	} else {
		return postgres.DB{}, errors.New("failed to connect to database")
	}
	return db, nil
}
