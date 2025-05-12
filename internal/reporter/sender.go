package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"go.uber.org/multierr"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	gauge     = "gauge"
	counter   = "counter"
	batchSize = 1000
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
			if err := sendJSONData(ctx, addr, "", repo, client); err != nil {
				log.Error("[send/json] failed to send data] for", mType, ":", err.Error())
			}
		}
	}
}

func sendJSONData(ctx context.Context, addr, mtype string, repo repository, client *httpclient.Client) (err error) {
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	var bufToSend = bytes.NewBuffer(nil)

	zw := gzip.NewWriter(bufToSend)
	err = serializeMetrics(ctx, mtype, repo, zw)
	if errZw := zw.Close(); errZw != nil {
		err = multierr.Append(err, fmt.Errorf("failed to close gzip writer: %w", errZw))
	}
	if bufToSend.Len() == 0 {
		if err != nil {
			return err
		}
		return nil
	}
	header.Set("Content-Encoding", "gzip")
	resp, errPost := client.Post(addr+"/updates/", bufToSend, header)
	if errPost != nil {
		err = multierr.Append(err, fmt.Errorf("failed to post update: %w", errPost))
		return err
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			err = multierr.Append(err, fmt.Errorf("failed to close response body: %w", errClose))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err = multierr.Append(err, fmt.Errorf("bad status for %s: %s", mtype, resp.Status))
	}
	return err
}

func sendCounterData(ctx context.Context, addr string, repo repository, client *httpclient.Client) error {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	var errRes error
	return repo.ForEach(ctx, func(k string, v []byte) error {
		if !strings.HasPrefix(k, counter) {
			return nil
		}
		urlData := addr + "/update/counter/" + strings.TrimPrefix(k, counter) + "/" + types.BytesToCounter(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s; %w", k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			errRes = multierr.Append(errRes, fmt.Errorf("bad response status for %s; %s", k, resp.Status))
		}
		if errClose := resp.Body.Close(); errClose != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to close response body for %s; %w", k, err))
		}
		return errRes
	})
}

func sendGaugeData(ctx context.Context, addr string, repo repository, client *httpclient.Client) error {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	var errRes error
	return repo.ForEach(ctx, func(k string, v []byte) error {
		if !strings.HasPrefix(k, gauge) {
			return nil
		}
		urlData := addr + "/update/gauge/" + strings.TrimPrefix(k, gauge) + "/" + types.BytesToGauge(v).String()
		resp, err := client.Post(urlData, nil, header)
		if err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to send data for %s; %w", k, err))
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			errRes = multierr.Append(errRes, fmt.Errorf("bad response status for %s; %s", k, resp.Status))
		}
		if errClose := resp.Body.Close(); errClose != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to close response body for %s; %w", k, err))
		}
		return errRes
	})
}

func serializeMetrics(ctx context.Context, mtype string, repo repository, wc writerCloser) error {
	var errRes error
	_ = repo.DropFn(ctx, func(k string, v []byte) (bool, error) {
		var (
			key        string
			typeToSend string
			drop       bool
		)
		switch {
		case strings.HasPrefix(k, gauge):
			key = strings.TrimPrefix(k, gauge)
			typeToSend = gauge
			// Drop this entry, because we will send it and have a new value every time.
			drop = true
		case strings.HasPrefix(k, counter):
			// Do not drop it because we need this value on the next pool call.
			key = strings.TrimPrefix(k, counter)
			typeToSend = counter
		default:
			return false, nil
		}

		if mtype != "" && mtype != typeToSend {
			return false, nil
		}

		metric, err := dataToMetrics(typeToSend, key, v)
		if err != nil {
			errRes = multierr.Append(errRes,
				fmt.Errorf("failed converting data to metric for %s(%s); %w", mtype, k, err))
			return false, nil
		}
		if err := transport.SerializeMetrics(&metric, wc); err != nil {
			errRes = multierr.Append(errRes, fmt.Errorf("failed to compress data for %s(%s); %w", mtype, k, err))
			return false, nil
		}
		return drop, nil
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
	Info(msg string, add ...string)
	Warn(msg string, add ...string)
	Error(msg string, add ...string)
	Debug(msg string, add ...string)
	Fatal(msg string, add ...string)
}

type repository interface {
	DropFn(ctx context.Context, fn func(k string, v []byte) (bool, error)) error
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
}

type writerCloser interface {
	Write(p []byte) (n int, err error)
	Close() error
}
