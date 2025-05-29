package poller

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/cache"
	"github.com/ASRafalsky/telemetry/internal/types"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

func TestGetMetrics(t *testing.T) {
	repo := cache.New[string, []byte]()

	t.Run("getCounterMetrics", func(t *testing.T) {
		for i := range 10 {
			require.NoError(t, GetCounterMetrics(context.Background(), repo))
			value, ok := repo.Get(counter + "PollCount")
			assert.True(t, ok)
			assert.Equal(t, types.Counter(i), types.BytesToCounter(value), i)
		}
	})

	t.Run("getGaugeMetrics", func(t *testing.T) {
		var previousValue types.Gauge
		for range 10 {
			require.NoError(t, GetGaugeMetrics(context.Background(), repo))
			value, ok := repo.Get(gauge + "RandomValue")
			assert.True(t, ok)
			gaugeValue := types.BytesToGauge(value)
			assert.NotZero(t, gaugeValue)
			assert.NotEqual(t, previousValue, gaugeValue)
			previousValue = gaugeValue
		}
	})

	t.Run("getPSMemMetrics", func(t *testing.T) {
		for range 10 {
			require.NoError(t, GetPSMemMetrics(context.Background(), repo))
			value, ok := repo.Get(gauge + "FreeMemory")
			assert.True(t, ok)
			gaugeValue := types.BytesToGauge(value)
			assert.NotZero(t, gaugeValue)
		}
	})

	t.Run("getPSCPUMetrics", func(t *testing.T) {
		numCPU := runtime.NumCPU()
		for range 3 {
			require.NoError(t, GetPSCPUMetrics(context.Background(), repo))
			for i := range numCPU {
				_, ok := repo.Get(gauge + "CPUutilization" + strconv.Itoa(i+1))
				assert.True(t, ok)
			}
		}
	})
}

func TestPoll(t *testing.T) {
	repo := cache.New[string, []byte]()
	ctx, cancel := context.WithCancel(context.Background())

	log, err := log.AddLoggerWith("info", "")
	require.NoError(t, err)

	cfg := Config{
		Interval: time.Duration(100) * time.Millisecond,
	}

	go Poll(ctx, GetGaugeMetrics, cfg, repo, log)
	go Poll(ctx, GetCounterMetrics, cfg, repo, log)
	go Poll(ctx, GetPSMemMetrics, cfg, repo, log)
	go Poll(ctx, GetPSCPUMetrics, cfg, repo, log)

	// Wait 90 ms, it is too early to have any data.
	time.Sleep(90 * time.Millisecond)
	require.Zero(t, repo.Size())

	// Next 90 ms we should have full  repositories.
	require.Eventually(t, func() bool {
		var (
			gaugeFound   bool
			counterFound bool
			psMemFound   bool
			psCPUcnt     int
		)
		require.NoError(t, repo.ForEach(context.Background(), func(k string, v []byte) error {
			if strings.HasPrefix(k, gauge) {
				if strings.Contains(k, "Memory") {
					psMemFound = true
				}
				if strings.Contains(k, "CPUutilization") {
					psCPUcnt++
				}
				gaugeFound = true
			}
			if strings.HasPrefix(k, counter) {
				counterFound = true
			}
			return nil
		}))
		return gaugeFound && counterFound && psMemFound && psCPUcnt == runtime.NumCPU()
	},
		20*time.Millisecond, 5*time.Millisecond)
	cancel()
}
