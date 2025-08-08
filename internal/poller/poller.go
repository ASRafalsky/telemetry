// Package poller... A-a-a-a-a-a! Stop it!
package poller

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

// Config of poller.
type Config struct {
	Interval time.Duration
}

// Poll polls everything what you need. You really need it, believe me!
func Poll(ctx context.Context,
	fn func(ctx context.Context, r repository) error, cfg Config, repo repository, log logger) {
	log.Info("Polling started with interval:", cfg.Interval.String())
	pollTimer := time.NewTicker(cfg.Interval)
	defer pollTimer.Stop()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-pollTimer.C:
			if err := fn(ctx, repo); err != nil {
				log.Error("Polling failed with err.", err.Error())
			}
		}
	}
}

// GetCounterMetrics collects counter metrics and saves them to the repo.
func GetCounterMetrics(_ context.Context, repo repository) error {
	name := counter + "PollCount"
	cnt, ok := repo.Get(name)
	if !ok {
		repo.Set(name, types.CounterToBytes(types.Counter(0)))
		return nil
	}
	cntToSet := types.BytesToCounter(cnt)
	cntToSet++
	repo.Set(name, types.CounterToBytes(cntToSet))
	return nil
}

// GetGaugeMetrics collects gauge metrics and saves them to the repo.
func GetGaugeMetrics(_ context.Context, repo repository) error {
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
	return nil
}

// GetPSMemMetrics collects memory metrics and saves them to the repo.
func GetPSMemMetrics(_ context.Context, repo repository) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	repo.Set(gauge+"FreeMemory", types.GaugeToBytes(types.Gauge(v.Free)))
	repo.Set(gauge+"TotalMemory", types.GaugeToBytes(types.Gauge(v.Total)))
	return nil
}

// GetPSCPUMetrics collects cpu metrics and saves them to the repo.
func GetPSCPUMetrics(ctx context.Context, repo repository) error {
	cpu, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return err
	}
	for n := range cpu {
		repo.Set(gauge+"CPUutilization"+strconv.Itoa(n+1), types.GaugeToBytes(types.Gauge(cpu[n])))
	}
	return nil
}

type logger interface {
	Info(msg string, add ...string)
	Warn(msg string, add ...string)
	Error(msg string, add ...string)
	Debug(msg string, add ...string)
	Fatal(msg string, add ...string)
}

type repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
}
