package main

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
)

func updateCfg() (config.Agent, error) {
	cfg := config.Agent{}
	flag.StringVar(&cfg.Addr, "a", config.DefaultAddr, "address and port to run agent")
	flag.StringVar(&cfg.LogLevel, "ll", config.DefaultLogLevel, "log level")
	flag.StringVar(&cfg.LogPath, "f", "", "log file path")
	flag.StringVar(&cfg.Key, "k", "", "key for sign")
	flag.IntVar(&cfg.ReportPeriod, "r", config.DefaultReportInterval, "send data time interval")
	flag.IntVar(&cfg.PollingPeriod, "p", config.DefaultPollInterval, "get data time interval")
	flag.IntVar(&cfg.RateLimit, "l", config.DefaultRateLimit, "rate limit")
	flag.Parse()

	err := env.Parse(&cfg)
	return cfg, err
}

func newPollerCfg(cfg config.Agent) poller.Config {
	return poller.Config{
		Interval: time.Duration(cfg.PollingPeriod) * time.Second,
	}
}

func newSenderCfg(cfg config.Agent) reporter.Config {
	return reporter.Config{
		Interval:  time.Duration(cfg.ReportPeriod) * time.Second,
		Address:   cfg.Addr,
		Key:       cfg.Key,
		RateLimit: cfg.RateLimit,
	}
}
