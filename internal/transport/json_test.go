package transport

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const cnt = 100

func TestSerializeMetrics(t *testing.T) {
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
	t.Run("custom serialize method", func(t *testing.T) {
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
	})

	t.Run("json serialize method", func(t *testing.T) {
		jsonMetrics, err := json.Marshal(metricsList)
		require.NoError(t, err)

		outputMetrics, err := DeserializeMetrics(jsonMetrics)
		require.NoError(t, err)
		require.Len(t, outputMetrics, cnt)

		for i := range outputMetrics {
			require.Equal(t, metricsList[i].ID, outputMetrics[i].ID)
			require.Equal(t, metricsList[i].MType, outputMetrics[i].MType)
			require.Equal(t, *metricsList[i].Value, *outputMetrics[i].Value)
			require.Equal(t, *metricsList[i].Delta, *outputMetrics[i].Delta)
		}
	})
}

func BenchmarkCustomConvertation(b *testing.B) {
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
	require.Len(b, metricsList, cnt)
	var (
		outputMetrics []Metrics
		err           error
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(nil)
		for _, m := range metricsList {
			err = SerializeMetrics(&m, buf)
			if err != nil {
				b.Fatal(err)
			}
			outputMetrics, err = DeserializeMetrics(buf.Bytes())
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	b.StopTimer()
	require.Len(b, outputMetrics, cnt)
}

func BenchmarkJSONConvertation(b *testing.B) {
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
	require.Len(b, metricsList, cnt)
	var (
		outputMetrics []Metrics
		buf           []byte
		err           error
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err = json.Marshal(metricsList)
		if err != nil {
			b.Fatal(err)
		}
		err = json.Unmarshal(buf, &outputMetrics)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	require.Len(b, outputMetrics, cnt)
}

func BenchmarkComlexConvertation(b *testing.B) {
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
	require.Len(b, metricsList, cnt)
	var (
		outputMetrics []Metrics
		buf           []byte
		err           error
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err = json.Marshal(metricsList)
		if err != nil {
			b.Fatal(err)
		}
		outputMetrics, err = DeserializeMetrics(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	require.Len(b, outputMetrics, cnt)
}
