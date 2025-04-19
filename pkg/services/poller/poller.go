package poller

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

func Poll(ctx context.Context, fn func(r repository), interval time.Duration, repo repository, log logger) {
	log.Info("Polling started with interval:", interval.String())
	pollTimer := time.NewTicker(interval)
	defer pollTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-pollTimer.C:
			fn(repo)
		}
	}
}

func GetCounterMetrics(repo repository) {
	name := counter + "PollCount"
	cnt, ok := repo.Get(name)
	if !ok {
		repo.Set(name, types.CounterToBytes(types.Counter(0)))
		return
	}
	cntToSet := types.BytesToCounter(cnt)
	cntToSet++
	repo.Set(name, types.CounterToBytes(cntToSet))
}

func GetGaugeMetrics(repo repository) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	repo.Set(gauge+"Alloc", types.GaugeToBytes(types.Gauge(memStats.Alloc)))
	repo.Set(gauge+"BuckHashSys", types.GaugeToBytes(types.Gauge(memStats.BuckHashSys)))
	repo.Set(gauge+"Frees", types.GaugeToBytes(types.Gauge(memStats.Frees)))
	repo.Set(gauge+"GCCPUFraction", types.GaugeToBytes(types.Gauge(memStats.GCCPUFraction)))
	repo.Set(gauge+"GCSys", types.GaugeToBytes(types.Gauge(memStats.GCSys)))
	repo.Set(gauge+"HeapAlloc", types.GaugeToBytes(types.Gauge(memStats.HeapAlloc)))
	repo.Set(gauge+"HeapIdle", types.GaugeToBytes(types.Gauge(memStats.HeapIdle)))
	repo.Set(gauge+"HeapInuse", types.GaugeToBytes(types.Gauge(memStats.HeapInuse)))
	repo.Set(gauge+"HeapObjects", types.GaugeToBytes(types.Gauge(memStats.HeapObjects)))
	repo.Set(gauge+"HeapReleased", types.GaugeToBytes(types.Gauge(memStats.HeapReleased)))
	repo.Set(gauge+"HeapSys", types.GaugeToBytes(types.Gauge(memStats.HeapSys)))
	repo.Set(gauge+"LastGC", types.GaugeToBytes(types.Gauge(memStats.LastGC)))
	repo.Set(gauge+"Lookups", types.GaugeToBytes(types.Gauge(memStats.Lookups)))
	repo.Set(gauge+"MCacheInuse", types.GaugeToBytes(types.Gauge(memStats.MCacheInuse)))
	repo.Set(gauge+"MCacheSys", types.GaugeToBytes(types.Gauge(memStats.MCacheSys)))
	repo.Set(gauge+"MSpanInuse", types.GaugeToBytes(types.Gauge(memStats.MSpanInuse)))
	repo.Set(gauge+"MSpanSys", types.GaugeToBytes(types.Gauge(memStats.MSpanSys)))
	repo.Set(gauge+"Mallocs", types.GaugeToBytes(types.Gauge(memStats.Mallocs)))
	repo.Set(gauge+"NextGC", types.GaugeToBytes(types.Gauge(memStats.NextGC)))
	repo.Set(gauge+"NumForcedGC", types.GaugeToBytes(types.Gauge(memStats.NumForcedGC)))
	repo.Set(gauge+"NumGC", types.GaugeToBytes(types.Gauge(memStats.NumGC)))
	repo.Set(gauge+"OtherSys", types.GaugeToBytes(types.Gauge(memStats.OtherSys)))
	repo.Set(gauge+"PauseTotalNs", types.GaugeToBytes(types.Gauge(memStats.PauseTotalNs)))
	repo.Set(gauge+"StackInuse", types.GaugeToBytes(types.Gauge(memStats.StackInuse)))
	repo.Set(gauge+"StackSys", types.GaugeToBytes(types.Gauge(memStats.StackSys)))
	repo.Set(gauge+"Sys", types.GaugeToBytes(types.Gauge(memStats.Sys)))
	repo.Set(gauge+"TotalAlloc", types.GaugeToBytes(types.Gauge(memStats.TotalAlloc)))
	repo.Set(gauge+"RandomValue", types.GaugeToBytes(types.Gauge(rand.Float64())))
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
	Delete(k string)
}
