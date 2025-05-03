package config

const (
	DefaultAddr           = ":8080"
	DefaultLogLevel       = "info"
	DefaultDumpPath       = "./dump/dump"
	DefaultDumpInterval   = 300
	DefaultReportInterval = 10
	DefaultPollInterval   = 2
)

type CommonFields struct {
	Addr     string `env:"ADDRESS" envDefault:":8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	LogPath  string `env:"LOG_PATH" envDefault:""`
}

type ServerConfig struct {
	CommonFields
	DumpPath    string `env:"FILE_STORAGE_PATH" envDefault:"./dump/dump"`
	StorePeriod int    `env:"STORE_INTERVAL" envDefault:"300"`
	Restore     bool   `env:"RESTORE" envDefault:"false"`
}

type AgentConfig struct {
	CommonFields
	ReportPeriod  int `env:"REPORT_INTERVAL" envDefault:"10"`
	PollingPeriod int `env:"POLL_INTERVAL" envDefault:"2"`
}
