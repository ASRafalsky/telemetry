package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func updateCfg() (config.ServerConfig, error) {
	cfg := config.ServerConfig{}
	err := env.Parse(&cfg)

	if envRestore := os.Getenv("RESTORE"); envRestore == "" {
		flag.BoolVar(&cfg.Restore, "r", false, "need to restore from the dump")
	}
	if cfg.Addr == config.DefaultAddr {
		flag.StringVar(&cfg.Addr, "a", config.DefaultAddr, "address and port to run server")
	}
	if cfg.LogLevel == config.DefaultLogLevel {
		flag.StringVar(&cfg.LogLevel, "l", config.DefaultLogLevel, "log level")
	}
	if cfg.LogPath == "" {
		flag.StringVar(&cfg.LogPath, "p", "", "log file path")
	}
	if cfg.DumpPath == config.DefaultDumpPath {
		flag.StringVar(&cfg.DumpPath, "f", config.DefaultDumpPath, "dump file path")
	}
	if cfg.StorePeriod == config.DefaultDumpInterval {
		flag.IntVar(&cfg.StorePeriod, "i", config.DefaultDumpInterval, "dump interval in seconds")
	}

	flag.Parse()
	return cfg, err
}
