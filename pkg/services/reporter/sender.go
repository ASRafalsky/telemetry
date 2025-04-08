package reporter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/types"
	"github.com/ASRafalsky/telemetry/pkg/services/repository"
)

func Send(ctx context.Context, addr string, interval time.Duration, client *httpclient.Client,
	repos map[string]repository.Repository) {
	fmt.Printf("Reporeter started with interval %v\n", interval)

	sendTimer := time.NewTicker(interval)
	defer sendTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-sendTimer.C:
			for name := range repos {
				switch name {
				case repository.Gauge:
					sendGaugeData(ctx, addr, repos[name], client)
				case repository.Counter:
					sendCounterData(ctx, addr, repos[name], client)
				default:
				}
			}
		}
	}
}

func sendCounterData(ctx context.Context, addr string, repo repository.Repository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	var errRes error
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/counter/" + k + "/" + types.BytesToCounter(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s; %w", k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("[send/counter] Status code: %s\n", resp.Status)
		}
		return resp.Body.Close()
	})
	if errRes = multierr.Append(errRes, err); err != nil {
		fmt.Printf("[send/counter] Failed to send data; %s\n", err)
	}
}

func sendGaugeData(ctx context.Context, addr string, repo repository.Repository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	var errRes error
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/gauge/" + k + "/" + types.BytesToGauge(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s; %w", k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("[send/counter] Status code: %s\n", resp.Status)
		}
		return resp.Body.Close()
	})
	if errRes = multierr.Append(errRes, err); err != nil {
		fmt.Printf("[send/counter] Failed to send data; %s\n", err)
	}
}
