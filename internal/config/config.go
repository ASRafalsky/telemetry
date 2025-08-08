// Package config contains agent and server config structs.
package config

const (
	DefaultAddr           = ":8080"
	DefaultDiagAddr       = ":8081"
	DefaultDiagPath       = "./profile/"
	DefaultDBAddr         = ""
	DefaultLogLevel       = "info"
	DefaultDumpPath       = "./dump/dump"
	DefaultDumpInterval   = 300
	DefaultReportInterval = 10
	DefaultPollInterval   = 2
	DefaultRateLimit      = 0
)

type CommonFields struct {
	Addr     string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
	LogPath  string `env:"LOG_PATH"`
	Key      string `env:"KEY"`
}

type DB struct {
	DSN             string `env:"DATABASE_DSN"`
	MigrationsPath  string
	MigrationsTable string
}

type Server struct {
	CommonFields
	DB          DB
	Diag        Diagnostics
	DumpPath    string `env:"FILE_STORAGE_PATH"`
	StorePeriod int    `env:"STORE_INTERVAL"`
	Restore     bool   `env:"RESTORE"`
}

type Diagnostics struct {
	Addr string `env:"DIAG_ADDRESS"`
	Path string `env:"DIAG_PATH"`
}

type Agent struct {
	CommonFields
	ReportPeriod  int `env:"REPORT_INTERVAL"`
	PollingPeriod int `env:"POLL_INTERVAL"`
	RateLimit     int `env:"RATE_LIMIT"`
}
