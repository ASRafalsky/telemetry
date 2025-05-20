package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func updateCfg() (config.Server, error) {
	cfg := config.Server{}
	err := env.Parse(&cfg)

	var (
		restore         bool
		addr            string
		loglevel        string
		logPath         string
		dumpPath        string
		storePeriod     int
		dsn             string
		migrationsPath  string
		migrationsTable string
		key             string
	)

	flag.BoolVar(&restore, "r", false, "need to restore from the dump")
	flag.StringVar(&addr, "a", config.DefaultAddr, "address and port to run server")
	flag.StringVar(&loglevel, "l", config.DefaultLogLevel, "log level")
	flag.StringVar(&logPath, "p", "", "log file path")
	flag.StringVar(&dumpPath, "f", config.DefaultDumpPath, "dump file path")
	flag.IntVar(&storePeriod, "i", config.DefaultDumpInterval, "dump interval in seconds")
	flag.StringVar(&dsn, "d", "", "database address")
	flag.StringVar(&migrationsPath, "m", "", "path to migrations")
	flag.StringVar(&migrationsTable, "t", "", "name of migration table, where migrator writes own data")
	flag.StringVar(&key, "k", "", "key for sign")
	flag.Parse()

	if migrationsPath == "" {
		_, file, _, ok := runtime.Caller(0)
		if ok {
			cfg.DB.MigrationsPath = filepath.Dir(file) + "/db/migrations"
		}
	}

	if envRestore := os.Getenv("RESTORE"); envRestore == "" {
		cfg.Restore = restore
	}
	if cfg.Addr == config.DefaultAddr {
		cfg.Addr = addr
	}
	if cfg.LogLevel == config.DefaultLogLevel {
		cfg.LogLevel = loglevel
	}
	if cfg.LogPath == "" {
		cfg.LogPath = logPath
	}
	if cfg.DumpPath == config.DefaultDumpPath {
		cfg.DumpPath = dumpPath
	}
	if cfg.StorePeriod == config.DefaultDumpInterval {
		cfg.StorePeriod = storePeriod
	}
	if cfg.DB.DSN == "" {
		cfg.DB.DSN = dsn
	}
	if cfg.Key == "" {
		cfg.Key = key
	}

	return cfg, err
}
