package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func TestUpdateCfg_Default(t *testing.T) {
	defAgentCfg := config.Agent{
		CommonFields: config.CommonFields{
			Addr:     config.DefaultAddr,
			LogLevel: config.DefaultLogLevel,
			LogPath:  config.DefaultEmptyPath,
		},
		ReportPeriodStr:  config.DefaultReportInterval,
		PollingPeriodStr: config.DefaultPollInterval,
		ReportPeriod:     config.DefaultReportIntervalTime,
		PollingPeriod:    config.DefaultPollIntervalTime,
		RateLimit:        config.DefaultEmptyValue,
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
		reportPeriodNum  = time.Duration(111) * time.Second
		pollingPeriodNum = time.Duration(222) * time.Second
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
		ReportPeriodStr:  reportPeriod,
		PollingPeriodStr: pollingPeriod,
		ReportPeriod:     reportPeriodNum,
		PollingPeriod:    pollingPeriodNum,
		RateLimit:        rateLimitNum,
	}

	require.Equal(t, expectedCfg, cfg)
}
