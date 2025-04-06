package poller

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/ASRafalsky/telemetry/cmd/agent/repository"
	"github.com/ASRafalsky/telemetry/internal/types"
)

func Poll(ctx context.Context, interval time.Duration, repos map[string]*repository.RepositoryUnit) {
	fmt.Printf("Polling started with interval %v\n", interval)
	pollTimer := time.NewTicker(interval)
	defer pollTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-pollTimer.C:
			for name := range repos {
				switch name {
				case repository.Gauge:
					getGaugeMetrics(repos[name])
				case repository.Counter:
					getCounterMetrics(repos[name])
				default:
				}
			}
		}
	}
}

func getCounterMetrics(repo *repository.RepositoryUnit) {
	repo.Mx.Lock()
	defer repo.Mx.Unlock()

	cnt, ok := repo.Get("PollCount")
	if !ok {
		repo.Set("PollCount", types.CounterToBytes(types.Counter(0)))
		return
	}
	cntToSet := types.BytesToCounter(cnt)
	cntToSet++
	repo.Set("PollCount", types.CounterToBytes(cntToSet))
}

func getGaugeMetrics(repo *repository.RepositoryUnit) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	repo.Mx.Lock()
	defer repo.Mx.Unlock()

	repo.Set("Alloc", types.GaugeToBytes(types.Gauge(memStats.Alloc)))
	repo.Set("BuckHashSys", types.GaugeToBytes(types.Gauge(memStats.BuckHashSys)))
	repo.Set("Frees", types.GaugeToBytes(types.Gauge(memStats.Frees)))
	repo.Set("GCCPUFraction", types.GaugeToBytes(types.Gauge(memStats.GCCPUFraction)))
	repo.Set("GCSys", types.GaugeToBytes(types.Gauge(memStats.GCSys)))
	repo.Set("HeapAlloc", types.GaugeToBytes(types.Gauge(memStats.HeapAlloc)))
	repo.Set("HeapIdle", types.GaugeToBytes(types.Gauge(memStats.HeapIdle)))
	repo.Set("HeapInuse", types.GaugeToBytes(types.Gauge(memStats.HeapInuse)))
	repo.Set("HeapObjects", types.GaugeToBytes(types.Gauge(memStats.HeapObjects)))
	repo.Set("HeapReleased", types.GaugeToBytes(types.Gauge(memStats.HeapReleased)))
	repo.Set("HeapSys", types.GaugeToBytes(types.Gauge(memStats.HeapSys)))
	repo.Set("LastGC", types.GaugeToBytes(types.Gauge(memStats.LastGC)))
	repo.Set("Lookups", types.GaugeToBytes(types.Gauge(memStats.Lookups)))
	repo.Set("MCacheInuse", types.GaugeToBytes(types.Gauge(memStats.MCacheInuse)))
	repo.Set("MCacheSys", types.GaugeToBytes(types.Gauge(memStats.MCacheSys)))
	repo.Set("MSpanInuse", types.GaugeToBytes(types.Gauge(memStats.MSpanInuse)))
	repo.Set("MSpanSys", types.GaugeToBytes(types.Gauge(memStats.MSpanSys)))
	repo.Set("Mallocs", types.GaugeToBytes(types.Gauge(memStats.Mallocs)))
	repo.Set("NextGC", types.GaugeToBytes(types.Gauge(memStats.NextGC)))
	repo.Set("NumForcedGC", types.GaugeToBytes(types.Gauge(memStats.NumForcedGC)))
	repo.Set("NumGC", types.GaugeToBytes(types.Gauge(memStats.NumGC)))
	repo.Set("OtherSys", types.GaugeToBytes(types.Gauge(memStats.OtherSys)))
	repo.Set("PauseTotalNs", types.GaugeToBytes(types.Gauge(memStats.PauseTotalNs)))
	repo.Set("StackInuse", types.GaugeToBytes(types.Gauge(memStats.StackInuse)))
	repo.Set("StackSys", types.GaugeToBytes(types.Gauge(memStats.StackSys)))
	repo.Set("Sys", types.GaugeToBytes(types.Gauge(memStats.Sys)))
	repo.Set("TotalAlloc", types.GaugeToBytes(types.Gauge(memStats.TotalAlloc)))
	repo.Set("RandomValue", types.GaugeToBytes(types.Gauge(rand.Float64())))
}
