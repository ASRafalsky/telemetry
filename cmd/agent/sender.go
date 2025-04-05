package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"

	"github.com/ASRafalsky/telemetry/internal"
)

func send(ctx context.Context, addr string, interval time.Duration, client *httpclient.Client,
	repos map[string]*repositoryUnit) {

	sendTimer := time.NewTicker(interval)
	defer sendTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-sendTimer.C:
			for name := range repos {
				switch name {
				case gauge:
					sendGaugeData(ctx, addr, repos[name], client)
				case counter:
					sendCounterData(ctx, addr, repos[name], client)
				default:
				}
			}
		}
	}
}

func sendCounterData(ctx context.Context, addr string, repo *repositoryUnit, client *httpclient.Client) {
	repo.mx.Lock()
	defer repo.mx.Unlock()

	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/counter/" + k + "/" + internal.BytesToCounter(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	})
	if err != nil {
		fmt.Printf("[poller/counter] Failed to send data; %s\n", err)
	}
}

func sendGaugeData(ctx context.Context, addr string, repo *repositoryUnit, client *httpclient.Client) {
	repo.mx.Lock()
	defer repo.mx.Unlock()

	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/counter/" + k + "/" + internal.BytesToGauge(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	})
	if err != nil {
		fmt.Printf("[poller/counter] Failed to send data; %s\n", err)
	}
}
