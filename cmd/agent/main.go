package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
	"github.com/ASRafalsky/telemetry/pkg/cache"
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	repo := cache.New[string, []byte]()

	logger.Info("Agent started with address:", "http://"+cfg.Addr)

	pollerCfg := newPollerCfg(cfg)
	go poller.Poll(ctx, poller.GetGaugeMetrics, pollerCfg, repo, logger)
	go poller.Poll(ctx, poller.GetCounterMetrics, pollerCfg, repo, logger)
	go poller.Poll(ctx, poller.GetPSMemMetrics, pollerCfg, repo, logger)
	go poller.Poll(ctx, poller.GetPSCPUMetrics, pollerCfg, repo, logger)

	go reporter.Send(ctx, "", newSenderCfg(cfg), client, repo, logger)

	<-ctx.Done()
	logger.Info("Agent stopped")
}
