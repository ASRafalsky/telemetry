package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
)

func TestUpdateCfg_Default(t *testing.T) {
	defSrvCfg := config.Server{
		CommonFields: config.CommonFields{
			Addr:     config.DefaultAddr,
			LogLevel: config.DefaultLogLevel,
			LogPath:  config.DefaultEmptyPath,
		},
		DB: config.DB{
			DSN:            config.DefaultDBAddr,
			MigrationsPath: migrationPath(t),
		},
		Diag: config.Diagnostics{
			Addr: config.DefaultDiagAddr,
			Path: config.DefaultDiagPath,
		},
		DumpPath:       config.DefaultDumpPath,
		StorePeriodStr: config.DefaultDumpInterval,
		StorePeriod:    time.Duration(300) * time.Second,
		Restore:        false,
	}

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := updateCfg()
	require.NoError(t, err)
	require.Equal(t, defSrvCfg, cfg)
}

func TestUpdateCfg_EnvVarPriority(t *testing.T) {
	const (
		addr           = "env_address"
		logLevel       = "env_log_level"
		logPath        = "env_log_path"
		secret         = "env_secret"
		dbAddr         = "env_db_addr"
		migrationPath  = "/env/path/to/migration"
		migrationTable = "env_migration_table"
		dumpPath       = "/env/path/to/dump"
		storePeriod    = "555"
		storePeriodNum = time.Duration(555) * time.Second
		restore        = "true"
	)

	require.NoError(t, os.Setenv("ADDRESS", addr))
	defer require.NoError(t, os.Unsetenv("ADDRESS"))
	if os.Getenv("ADDRESS") == "" {
		t.Skip("Can not set ADDRESS environment variable")
	}
	require.NoError(t, os.Setenv("LOG_LEVEL", logLevel))
	defer require.NoError(t, os.Unsetenv("LOG_LEVEL"))
	if os.Getenv("LOG_LEVEL") == "" {
		t.Skip("Can not set LOG_LEVEL environment variable")
	}
	require.NoError(t, os.Setenv("LOG_PATH", logPath))
	defer require.NoError(t, os.Unsetenv("LOG_PATH"))
	if os.Getenv("LOG_PATH") == "" {
		t.Skip("Can not set LOG_PATH environment variable")
	}
	require.NoError(t, os.Setenv("KEY", secret))
	defer require.NoError(t, os.Unsetenv("KEY"))
	if os.Getenv("KEY") == "" {
		t.Skip("Can not set KEY environment variable")
	}
	require.NoError(t, os.Setenv("DATABASE_DSN", dbAddr))
	defer require.NoError(t, os.Unsetenv("DATABASE_DSN"))
	if os.Getenv("DATABASE_DSN") == "" {
		t.Skip("Can not set DATABASE_DSN environment variable")
	}
	require.NoError(t, os.Setenv("FILE_STORAGE_PATH", dumpPath))
	defer require.NoError(t, os.Unsetenv("FILE_STORAGE_PATH"))
	if os.Getenv("FILE_STORAGE_PATH") == "" {
		t.Skip("Can not set FILE_STORAGE_PATH environment variable")
	}
	require.NoError(t, os.Setenv("STORE_INTERVAL", storePeriod))
	defer require.NoError(t, os.Unsetenv("STORE_INTERVAL"))
	if os.Getenv("STORE_INTERVAL") == "" {
		t.Skip("Can not set STORE_INTERVAL environment variable")
	}
	require.NoError(t, os.Setenv("RESTORE", restore))
	defer require.NoError(t, os.Unsetenv("RESTORE"))
	if os.Getenv("RESTORE") == "" {
		t.Skip("Can not set RESTORE environment variable")
	}

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := updateCfg()
	require.NoError(t, err)

	expectedSrvCfg := config.Server{
		CommonFields: config.CommonFields{
			Addr:     addr,
			LogLevel: logLevel,
			LogPath:  logPath,
		},
		DB: config.DB{
			DSN:             dbAddr,
			MigrationsPath:  migrationPath,
			MigrationsTable: migrationTable,
		},
		DumpPath:    dumpPath,
		StorePeriod: storePeriodNum,
		Restore:     true,
	}

	require.Equal(t, expectedSrvCfg, cfg)
}

func TestUpdateCfg_FlagPriority(t *testing.T) {
	const (
		addr     = "flag_address"
		logLevel = "flag_log_level"
		secret   = "flag_secret"
		dbAddr   = "flag_db_addr"
		restore  = "true"
	)
	os.Args = []string{"server",
		"-d", dbAddr,
		"-a", addr,
		"-l", logLevel,
		"-k", secret,
		"-r", restore,
	}

	defer func() { os.Args = []string{"cmd"} }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := updateCfg()
	require.NoError(t, err)

	expectedSrvCfg := config.Server{
		CommonFields: config.CommonFields{
			Addr:     addr,
			LogLevel: logLevel,
			Key:      secret,
		},
		DB: config.DB{
			DSN:            dbAddr,
			MigrationsPath: migrationPath(t),
		},
		Diag: config.Diagnostics{
			Addr: config.DefaultDiagAddr,
			Path: config.DefaultDiagPath,
		},
		DumpPath:       config.DefaultDumpPath,
		StorePeriodStr: config.DefaultDumpInterval,
		StorePeriod:    time.Duration(300) * time.Second,
		Restore:        true,
	}

	require.Equal(t, expectedSrvCfg, cfg)
}

func migrationPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Dir(file) + "/db/migrations"
}
