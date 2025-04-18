package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

func Send(ctx context.Context, addr, mType string, interval time.Duration, client *httpclient.Client,
	repo repository, log logger) {
	log.Info("Reporeter started with interval:", interval.String())

	sendTimer := time.NewTicker(interval)
	defer sendTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-sendTimer.C:
			if err := sendJSONData(ctx, addr, mType, repo, client); err != nil {
				log.Error("[send/json] failed to send data] for", mType, ":", err.Error())
			}
		}
	}
}

func sendJSONData(ctx context.Context, addr, mtype string, repo repository, client *httpclient.Client) error {
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}
	var (
		bufToSend = bytes.NewBuffer(nil)
		errRes    error
	)

	zw := gzip.NewWriter(bufToSend)
	err := serializeMetrics(ctx, mtype, repo, zw)
	zw.Close()
	if bufToSend.Len() == 0 {
		if err != nil {
			return err
		}
		return nil
	}
	header.Set("Content-Encoding", "gzip")
	resp, err := client.Post(addr+"/update/", bufToSend, header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errRes = multierr.Append(errRes, fmt.Errorf("bad status for %s: %s", mtype, resp.Status))
	}
	return errRes
}

func sendCounterData(ctx context.Context, addr string, repo repository, client *httpclient.Client) {
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

func sendGaugeData(ctx context.Context, addr string, repo repository, client *httpclient.Client) {
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

func serializeMetrics(ctx context.Context, mtype string, repo repository, wc writerCloser) error {
	var errRes error
	_ = repo.ForEach(ctx, func(k string, v []byte) error {
		metric, err := dataToMetrics(mtype, k, v)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to marshal data for %s(%s); %w", mtype, k, err))
			return nil
		}
		if err := transport.SerializeMetrics(metric, wc); err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to compress data for %s(%s); %w", mtype, k, err))
			return nil
		}
		return nil
	})
	return errRes
}

func dataToMetrics(mtype, name string, d []byte) (transport.Metrics, error) {
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
		return transport.Metrics{}, fmt.Errorf("unknown metrics type: %s", mtype)
	}
	return metrics, nil
}

type logger interface {
	Info(msg ...string)
	Warn(msg ...string)
	Error(msg ...string)
	Debug(msg ...string)
	Fatal(msg ...string)
}

type repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
}

type writerCloser interface {
	Write(p []byte) (n int, err error)
	Close() error
}
