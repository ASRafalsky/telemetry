package main

import (
	"context"
	"errors"
	"time"

	"github.com/ASRafalsky/telemetry/internal/db/postgres"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

func initDB(ctx context.Context, dsn string, l log.Logger) (postgres.DB, error) {
	if dsn == "" {
		return postgres.DB{}, errors.New("database not configured")
	}
	db, err := postgres.Open(dsn)
	if err == nil {
		l.Info("Opened database with dsn", "dsn", dsn)
		if err = db.WaitDBIsReady(ctx, 1, time.Second); err != nil {
			l.Error("failed to ping to database: ", err.Error())
		}
		if err = db.Bootstrap(ctx); err != nil {
			l.Error("failed to bootstrap database: ", err.Error())
		}
	} else {
		return postgres.DB{}, errors.New("failed to connect to database")
	}
	return db, nil
}
