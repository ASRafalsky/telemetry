// Package config contains agent and server config structs.
package config

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v10"
)

const (
	DefaultAddr               = ":8080"
	DefaultDiagAddr           = ":8081"
	DefaultDiagPath           = "./profile/"
	DefaultDBAddr             = ""
	DefaultLogLevel           = "info"
	DefaultDumpPath           = "./dump/dump"
	DefaultEmptyPath          = ""
	DefaultEmptyStr           = ""
	DefaultDumpInterval       = "300"
	DefaultReportInterval     = "10"
	DefaultPollInterval       = "2"
	DefaultReportIntervalTime = time.Duration(10) * time.Second
	DefaultPollIntervalTime   = time.Duration(2) * time.Second
	DefaultEmptyValue         = 0
)

type CommonFields struct {
	Addr     string `env:"ADDRESS" json:"address"`
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`
	LogPath  string `env:"LOG_PATH" json:"log_path"`
	Key      string `env:"KEY" json:"key"`
	Crypto   string `env:"CRYPTO_KEY" json:"crypto_key"`
}

type DB struct {
	DSN             string `env:"DATABASE_DSN" json:"database_dsn"`
	MigrationsPath  string
	MigrationsTable string
}

type Server struct {
	CommonFields
	DB             DB
	Diag           Diagnostics
	DumpPath       string `env:"FILE_STORAGE_PATH" json:"store_file"`
	StorePeriod    time.Duration
	StorePeriodStr string `env:"STORE_INTERVAL" json:"store_interval"`
	Restore        bool   `env:"RESTORE" json:"restore"`
	Subnet         string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	CIDR           *net.IPNet
	PrivateKey     *rsa.PrivateKey
}

type Diagnostics struct {
	Addr string `env:"DIAG_ADDRESS" json:"diag_address"`
	Path string `env:"DIAG_PATH" json:"diag_path"`
}

type Agent struct {
	CommonFields
	ReportPeriodStr  string `env:"REPORT_INTERVAL" json:"report_interval"`
	PollingPeriodStr string `env:"POLL_INTERVAL" json:"poll_interval"`
	RateLimit        int    `env:"RATE_LIMIT" json:"rate_limit"`
	ReportPeriod     time.Duration
	PollingPeriod    time.Duration
}

type configInit struct {
	path string `env:"CONFIG_PATH"`
}

func ReadCfg[T any](cfg T) (T, error) {
	var initCfg configInit
	// I'm too lazy to solve useless tasks.
	for n := range os.Args {
		if os.Args[n] == "-c" || os.Args[n] == "-config" {
			if len(os.Args) > n+1 {
				initCfg.path = os.Args[n]
			}
			break
		}
	}
	_ = env.Parse(&initCfg)
	return readCfg[T](initCfg.path, cfg)
}

var ErrEmptyCfgPath = errors.New("empty config path")

func readCfg[T any](path string, cfg T) (T, error) {
	if len(path) == 0 {
		return cfg, ErrEmptyCfgPath
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err = json.Unmarshal(buf, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func DefaultServer() Server {
	return Server{
		CommonFields: CommonFields{
			Addr:     DefaultAddr,
			LogLevel: DefaultLogLevel,
			LogPath:  DefaultEmptyPath,
			Key:      DefaultEmptyStr,
			Crypto:   DefaultEmptyPath,
		},
		DB: DB{
			DSN:             DefaultDBAddr,
			MigrationsPath:  DefaultEmptyStr,
			MigrationsTable: DefaultEmptyStr,
		},
		Diag: Diagnostics{
			Addr: DefaultDiagAddr,
			Path: DefaultDiagPath,
		},
		DumpPath:       DefaultDumpPath,
		StorePeriodStr: DefaultDumpInterval,
	}
}

func DefaultAgent() Agent {
	return Agent{
		CommonFields: CommonFields{
			Addr:     DefaultAddr,
			LogLevel: DefaultLogLevel,
			LogPath:  DefaultEmptyPath,
			Key:      DefaultEmptyStr,
			Crypto:   DefaultEmptyPath,
		},
		ReportPeriodStr:  DefaultReportInterval,
		PollingPeriodStr: DefaultPollInterval,
		RateLimit:        DefaultEmptyValue,
	}
}

func (s *Server) ParseDuration() error {
	var err error
	s.StorePeriod, err = time.ParseDuration(s.StorePeriodStr)
	if err != nil {
		period, err := strconv.Atoi(s.StorePeriodStr)
		if err != nil {
			return err
		}
		s.StorePeriod = time.Duration(period) * time.Second
	}
	return nil
}

func (s *Agent) ParseDuration() error {
	var err error
	s.PollingPeriod, err = time.ParseDuration(s.PollingPeriodStr)
	if err != nil {
		period, err := strconv.Atoi(s.PollingPeriodStr)
		if err != nil {
			return err
		}
		s.PollingPeriod = time.Duration(period) * time.Second
	}
	s.ReportPeriod, err = time.ParseDuration(s.ReportPeriodStr)
	if err != nil {
		period, err := strconv.Atoi(s.ReportPeriodStr)
		if err != nil {
			return err
		}
		s.ReportPeriod = time.Duration(period) * time.Second
	}
	return nil
}
