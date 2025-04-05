package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"

	"github.com/ASRafalsky/telemetry/internal"
	"github.com/ASRafalsky/telemetry/internal/storage"
)

func poller(ctx context.Context, pollInterval, sendInterval time.Duration, client *httpclient.Client) {
	gaugeRepo := storage.New[string, internal.Gauge]()
	counterRepo := storage.New[string, internal.Counter]()

	pollTimer := time.NewTicker(pollInterval)
	defer pollTimer.Stop()
	sendTimer := time.NewTicker(sendInterval)
	defer sendTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-pollTimer.C:
			getGaugeMetrics(gaugeRepo)
			getCounterMetrics(counterRepo)
		case <-sendTimer.C:
			sendCounterData(ctx, counterRepo, client)
			sendGaugeData(ctx, gaugeRepo, client)
		}
	}
}

type CommonRepository interface {
	Size() int
	Delete(k string)
}

type GaugeRepository interface {
	CommonRepository
	Set(k string, v internal.Gauge)
	Get(k string) (internal.Gauge, bool)
	ForEach(ctx context.Context, fn func(k string, v internal.Gauge) error) error
}

type CounterRepository interface {
	CommonRepository
	Set(k string, v internal.Counter)
	Get(k string) (internal.Counter, bool)
	ForEach(ctx context.Context, fn func(k string, v internal.Counter) error) error
}

const url = "http://localhost:8080"

func sendCounterData(ctx context.Context, repo CounterRepository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	err := repo.ForEach(ctx, func(k string, v internal.Counter) error {
		urlData := url + "/update/counter/" + k + "/" + v.String()
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

func sendGaugeData(ctx context.Context, repo GaugeRepository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	err := repo.ForEach(ctx, func(k string, v internal.Gauge) error {
		urlData := url + "/update/counter/" + k + "/" + v.String()
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

func getCounterMetrics(repo CounterRepository) {
	cnt, ok := repo.Get("PollCount")
	if !ok {
		repo.Set("PollCount", 1)
		return
	}
	cnt++
	repo.Set("PollCount", cnt)
}

func getGaugeMetrics(repo GaugeRepository) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	repo.Set("Alloc", internal.Gauge(memStats.Alloc))
	repo.Set("BuckHashSys", internal.Gauge(memStats.BuckHashSys))
	repo.Set("Frees", internal.Gauge(memStats.Frees))
	repo.Set("GCCPUFraction", internal.Gauge(memStats.GCCPUFraction))
	repo.Set("GCSys", internal.Gauge(memStats.GCSys))
	repo.Set("HeapAlloc", internal.Gauge(memStats.HeapAlloc))
	repo.Set("HeapIdle", internal.Gauge(memStats.HeapIdle))
	repo.Set("HeapInuse", internal.Gauge(memStats.HeapInuse))
	repo.Set("HeapObjects", internal.Gauge(memStats.HeapObjects))
	repo.Set("HeapReleased", internal.Gauge(memStats.HeapReleased))
	repo.Set("HeapSys", internal.Gauge(memStats.HeapSys))
	repo.Set("LastGC", internal.Gauge(memStats.LastGC))
	repo.Set("Lookups", internal.Gauge(memStats.Lookups))
	repo.Set("MCacheInuse", internal.Gauge(memStats.MCacheInuse))
	repo.Set("MCacheSys", internal.Gauge(memStats.MCacheSys))
	repo.Set("MSpanInuse", internal.Gauge(memStats.MSpanInuse))
	repo.Set("MSpanSys", internal.Gauge(memStats.MSpanSys))
	repo.Set("Mallocs", internal.Gauge(memStats.Mallocs))
	repo.Set("NextGC", internal.Gauge(memStats.NextGC))
	repo.Set("NumForcedGC", internal.Gauge(memStats.NumForcedGC))
	repo.Set("NumGC", internal.Gauge(memStats.NumGC))
	repo.Set("OtherSys", internal.Gauge(memStats.OtherSys))
	repo.Set("PauseTotalNs", internal.Gauge(memStats.PauseTotalNs))
	repo.Set("StackInuse", internal.Gauge(memStats.StackInuse))
	repo.Set("StackSys", internal.Gauge(memStats.StackSys))
	repo.Set("Sys", internal.Gauge(memStats.Sys))
	repo.Set("TotalAlloc", internal.Gauge(memStats.TotalAlloc))
	repo.Set("RandomValue", internal.Gauge(rand.Float64()))
}
