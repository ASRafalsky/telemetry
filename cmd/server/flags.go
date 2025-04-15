package main

import (
	"flag"
	"os"
)

func parseFlags() string {
	var flagRunAddr string
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		return envRunAddr
	}

	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	return flagRunAddr
}
