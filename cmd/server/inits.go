package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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

func findPIDByPort(port string) (int, error) {
	cmd := exec.Command("lsof", "-t", "-i", ":"+port)
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return -1, nil // Порт свободен
		}
		return -1, err
	}

	pidStr := strings.TrimSpace(string(output))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return -1, fmt.Errorf("invalid pid: %s", pidStr)
	}

	return pid, nil
}

func killProcess(pid int) error {
	cmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	return cmd.Run()
}
