package main

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/internal/cache"
	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

func main() {
	cfg, err := updateCfg()
	if err != nil {
		panic(err)
	}

	logger, err := log.AddLoggerWith("info", "")
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	client := newClient()
	ctx := context.Background()

	repo := cache.New[string, []byte]()

	logger.Info("Agent started with address:", "http://"+cfg.Addr)

	go poller.Poll(ctx, poller.GetGaugeMetrics, time.Duration(cfg.PollingPeriod)*time.Second, repo, logger)
	go poller.Poll(ctx, poller.GetCounterMetrics, time.Duration(cfg.ReportPeriod)*time.Second, repo, logger)

	go reporter.Send(ctx, "http://"+cfg.Addr, "", cfg.Key, time.Duration(cfg.ReportPeriod)*time.Second, client, repo, logger)

	<-ctx.Done()
}
