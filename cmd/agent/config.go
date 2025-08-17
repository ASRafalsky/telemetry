package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v10"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
	"github.com/ASRafalsky/telemetry/internal/utils"
)

func updateCfg() (config.Agent, error) {
	cfg, errReadCfg := config.ReadCfg(config.DefaultAgent())
	if errReadCfg != nil {
		cfg = config.DefaultAgent()
	}
	if errors.Is(errReadCfg, config.ErrEmptyCfgPath) {
		errReadCfg = nil
	}
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "address and port to run agent")
	flag.StringVar(&cfg.LogLevel, "ll", cfg.LogLevel, "log level")
	flag.StringVar(&cfg.LogPath, "f", cfg.LogPath, "log file path")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "key for sign")
	flag.StringVar(&cfg.ReportPeriodStr, "r", cfg.ReportPeriodStr, "send data time interval")
	flag.StringVar(&cfg.PollingPeriodStr, "p", cfg.PollingPeriodStr, "get data time interval")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit")
	flag.StringVar(&cfg.Crypto, "crypto-key", cfg.Crypto, "public key path")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return config.Agent{}, err
	}

	if err = cfg.ParseDuration(); err != nil {
		return config.Agent{}, err
	}

	return cfg, nil
}

func newPollerCfg(cfg config.Agent) poller.Config {
	return poller.Config{
		Interval: cfg.PollingPeriod,
	}
}

func newSenderCfg(cfg config.Agent) reporter.Config {
	res := reporter.Config{
		Interval:  cfg.ReportPeriod,
		Address:   "http://" + cfg.Addr,
		Key:       cfg.Key,
		RateLimit: cfg.RateLimit,
	}

	var err error
	res.PubKey, err = utils.ParsePublicKey(cfg.Crypto)
	if err != nil {
		panic(fmt.Errorf("failed to parse public key from config: %w", err))
	}
	return res
}
