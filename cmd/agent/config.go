package main

import (
	"flag"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func updateCfg() (config.Agent, error) {
	cfg := config.Agent{}
	err := env.Parse(&cfg)

	var (
		addr          string
		loglevel      string
		logPath       string
		key           string
		reportPeriod  int
		pollingPeriod int
		rateLimit     int
	)

	flag.StringVar(&addr, "a", config.DefaultAddr, "address and port to run agent")
	flag.StringVar(&loglevel, "ll", config.DefaultLogLevel, "log level")
	flag.StringVar(&logPath, "f", "", "log file path")
	flag.StringVar(&key, "k", "", "key for sign")
	flag.IntVar(&reportPeriod, "r", config.DefaultReportInterval, "send data time interval")
	flag.IntVar(&pollingPeriod, "p", config.DefaultPollInterval, "get data time interval")
	flag.IntVar(&rateLimit, "l", config.DefaultRateLimit, "rate limit")
	flag.Parse()

	if cfg.Addr == config.DefaultAddr {
		cfg.Addr = addr
	}
	if cfg.LogLevel == config.DefaultLogLevel {
		cfg.LogLevel = loglevel
	}
	if cfg.LogPath == "" {
		cfg.LogPath = logPath
	}
	if cfg.ReportPeriod == config.DefaultReportInterval {
		cfg.ReportPeriod = reportPeriod
	}
	if cfg.PollingPeriod == config.DefaultPollInterval {
		cfg.PollingPeriod = pollingPeriod
	}
	if cfg.Key == "" {
		cfg.Key = key
	}
	if cfg.RateLimit == config.DefaultRateLimit {
		cfg.RateLimit = rateLimit
	}

	return cfg, err
}
