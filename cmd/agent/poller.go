package main

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/ASRafalsky/telemetry/internal"
)

func poll(ctx context.Context, pollInterval time.Duration, repos map[string]*repositoryUnit) {

	pollTimer := time.NewTicker(pollInterval)
	defer pollTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-pollTimer.C:
			for name := range repos {
				switch name {
				case gauge:
					getGaugeMetrics(repos[name])
				case counter:
					getCounterMetrics(repos[name])
				default:
				}
			}
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

func getCounterMetrics(repo *repositoryUnit) {
	repo.mx.Lock()
	defer repo.mx.Unlock()

	cnt, ok := repo.Get("PollCount")
	if !ok {
		repo.Set("PollCount", internal.CounterToBytes(internal.Counter(1)))
		return
	}
	cntToSet := internal.BytesToCounter(cnt)
	cntToSet++
	repo.Set("PollCount", internal.CounterToBytes(cntToSet))
}

func getGaugeMetrics(repo *repositoryUnit) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	repo.mx.Lock()
	defer repo.mx.Unlock()

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
