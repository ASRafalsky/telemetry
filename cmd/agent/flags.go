package main

import (
	"flag"
	"os"
	"strconv"
)

func parseFlags() (addr string, polling int, report int) {
	var (
		err error
	)

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if addr == "" {
		flag.StringVar(&addr, "a", ":8080", "address and port to run server")
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if report, err = strconv.Atoi(envReportInterval); err != nil {
			report = 0
		}
	}
	if report == 0 {
		flag.IntVar(&report, "r", 10, "send data time interval")
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if polling, err = strconv.Atoi(envPollInterval); err != nil {
			polling = 0
		}
	}
	if polling == 0 {
		flag.IntVar(&polling, "p", 2, "get data time interval")
	}

	flag.Parse()

	return addr, polling, report
}
