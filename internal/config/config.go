package config

const (
	DefaultAddr           = ":8080"
	DefaultDBAddr         = ""
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

type DB struct {
	DSN string `env:"DATABASE_DSN" envDefault:""`
}

type Server struct {
	CommonFields
	DB          DB
	DumpPath    string `env:"FILE_STORAGE_PATH" envDefault:"./dump/dump"`
	StorePeriod int    `env:"STORE_INTERVAL" envDefault:"300"`
	Restore     bool   `env:"RESTORE" envDefault:"false"`
}

type Agent struct {
	CommonFields
	ReportPeriod  int `env:"REPORT_INTERVAL" envDefault:"10"`
	PollingPeriod int `env:"POLL_INTERVAL" envDefault:"2"`
}
