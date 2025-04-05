package internal

import (
	"errors"
	"math"
	"testing"

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
			require.NoError(t, err)
			require.Equal(t, tc.expected, testVal)
			require.Equal(t, tc.stringVal, testVal.String())
		})
	}
}
