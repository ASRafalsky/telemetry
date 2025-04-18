package main

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/pkg/services/poller"
	"github.com/ASRafalsky/telemetry/pkg/services/reporter"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

func main() {
	addr, pollingPeriod, sendPeriod := parseFlags()

	logger, err := log.AddLoggerWith("info", "")
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	client := NewClient()
	ctx := context.Background()

	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	logger.Info("Agent started with address:", "http://"+addr)

	go poller.Poll(ctx, poller.GetGaugeMetrics, time.Duration(pollingPeriod)*time.Second, gaugeRepo, logger)
	go poller.Poll(ctx, poller.GetCounterMetrics, time.Duration(pollingPeriod)*time.Second, counterRepo, logger)

	go reporter.Send(ctx, "http://"+addr, gauge, time.Duration(sendPeriod)*time.Second, client, gaugeRepo, logger)
	go reporter.Send(ctx, "http://"+addr, counter, time.Duration(sendPeriod)*time.Second, client, counterRepo, logger)

	<-ctx.Done()
}
