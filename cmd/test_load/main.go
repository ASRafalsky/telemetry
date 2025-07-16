package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/mailru/easyjson"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/pkg/cache"
	lllog "github.com/ASRafalsky/telemetry/pkg/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	repo := cache.New[string, []byte]()

	if err := poller.GetGaugeMetrics(ctx, repo); err != nil {
		log.Fatal(err)
	}
	if err := poller.GetCounterMetrics(ctx, repo); err != nil {
		log.Fatal(err)
	}
	if err := poller.GetPSMemMetrics(ctx, repo); err != nil {
		log.Fatal(err)
	}
	if err := poller.GetPSCPUMetrics(ctx, repo); err != nil {
		log.Fatal(err)
	}

	logger, err := lllog.AddLoggerWith("error", "")
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg := newSenderCfg(config.Agent{
		CommonFields: config.CommonFields{
			Addr: config.DefaultAddr,
			Key:  "secret-key",
		},
		ReportPeriod: 50,
		RateLimit:    0,
	})
	var cnt atomic.Uint64
	timeout := 1 * time.Second
	for range 20 {
		if ctx.Err() != nil {
			return
		}
		go func() {
			client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout), httpclient.WithRetryCount(1))
			fmt.Println(client)
			var clientWg sync.WaitGroup
			for range 100 {
				if ctx.Err() != nil {
					return
				}
				clientWg.Add(1)
				go func() {
					defer clientWg.Done()
					reporter.Send(ctx, "", cfg, client, repo, logger)
				}()
				clientWg.Add(1)
				go func() {
					defer clientWg.Done()
					tc := time.NewTicker(cfg.Interval)
					defer tc.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-tc.C:
							if err := getValues(ctx, cfg, client, repo); err != nil {
								logger.Error("Failed to get values", err.Error())
							}
							cnt.Add(1)
						default:
						}
					}
				}()
				clientWg.Wait()
			}
		}()
	}

	time.Sleep(10 * time.Second)

	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	resp, err := client.Post("http://"+config.DefaultDiagAddr+"/profile/", nil, header)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}

	for {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println(cnt.Load())
		}
	}
}

func newSenderCfg(cfg config.Agent) reporter.Config {
	return reporter.Config{
		Interval:  time.Duration(cfg.ReportPeriod) * time.Millisecond,
		Address:   "http://" + cfg.Addr,
		Key:       cfg.Key,
		RateLimit: cfg.RateLimit,
	}
}

func getValues(
	ctx context.Context,
	cfg reporter.Config,
	client *httpclient.Client,
	repo *cache.MemStorage[string, []byte],
) error {
	return repo.ForEach(ctx, func(k string, _ []byte) error {
		m, err := keyToMetrics(k)
		if err != nil {
			return err
		}
		buf, err := easyjson.Marshal(m)
		if err != nil {
			return err
		}
		bufToSend := bytes.NewBuffer(nil)
		zr := gzip.NewWriter(bufToSend)
		_, err = zr.Write(buf)
		if err != nil {
			return err
		}
		if err := zr.Close(); err != nil {
			return err
		}
		header := http.Header{
			"Content-Type":     []string{"application/json"},
			"Content-Encoding": []string{"gzip"},
		}

		h := hmac.New(sha256.New, []byte(cfg.Key))
		if _, err = h.Write(bufToSend.Bytes()); err != nil {
			return err
		}
		header.Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))
		resp, err := client.Post(cfg.Address+"/value/", bufToSend, header)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(resp.Status)
		}
		return nil
	})
}

func keyToMetrics(key string) (transport.Metrics, error) {
	const (
		gauge   = "gauge"
		counter = "counter"
	)
	switch {
	case strings.HasPrefix(key, gauge):
		key = strings.TrimPrefix(key, gauge)
		return transport.Metrics{
			ID:    key,
			MType: gauge,
		}, nil
	case strings.HasPrefix(key, counter):
		key = strings.TrimPrefix(key, counter)
		return transport.Metrics{
			ID:    key,
			MType: counter,
		}, nil
	default:
		return transport.Metrics{}, errors.New("unknown type")
	}
}
