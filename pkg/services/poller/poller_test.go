package poller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/internal/types"
)

func TestGetMetrics(t *testing.T) {
	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	t.Run("getCounterMetrics", func(t *testing.T) {
		for i := range 10 {
			GetCounterMetrics(counterRepo)
			value, ok := counterRepo.Get("PollCount")
			assert.True(t, ok)
			assert.Equal(t, types.Counter(i), types.BytesToCounter(value), i)
		}
	})

	t.Run("getGaugeMetrics", func(t *testing.T) {
		var previousValue types.Gauge
		for range 10 {
			GetGaugeMetrics(gaugeRepo)
			value, ok := gaugeRepo.Get("RandomValue")
			assert.True(t, ok)
			gaugeValue := types.BytesToGauge(value)
			assert.NotZero(t, gaugeValue)
			assert.NotEqual(t, previousValue, gaugeValue)
			previousValue = gaugeValue
		}
	})
}

func TestPoll(t *testing.T) {
	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()
	ctx, cancel := context.WithCancel(context.Background())

	log, err := log.AddLoggerWith("info", "")
	require.NoError(t, err)
	defer log.Sync()
	go Poll(ctx, GetGaugeMetrics, 100*time.Millisecond, gaugeRepo, log)
	go Poll(ctx, GetCounterMetrics, 100*time.Millisecond, counterRepo, log)

	// Wait 90 ms, it is too early to have any data.
	time.Sleep(90 * time.Millisecond)
	require.Zero(t, gaugeRepo.Size())
	require.Zero(t, counterRepo.Size())

	// Next 20 ms we should have full  repositories.
	require.Eventually(t, func() bool { return gaugeRepo.Size() > 0 && counterRepo.Size() > 0 },
		20*time.Millisecond, 5*time.Millisecond)

	cancel()
}
