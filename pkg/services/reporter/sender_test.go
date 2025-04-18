package reporter

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

const (
	testValStr   = "1234"
	testValInt64 = int64(1234)
	testValFloat = float64(1234)
)

func TestSend(t *testing.T) {
	var (
		gFound, cFound, gJSONFound, cJSONFound bool
	)

	// Add handlers and router.
	gaugeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Contains(t, []string{"gauge_var1", "gauge_var2"}, chi.URLParam(r, "name"))
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			value := chi.URLParam(r, "value")
			require.Equal(t, testValStr, value)
			gFound = true
		}
	}
	counterHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Contains(t, []string{"counter_var1", "counter_var2"}, chi.URLParam(r, "name"))
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			value := chi.URLParam(r, "value")
			require.Equal(t, testValStr, value)
			cFound = true
		}
	}
	jsonHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.Header.Get("Content-Type"), "application/json")
			var (
				buf []byte
				err error
			)
			switch r.Header.Get("Content-Encoding") {
			case "gzip":
				zr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				buf, err = io.ReadAll(zr)
				require.NoError(t, err)
			default:
				buf, err = io.ReadAll(r.Body)
				require.NoError(t, err)
			}
			defer require.NoError(t, r.Body.Close())
			metricList, err := transport.DeserializeMetrics(buf)
			require.NoError(t, err)
			require.NotEmpty(t, metricList)

			for _, m := range metricList {
				switch m.MType {
				case counter:
					require.Contains(t, []string{"counter_var1", "counter_var2"}, m.ID)
					require.NotNil(t, m.Delta)
					require.Equal(t, testValInt64, *m.Delta)
					require.Nil(t, m.Value)
					cJSONFound = true
				case gauge:
					require.Contains(t, []string{"gauge_var1", "gauge_var2"}, m.ID)
					require.NotNil(t, m.Value)
					require.Equal(t, testValFloat, *m.Value)
					require.Nil(t, m.Delta)
					gJSONFound = true
				default:
				}
			}
		}
	}

	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Post("/", jsonHandler())
		r.Post("/gauge/{name}/{value}", gaugeHandler())
		r.Post("/counter/{name}/{value}", counterHandler())
		r.Post("/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
			panic("wrong request")
		})
	})

	// Create test server.
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create a new HTTP client with a default timeout
	timeout := 1000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// Init repositories.
	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	// Prepare data and set to repos.
	gaugeData, err := types.ParseGauge(testValStr)
	require.NoError(t, err)
	counterData, err := types.ParseCounter(testValStr)
	require.NoError(t, err)

	gaugeRepo.Set("gauge_var1", types.GaugeToBytes(gaugeData))
	gaugeRepo.Set("gauge_var2", types.GaugeToBytes(gaugeData))
	counterRepo.Set("counter_var1", types.CounterToBytes(counterData))
	counterRepo.Set("counter_var2", types.CounterToBytes(counterData))

	sendGaugeData(context.Background(), srv.URL, gaugeRepo, client)
	sendCounterData(context.Background(), srv.URL, counterRepo, client)
	require.NoError(t,
		sendJSONData(context.Background(), srv.URL, counter, counterRepo, client))
	require.NoError(t,
		sendJSONData(context.Background(), srv.URL, gauge, gaugeRepo, client))

	require.Eventually(t,
		func() bool {
			return gFound && cFound && gJSONFound && cJSONFound
		},
		200*time.Millisecond, 50*time.Millisecond)
}
