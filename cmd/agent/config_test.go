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

func TestUpdateCfg_EnvVarPriority(t *testing.T) {
	const (
		addr             = "env_address"
		logLevel         = "env_log_level"
		logPath          = "env_log_path"
		secret           = "env_secret"
		reportPeriod     = "333"
		pollingPeriod    = "555"
		rateLimit        = "777"
		reportPeriodNum  = 333
		pollingPeriodNum = 555
		rateLimitNum     = 777
	)
	require.NoError(t, os.Setenv("ADDRESS", addr))
	defer require.NoError(t, os.Unsetenv("ADDRESS"))
	require.NoError(t, os.Setenv("LOG_LEVEL", logLevel))
	defer require.NoError(t, os.Unsetenv("LOG_LEVEL"))
	require.NoError(t, os.Setenv("LOG_PATH", logPath))
	defer require.NoError(t, os.Unsetenv("LOG_PATH"))
	require.NoError(t, os.Setenv("KEY", secret))
	defer require.NoError(t, os.Unsetenv("KEY"))
	require.NoError(t, os.Setenv("REPORT_INTERVAL", reportPeriod))
	defer require.NoError(t, os.Unsetenv("REPORT_INTERVAL"))
	require.NoError(t, os.Setenv("POLL_INTERVAL", pollingPeriod))
	defer require.NoError(t, os.Unsetenv("POLL_INTERVAL"))
	require.NoError(t, os.Setenv("RATE_LIMIT", rateLimit))
	defer require.NoError(t, os.Unsetenv("RATE_LIMIT"))

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := updateCfg()
	require.NoError(t, err)

	expectedCfg := config.Agent{
		CommonFields: config.CommonFields{
			Addr:     addr,
			LogLevel: logLevel,
			LogPath:  logPath,
		},
		ReportPeriod:  reportPeriodNum,
		PollingPeriod: pollingPeriodNum,
		RateLimit:     rateLimitNum,
	}

	require.Equal(t, expectedCfg, cfg)
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
