package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/cache"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

var (
	buildVersion, buildDate, buildCommit string
)

func main() {
	utils.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	cfg, err := updateCfg()
	if err != nil {
		panic(fmt.Errorf("failed to update config: %w", err))
	}

	logger, err := log.AddLoggerWith("info", "")
	if err != nil {
		panic(fmt.Errorf("failed to add logger: %w", err))
	}
	defer logger.Sync()

	client, laddr := newClient(cfg, logger)
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	repo := cache.New[string, []byte]()

	logger.Info("Agent started with address:", "http://"+cfg.Addr)

	pollerModule := poller.New(newPollerCfg(cfg))
	pollerModule.Run(ctx, poller.GetGaugeMetrics, repo, logger)
	pollerModule.Run(ctx, poller.GetCounterMetrics, repo, logger)
	pollerModule.Run(ctx, poller.GetPSMemMetrics, repo, logger)
	pollerModule.Run(ctx, poller.GetPSCPUMetrics, repo, logger)

	sendCtx, cancelSend := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		reporter.Send(sendCtx, "", newSenderCfg(cfg, laddr), client, repo, logger)
	}()
	pollerModule.WaitShutdown(logger)
	// Stop reporter.
	cancelSend()
	// Wait until reporter is stopped.
	wg.Wait()

	logger.Info("Agent stopped")
}
