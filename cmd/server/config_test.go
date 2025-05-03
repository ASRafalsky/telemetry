package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func TestUpdateCfg_Default(t *testing.T) {
	defSrvCfg := config.ServerConfig{
		CommonFields: config.CommonFields{
			Addr:     config.DefaultAddr,
			LogLevel: config.DefaultLogLevel,
			LogPath:  "",
		},
		DumpPath:    config.DefaultDumpPath,
		StorePeriod: config.DefaultDumpInterval,
		Restore:     false,
	}

	cfg, err := updateCfg()
	require.NoError(t, err)
	require.Equal(t, defSrvCfg, cfg)
}
