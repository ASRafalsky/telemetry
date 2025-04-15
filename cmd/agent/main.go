package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ASRafalsky/telemetry/pkg/services/poller"
	"github.com/ASRafalsky/telemetry/pkg/services/reporter"
	"github.com/ASRafalsky/telemetry/pkg/services/repository"
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
