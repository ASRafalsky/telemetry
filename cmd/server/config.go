package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func updateCfg() (config.Server, error) {
	cfg := config.Server{}
	err := env.Parse(&cfg)

	var (
		restore     bool
		addr        string
		loglevel    string
		logPath     string
		dumpPath    string
		storePeriod int
		dsn         string
	)
	flag.BoolVar(&restore, "r", false, "need to restore from the dump")
	flag.StringVar(&addr, "a", config.DefaultAddr, "address and port to run server")
	flag.StringVar(&loglevel, "l", config.DefaultLogLevel, "log level")
	flag.StringVar(&logPath, "p", "", "log file path")
	flag.StringVar(&dumpPath, "f", config.DefaultDumpPath, "dump file path")
	flag.IntVar(&storePeriod, "i", config.DefaultDumpInterval, "dump interval in seconds")
	flag.StringVar(&dsn, "d", "", "database address")
	flag.Parse()

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

	return cfg, err
}
