package main

import (
	"flag"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func updateCfg() (config.Server, error) {
	var migrationPath string
	_, file, _, ok := runtime.Caller(0)
	if ok {
		migrationPath = filepath.Dir(file) + "/db/migrations"
	}
	cfg := config.Server{}
	flag.BoolVar(&cfg.Restore, "r", false, "need to restore from the dump")
	flag.StringVar(&cfg.Addr, "a", config.DefaultAddr, "address and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", config.DefaultLogLevel, "log level")
	flag.StringVar(&cfg.LogPath, "p", "", "log file path")
	flag.StringVar(&cfg.DumpPath, "f", config.DefaultDumpPath, "dump file path")
	flag.IntVar(&cfg.StorePeriod, "i", config.DefaultDumpInterval, "dump interval in seconds")
	flag.StringVar(&cfg.DB.DSN, "d", "", "database address")
	flag.StringVar(&cfg.DB.MigrationsPath, "m", migrationPath, "path to migrations")
	flag.StringVar(&cfg.DB.MigrationsTable, "t", "", "name of migration table, where migrator writes own data")
	flag.StringVar(&cfg.Key, "k", "secret-key", "key for sign")
	flag.Parse()

	err := env.Parse(&cfg)
	return cfg, err
}

func newDiagnosticCfg(_ config.Server) config.Server {
	return config.Server{
		CommonFields: config.CommonFields{
			Addr: config.DefaultDiagAddr,
		},
		Diag: config.Diagnostics{
			Addr: config.DefaultDiagAddr,
			Path: config.DefaultDiagPath,
		},
	}
}
