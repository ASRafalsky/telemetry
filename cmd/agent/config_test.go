package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func TestUpdateCfg_Default(t *testing.T) {
	defAgentCfg := config.Agent{
		CommonFields: config.CommonFields{
			Addr:     config.DefaultAddr,
			LogLevel: config.DefaultLogLevel,
			LogPath:  "",
		},
		ReportPeriod:  config.DefaultReportInterval,
		PollingPeriod: config.DefaultPollInterval,
		RateLimit:     config.DefaultRateLimit,
	}
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := updateCfg()
	require.NoError(t, err)
	require.Equal(t, defAgentCfg, cfg)
}

func TestUpdateCfg_FlagPriority(t *testing.T) {
	const (
		addr             = "flag_address"
		logLevel         = "flag_log_level"
		logPath          = "flag_log_path"
		secret           = "flag_secret"
		reportPeriod     = "111"
		pollingPeriod    = "222"
		rateLimit        = "333"
		reportPeriodNum  = 111
		pollingPeriodNum = 222
		rateLimitNum     = 333
	)
	os.Args = []string{"agent",
		"-a", addr,
		"-ll", logLevel,
		"-f", logPath,
		"-k", secret,
		"-r", reportPeriod,
		"-p", pollingPeriod,
		"-l", rateLimit,
	}
	defer func() { os.Args = []string{"cmd"} }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := updateCfg()
	require.NoError(t, err)

	expectedCfg := config.Agent{
		CommonFields: config.CommonFields{
			Addr:     "flag_address",
			LogLevel: logLevel,
			LogPath:  logPath,
			Key:      secret,
		},
		ReportPeriod:  reportPeriodNum,
		PollingPeriod: pollingPeriodNum,
		RateLimit:     rateLimitNum,
	}

	require.Equal(t, expectedCfg, cfg)
}
