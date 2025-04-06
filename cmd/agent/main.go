package main

import (
	"context"
	"time"

	"github.com/ASRafalsky/telemetry/cmd/agent/poller"
	"github.com/ASRafalsky/telemetry/cmd/agent/reporter"
	"github.com/ASRafalsky/telemetry/cmd/agent/repository"
)

func main() {
	addr, pollingPeriod, sendPeriod := parseFlags()

	client := NewClient()
	ctx := context.Background()

	repos := repository.NewRepositories()

	go poller.Poll(ctx, time.Duration(pollingPeriod)*time.Second, repos)
	go reporter.Send(ctx, addr, time.Duration(sendPeriod)*time.Second, client, repos)

	<-ctx.Done()
}
