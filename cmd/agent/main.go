package main

import (
	"context"
	"fmt"
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

	fmt.Printf("Agent started with address: %s\n", "http://"+addr)
	go poller.Poll(ctx, time.Duration(pollingPeriod)*time.Second, repos)
	go reporter.Send(ctx, "http://"+addr, time.Duration(sendPeriod)*time.Second, client, repos)

	<-ctx.Done()
}
