package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReadSrvCfg(t *testing.T) {
	cfg, err := readCfg[Server]("./testdata/server.cfg", DefaultServer())
	require.NoError(t, err)

	require.NoError(t, cfg.ParseDuration())

	exp := Server{
		CommonFields: CommonFields{
			Addr:     "localhost:8080",
			Crypto:   "/path/to/key.pem",
			LogLevel: DefaultLogLevel,
			LogPath:  DefaultEmptyPath,
		},
		Diag: Diagnostics{
			Addr: DefaultDiagAddr,
			Path: DefaultDiagPath,
		},
		DB: DB{
			DSN:             DefaultDBAddr,
			MigrationsPath:  DefaultEmptyPath,
			MigrationsTable: DefaultEmptyStr,
		},
		DumpPath:       "/path/to/file.db",
		Restore:        true,
		StorePeriodStr: "1s",
		StorePeriod:    time.Second,
	}
	require.Equal(t, exp, cfg)
}

func TestReadAgentCfg(t *testing.T) {
	cfg, err := readCfg[Agent]("./testdata/agent.cfg", DefaultAgent())
	require.NoError(t, err)

	require.NoError(t, cfg.ParseDuration())

	exp := Agent{
		CommonFields: CommonFields{
			Addr:     "localhost:8080",
			Crypto:   "/path/to/key.pem",
			LogLevel: DefaultLogLevel,
			LogPath:  DefaultEmptyPath,
		},
		PollingPeriodStr: "2",
		PollingPeriod:    time.Duration(2) * time.Second,
		ReportPeriodStr:  "10s",
		ReportPeriod:     time.Duration(10) * time.Second,
	}
	require.Equal(t, exp, cfg)
}
