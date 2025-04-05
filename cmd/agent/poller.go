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
	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

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

type Repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Keys() []string
	Size() int
	Delete(k string)
}

const url = "http://localhost:8080"

func sendCounterData(ctx context.Context, repo Repository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := url + "/update/counter/" + k + "/" + internal.BytesToCounter(v).String()
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

func sendGaugeData(ctx context.Context, repo Repository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	err := repo.ForEach(ctx, func(k string, v []byte) error {
		urlData := url + "/update/counter/" + k + "/" + internal.BytesToGauge(v).String()
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

func getCounterMetrics(repo Repository) {
	cnt, ok := repo.Get("PollCount")
	if !ok {
		repo.Set("PollCount", internal.CounterToBytes(internal.Counter(1)))
		return
	}
	cntToSet := internal.BytesToCounter(cnt)
	cntToSet++
	repo.Set("PollCount", internal.CounterToBytes(cntToSet))
}

func getGaugeMetrics(repo Repository) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	repo.Set("Alloc", internal.GaugeToBytes(internal.Gauge(memStats.Alloc)))
	repo.Set("BuckHashSys", internal.GaugeToBytes(internal.Gauge(memStats.BuckHashSys)))
	repo.Set("Frees", internal.GaugeToBytes(internal.Gauge(memStats.Frees)))
	repo.Set("GCCPUFraction", internal.GaugeToBytes(internal.Gauge(memStats.GCCPUFraction)))
	repo.Set("GCSys", internal.GaugeToBytes(internal.Gauge(memStats.GCSys)))
	repo.Set("HeapAlloc", internal.GaugeToBytes(internal.Gauge(memStats.HeapAlloc)))
	repo.Set("HeapIdle", internal.GaugeToBytes(internal.Gauge(memStats.HeapIdle)))
	repo.Set("HeapInuse", internal.GaugeToBytes(internal.Gauge(memStats.HeapInuse)))
	repo.Set("HeapObjects", internal.GaugeToBytes(internal.Gauge(memStats.HeapObjects)))
	repo.Set("HeapReleased", internal.GaugeToBytes(internal.Gauge(memStats.HeapReleased)))
	repo.Set("HeapSys", internal.GaugeToBytes(internal.Gauge(memStats.HeapSys)))
	repo.Set("LastGC", internal.GaugeToBytes(internal.Gauge(memStats.LastGC)))
	repo.Set("Lookups", internal.GaugeToBytes(internal.Gauge(memStats.Lookups)))
	repo.Set("MCacheInuse", internal.GaugeToBytes(internal.Gauge(memStats.MCacheInuse)))
	repo.Set("MCacheSys", internal.GaugeToBytes(internal.Gauge(memStats.MCacheSys)))
	repo.Set("MSpanInuse", internal.GaugeToBytes(internal.Gauge(memStats.MSpanInuse)))
	repo.Set("MSpanSys", internal.GaugeToBytes(internal.Gauge(memStats.MSpanSys)))
	repo.Set("Mallocs", internal.GaugeToBytes(internal.Gauge(memStats.Mallocs)))
	repo.Set("NextGC", internal.GaugeToBytes(internal.Gauge(memStats.NextGC)))
	repo.Set("NumForcedGC", internal.GaugeToBytes(internal.Gauge(memStats.NumForcedGC)))
	repo.Set("NumGC", internal.GaugeToBytes(internal.Gauge(memStats.NumGC)))
	repo.Set("OtherSys", internal.GaugeToBytes(internal.Gauge(memStats.OtherSys)))
	repo.Set("PauseTotalNs", internal.GaugeToBytes(internal.Gauge(memStats.PauseTotalNs)))
	repo.Set("StackInuse", internal.GaugeToBytes(internal.Gauge(memStats.StackInuse)))
	repo.Set("StackSys", internal.GaugeToBytes(internal.Gauge(memStats.StackSys)))
	repo.Set("Sys", internal.GaugeToBytes(internal.Gauge(memStats.Sys)))
	repo.Set("TotalAlloc", internal.GaugeToBytes(internal.Gauge(memStats.TotalAlloc)))
	repo.Set("RandomValue", internal.GaugeToBytes(internal.Gauge(rand.Float64())))
}
