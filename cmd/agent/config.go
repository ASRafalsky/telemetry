package main

import (
	"flag"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func updateCfg() (config.AgentConfig, error) {
	cfg := config.AgentConfig{}
	err := env.Parse(&cfg)

	if cfg.Addr == config.DefaultAddr {
		flag.StringVar(&cfg.Addr, "a", config.DefaultAddr, "address and port to run agent")
	}
	if cfg.LogLevel == config.DefaultLogLevel {
		flag.StringVar(&cfg.LogLevel, "l", config.DefaultLogLevel, "log level")
	}
	if cfg.LogPath == "" {
		flag.StringVar(&cfg.LogPath, "f", "", "log file path")
	}
	if cfg.ReportPeriod == config.DefaultReportInterval {
		flag.IntVar(&cfg.ReportPeriod, "r", config.DefaultReportInterval, "send data time interval")
	}
	if cfg.PollingPeriod == config.DefaultPollInterval {
		flag.IntVar(&cfg.PollingPeriod, "p", config.DefaultPollInterval, "get data time interval")
	}

	flag.Parse()
	return cfg, err
}
