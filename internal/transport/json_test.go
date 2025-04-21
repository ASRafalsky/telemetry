package transport

import (
	"bytes"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerializeMetrics(t *testing.T) {
	const cnt = 100

	metricsList := make([]Metrics, cnt)
	for i := range cnt {
		var (
			val   = rand.Float64()
			delta = rand.Int64()
		)
		metricsList[i] = Metrics{
			MType: "MType" + strconv.Itoa(i),
			ID:    "ID" + strconv.Itoa(i),
			Value: &val,
			Delta: &delta,
		}
	}
	require.Len(t, metricsList, cnt)

	buf := bytes.NewBuffer(nil)
	for _, m := range metricsList {
		require.NoError(t, SerializeMetrics(&m, buf))
	}

	outputMetrics, err := DeserializeMetrics(buf.Bytes())
	require.NoError(t, err)
	require.Len(t, outputMetrics, cnt)

	for i := range outputMetrics {
		require.Equal(t, metricsList[i].ID, outputMetrics[i].ID)
		require.Equal(t, metricsList[i].MType, outputMetrics[i].MType)
		require.Equal(t, *metricsList[i].Value, *outputMetrics[i].Value)
		require.Equal(t, *metricsList[i].Delta, *outputMetrics[i].Delta)
	}
}
