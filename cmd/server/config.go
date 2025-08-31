package main

import (
	"errors"
	"flag"
	"net"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/env/v10"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/utils"
)

func updateCfg() (config.Server, error) {
	var migrationPath string
	_, file, _, ok := runtime.Caller(0)
	if ok {
		migrationPath = filepath.Dir(file) + "/db/migrations"
	}

	cfg, errReadCfg := config.ReadCfg(config.DefaultServer())
	if errors.Is(errReadCfg, config.ErrEmptyCfgPath) {
		errReadCfg = nil
	}
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "need to restore from the dump")
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "address and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "log level")
	flag.StringVar(&cfg.LogPath, "p", cfg.LogPath, "log file path")
	flag.StringVar(&cfg.DumpPath, "f", cfg.DumpPath, "dump file path")
	flag.StringVar(&cfg.StorePeriodStr, "i", cfg.StorePeriodStr, "dump interval in seconds")
	flag.StringVar(&cfg.DB.DSN, "d", cfg.DB.DSN, "database address")
	flag.StringVar(&cfg.DB.MigrationsPath, "m", migrationPath, "path to migrations")
	flag.StringVar(&cfg.DB.MigrationsTable, "mt", config.DefaultEmptyStr, "name of migration table, where migrator writes own data")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "key for sign")
	flag.StringVar(&cfg.Crypto, "crypto-key", cfg.Crypto, "private key path")
	flag.StringVar(&cfg.Subnet, "t", cfg.Crypto, "trusted subnet")
	flag.BoolVar(&cfg.GRPC, "g", cfg.GRPC, "use grpc")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, multierr.Append(err, errReadCfg)
	}

	if err = cfg.ParseDuration(); err != nil {
		return cfg, multierr.Append(err, errReadCfg)
	}

	if cfg.Subnet != "" {
		if _, cfg.CIDR, err = net.ParseCIDR(cfg.Subnet); err != nil {
			return cfg, multierr.Append(err, errReadCfg)
		}
	}

	cfg.PrivateKey, err = utils.ParsePrivateKey(cfg.Crypto)

	return cfg, multierr.Append(err, errReadCfg)
}

func newDiagnosticCfg(_ config.Server) config.Server {
	return config.Server{
		CommonFields: config.CommonFields{
			Addr: config.DefaultDiagAddr,
		},
		Diag: config.Diagnostics{
			Addr: config.DefaultDiagAddr,
			Path: config.DefaultDiagPath,
		},
	}
}
