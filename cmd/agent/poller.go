package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

type (
	gauge   float64
	counter int64
)

func (g gauge) String() string {
	return strconv.FormatFloat(float64(g), 'g', -1, 64)
}

func (c counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func poller(ctx context.Context, pollInterval, sendInterval time.Duration, client *httpclient.Client) {
	gaugeRepo := storage.New[string, gauge]()
	counterRepo := storage.New[string, counter]()

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
	Set(k string, v gauge)
	Get(k string) (gauge, bool)
	ForEach(ctx context.Context, fn func(k string, v gauge) error) error
}

type CounterRepository interface {
	CommonRepository
	Set(k string, v counter)
	Get(k string) (counter, bool)
	ForEach(ctx context.Context, fn func(k string, v counter) error) error
}

const url = "http://localhost:8080"

func sendCounterData(ctx context.Context, repo CounterRepository, client *httpclient.Client) {
	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	err := repo.ForEach(ctx, func(k string, v counter) error {
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
	err := repo.ForEach(ctx, func(k string, v gauge) error {
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

	repo.Set("Alloc", gauge(memStats.Alloc))
	repo.Set("BuckHashSys", gauge(memStats.BuckHashSys))
	repo.Set("Frees", gauge(memStats.Frees))
	repo.Set("GCCPUFraction", gauge(memStats.GCCPUFraction))
	repo.Set("GCSys", gauge(memStats.GCSys))
	repo.Set("HeapAlloc", gauge(memStats.HeapAlloc))
	repo.Set("HeapIdle", gauge(memStats.HeapIdle))
	repo.Set("HeapInuse", gauge(memStats.HeapInuse))
	repo.Set("HeapObjects", gauge(memStats.HeapObjects))
	repo.Set("HeapReleased", gauge(memStats.HeapReleased))
	repo.Set("HeapSys", gauge(memStats.HeapSys))
	repo.Set("LastGC", gauge(memStats.LastGC))
	repo.Set("Lookups", gauge(memStats.Lookups))
	repo.Set("MCacheInuse", gauge(memStats.MCacheInuse))
	repo.Set("MCacheSys", gauge(memStats.MCacheSys))
	repo.Set("MSpanInuse", gauge(memStats.MSpanInuse))
	repo.Set("MSpanSys", gauge(memStats.MSpanSys))
	repo.Set("Mallocs", gauge(memStats.Mallocs))
	repo.Set("NextGC", gauge(memStats.NextGC))
	repo.Set("NumForcedGC", gauge(memStats.NumForcedGC))
	repo.Set("NumGC", gauge(memStats.NumGC))
	repo.Set("OtherSys", gauge(memStats.OtherSys))
	repo.Set("PauseTotalNs", gauge(memStats.PauseTotalNs))
	repo.Set("StackInuse", gauge(memStats.StackInuse))
	repo.Set("StackSys", gauge(memStats.StackSys))
	repo.Set("Sys", gauge(memStats.Sys))
	repo.Set("TotalAlloc", gauge(memStats.TotalAlloc))
	repo.Set("RandomValue", gauge(rand.Float64()))
}
