package poller

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/internal/types"
)

func TestGetMetrics(t *testing.T) {
	repo := storage.New[string, []byte]()

	t.Run("getCounterMetrics", func(t *testing.T) {
		for i := range 10 {
			GetCounterMetrics(repo)
			value, ok := repo.Get(counter + "PollCount")
			assert.True(t, ok)
			assert.Equal(t, types.Counter(i), types.BytesToCounter(value), i)
		}
	})

	t.Run("getGaugeMetrics", func(t *testing.T) {
		var previousValue types.Gauge
		for range 10 {
			GetGaugeMetrics(repo)
			value, ok := repo.Get(gauge + "RandomValue")
			assert.True(t, ok)
			gaugeValue := types.BytesToGauge(value)
			assert.NotZero(t, gaugeValue)
			assert.NotEqual(t, previousValue, gaugeValue)
			previousValue = gaugeValue
		}
	})
}

func TestPoll(t *testing.T) {
	repo := storage.New[string, []byte]()
	ctx, cancel := context.WithCancel(context.Background())

	log, err := log.AddLoggerWith("info", "")
	require.NoError(t, err)
	defer log.Sync()
	go Poll(ctx, GetGaugeMetrics, 100*time.Millisecond, repo, log)
	go Poll(ctx, GetCounterMetrics, 100*time.Millisecond, repo, log)

	// Wait 90 ms, it is too early to have any data.
	time.Sleep(90 * time.Millisecond)
	require.Zero(t, repo.Size())

	// Next 20 ms we should have full  repositories.
	require.Eventually(t, func() bool {
		var (
			gaugeFound   bool
			counterFound bool
		)
		repo.ForEach(context.Background(), func(k string, v []byte) error {
			if strings.HasPrefix(k, gauge) {
				gaugeFound = true
			}
			if strings.HasPrefix(k, counter) {
				counterFound = true
			}
			return nil
		})
		return gaugeFound && counterFound
	},
		20*time.Millisecond, 5*time.Millisecond)

	cancel()
}
