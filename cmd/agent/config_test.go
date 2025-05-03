package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func TestUpdateCfg_Default(t *testing.T) {
	defAgentCfg := config.AgentConfig{
		CommonFields: config.CommonFields{
			Addr:     config.DefaultAddr,
			LogLevel: config.DefaultLogLevel,
			LogPath:  "",
		},
		ReportPeriod:  config.DefaultReportInterval,
		PollingPeriod: config.DefaultPollInterval,
	}

	cfg, err := updateCfg()
	require.NoError(t, err)
	require.Equal(t, defAgentCfg, cfg)
}
