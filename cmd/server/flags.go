package main

import (
	"flag"
	"os"
)

func parseFlags() (addr, logLevel, path string) {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
	if envLogPath := os.Getenv("LOG_PATH"); envLogPath != "" {
		path = envLogPath
	}
	if addr == "" {
		flag.StringVar(&addr, "a", ":8080", "address and port to run server")
	}
	if logLevel == "" {
		flag.StringVar(&logLevel, "l", "info", "log level")
	}
	if path == "" {
		flag.StringVar(&path, "p", "", "log file path")
	}

	flag.Parse()

	return
}
