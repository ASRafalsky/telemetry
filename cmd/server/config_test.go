package main

import (
	"path/filepath"
	"runtime"
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
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	defSrvCfg.DB.MigrationsPath = filepath.Dir(file) + "/db/migrations"

	cfg, err := updateCfg()
	require.NoError(t, err)
	require.Equal(t, defSrvCfg, cfg)
}
