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
			getCounterMetrics(counterRepo)
			value, ok := counterRepo.Get("PollCount")
			assert.True(t, ok)
			assert.Equal(t, types.Counter(i), types.BytesToCounter(value), i)
		}
	})

	t.Run("getGaugeMetrics", func(t *testing.T) {
		var previousValue types.Gauge
		for range 10 {
			getGaugeMetrics(gaugeRepo)
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
	repos := map[string]Repository{
		gauge:   gaugeRepo,
		counter: counterRepo,
	}
	ctx, cancel := context.WithCancel(context.Background())

	log, err := log.AddLoggerWith("info", "")
	require.NoError(t, err)
	defer log.Sync()
	go Poll(ctx, 100*time.Millisecond, repos, log)

	// Wait 90 ms, it is too early to have any data.
	time.Sleep(90 * time.Millisecond)
	for name := range repos {
		require.Zero(t, repos[name].Size())
	}

	// Next 20 ms we should have full  repositories.
	require.Eventually(t,
		func() bool {
			for name := range repos {
				if repos[name].Size() == 0 {
					return false
				}
			}
			return true
		},
		20*time.Millisecond, 5*time.Millisecond)

	cancel()
}
