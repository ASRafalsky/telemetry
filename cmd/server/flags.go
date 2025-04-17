package main

import (
	"flag"
	"os"
	"strconv"
)

func parseFlags() (addr, logLevel, logPath, dumpPath string, storePeriod int, restore bool) {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
	if envLogPath := os.Getenv("LOG_PATH"); envLogPath != "" {
		logPath = envLogPath
	}
	if envDumpPath := os.Getenv("FILE_STORAGE_PATH"); envDumpPath != "" {
		dumpPath = envDumpPath
	}
	var (
		err                error
		storePeriodFromEnv bool
	)
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if storePeriod, err = strconv.Atoi(envStoreInterval); err == nil {
			storePeriodFromEnv = true
		}
	}
	var restoreFromEnv bool
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if restore, err = strconv.ParseBool(envRestore); err == nil {
			restoreFromEnv = true
		}
	}

	if addr == "" {
		flag.StringVar(&addr, "a", ":8080", "address and port to run server")
	}
	if logLevel == "" {
		flag.StringVar(&logLevel, "l", "info", "log level")
	}
	if logPath == "" {
		flag.StringVar(&logPath, "p", "", "log file path")
	}
	if dumpPath == "" {
		flag.StringVar(&dumpPath, "f", "./dump/dump", "dump file path")
	}
	if !restoreFromEnv {
		flag.BoolVar(&restore, "r", false, "need to restore from the dump")
	}
	if !storePeriodFromEnv {
		flag.IntVar(&storePeriod, "i", 300, "dump interval in seconds")
	}

	flag.Parse()

	return
}
