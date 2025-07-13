package types

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGauge(t *testing.T) {

	tt := []struct {
		stringVal string
		expected  Gauge
		err       error
	}{
		{
			stringVal: "0",
			expected:  Gauge(0.0),
			err:       nil,
		},
		{
			stringVal: "123.456",
			expected:  Gauge(123.456),
			err:       nil,
		},
		{
			stringVal: "135",
			expected:  Gauge(135),
			err:       nil,
		},
		{
			stringVal: "NaN",
			expected:  Gauge(math.NaN()),
			err:       nil,
		},
		{
			stringVal: "lol",
			err:       errors.New("invalid syntax"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.stringVal, func(t *testing.T) {
			testVal, err := ParseGauge(tc.stringVal)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			if !math.IsNaN(float64(testVal)) {
				require.Equal(t, tc.expected, testVal)

				bytesValue := GaugeToBytes(testVal)
				require.Equal(t, tc.expected, BytesToGauge(bytesValue))
			}
			require.Equal(t, tc.stringVal, testVal.String())
		})
	}
}

func TestCounter(t *testing.T) {

	tt := []struct {
		stringVal string
		expected  Counter
		err       error
	}{
		{
			stringVal: "0",
			expected:  Counter(0),
			err:       nil,
		},
		{
			stringVal: "123.456",
			err:       errors.New("invalid syntax"),
		},
		{
			stringVal: "135",
			expected:  Counter(135),
			err:       nil,
		},
		{
			stringVal: "9223372036854775807",
			expected:  Counter(math.MaxInt64),
			err:       nil,
		},
		{
			stringVal: "lol",
			err:       errors.New("invalid syntax"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.stringVal, func(t *testing.T) {
			testVal, err := ParseCounter(tc.stringVal)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				return
			}
			bytesValue := CounterToBytes(testVal)
			require.Equal(t, tc.expected, BytesToCounter(bytesValue))

			require.NoError(t, err)
			require.Equal(t, tc.expected, testVal)
			require.Equal(t, tc.stringVal, testVal.String())
		})
	}
}

func TestBytesToTypeWithBadData(t *testing.T) {
	badData := [][]byte{
		nil,
		{},
		{1, 2, 3},
	}

	for _, data := range badData {
		assert.NotPanics(t, func() {
			BytesToGauge(data)
		})
		assert.NotPanics(t, func() {
			BytesToCounter(data)
		})
	}
}

const cnt = 1000

func BenchmarkGaugeToBytes(b *testing.B) {
	testVal := make([]Gauge, cnt)
	for range 100 {
		testVal = append(testVal, Gauge(rand.ExpFloat64()))
	}
	var result []byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for n := range cnt {
			result = GaugeToBytes(testVal[n])
		}
	}
	b.StopTimer()
	require.NotNil(b, result)
}

func BenchmarkCounterToBytes(b *testing.B) {
	testVal := make([]Counter, cnt)
	for range 100 {
		testVal = append(testVal, Counter(rand.Int64()))
	}
	var result []byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for n := range cnt {
			result = CounterToBytes(testVal[n])
		}
	}
	b.StopTimer()
	require.NotNil(b, result)
}
