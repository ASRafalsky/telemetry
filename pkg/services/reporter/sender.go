package reporter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/mailru/easyjson"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

func Send(ctx context.Context, addr string, interval time.Duration, client *httpclient.Client,
	repos map[string]Repository, log logger) {
	log.Info("Reporeter started with interval:", interval.String())

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
					if err := sendJSONData(ctx, addr, name, repos[name], client); err != nil {
						log.Error("[send/json] failed to send data] for", name, ":", err.Error())
					}
				case counter:
					if err := sendJSONData(ctx, addr, name, repos[name], client); err != nil {
						log.Error("[send/json] failed to send data] for", name, ":", err.Error())
					}
				default:
					log.Fatal("[Send] unknown metrics type:", name)
				}
			}
		}
	}
}

func sendJSONData(ctx context.Context, addr, mtype string, repo Repository, client *httpclient.Client) error {
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}
	var errRes error
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/"
		value, err := dataToMsg(mtype, k, v)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to marshal data for %s(%s); %w", mtype, k, err))
			return nil
		}
		resp, err := client.Post(urlData, bytes.NewReader(value), header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s(%s); %w", mtype, k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			errRes = multierr.Append(errRes, fmt.Errorf("bad status for %s(%s): %s", mtype, k, resp.Status))
		}
		return resp.Body.Close()
	})
	if err != nil {
		errRes = multierr.Append(errRes, err)
	}
	return errRes
}

func sendCounterData(ctx context.Context, addr string, repo Repository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	var errRes error
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/counter/" + k + "/" + types.BytesToCounter(v).String()
		value, err := dataToMsg("gauge", k, v)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to marshal data for %s; %w", k, err))
			return nil
		}
		resp, err := client.Post(urlData, bytes.NewReader(value), header)
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

func sendGaugeData(ctx context.Context, addr string, repo Repository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	var errRes error
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := addr + "/update/gauge/" + k + "/" + types.BytesToGauge(v).String()
		value, err := dataToMsg("gauge", k, v)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to marshal data for %s; %w", k, err))
			return nil
		}
		resp, err := client.Post(urlData, bytes.NewReader(value), header)
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

func dataToMsg(mtype, name string, d []byte) ([]byte, error) {
	metrics := transport.Metrics{
		ID:    name,
		MType: mtype,
	}
	switch mtype {
	case counter:
		value := int64(types.BytesToCounter(d))
		metrics.Delta = &value
	case gauge:
		value := float64(types.BytesToGauge(d))
		metrics.Value = &value
	default:
		return nil, fmt.Errorf("unknown metrics type: %s", mtype)
	}
	return easyjson.Marshal(metrics)
}

type logger interface {
	Info(msg ...string)
	Warn(msg ...string)
	Error(msg ...string)
	Debug(msg ...string)
	Fatal(msg ...string)
}

type Repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}
