package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

const (
	defaultURL = "http://localhost:8080"
	gauge      = "gauge"
	counter    = "counter"
)

func main() {
	adress := flag.String("a", defaultURL, "server address")
	sendPeriod := flag.Int("r", 10, "send data time interval")
	pollingPeriod := flag.Int("p", 2, "get data time interval")
	flag.Parse()
	fmt.Printf("Server address: %s, Polling time interval: %d sec, Send time interval: %d sec.\n",
		*adress, *pollingPeriod, *sendPeriod)

	client := NewClient()
	ctx := context.Background()

	repos := newRepositories()

	go poll(ctx, time.Duration(*pollingPeriod)*time.Second, repos)
	go send(ctx, *adress, time.Duration(*sendPeriod)*time.Second, client, repos)

	<-ctx.Done()
}

type repositoryUnit struct {
	mx sync.Mutex
	Repository
}

func newRepositories() map[string]*repositoryUnit {
	return map[string]*repositoryUnit{
		gauge: {
			mx:         sync.Mutex{},
			Repository: storage.New[string, []byte](),
		},
		counter: {
			mx:         sync.Mutex{},
			Repository: storage.New[string, []byte](),
		},
	}
}
