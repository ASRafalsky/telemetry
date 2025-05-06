package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func TestUpdateCfg_Default(t *testing.T) {
	defSrvCfg := config.Server{
		CommonFields: config.CommonFields{
			Addr:     config.DefaultAddr,
			LogLevel: config.DefaultLogLevel,
			LogPath:  "",
		},
		DB: config.DB{
			DSN: config.DefaultDBAddr,
		},
		DumpPath:    config.DefaultDumpPath,
		StorePeriod: config.DefaultDumpInterval,
		Restore:     false,
	}

	cfg, err := updateCfg()
	require.NoError(t, err)
	require.Equal(t, defSrvCfg, cfg)
}
