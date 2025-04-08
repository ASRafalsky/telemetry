package reporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/types"
	"github.com/ASRafalsky/telemetry/pkg/services/repository"
)

const testValStr = "1234"

func TestSend(t *testing.T) {
	var (
		gFound, cFound bool
	)

	// Add handlers and router.
	gaugeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "gauge_var", chi.URLParam(r, "name"))
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			value := chi.URLParam(r, "value")
			require.Equal(t, testValStr, value)
			gFound = true
		}
	}
	counterHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "counter_var", chi.URLParam(r, "name"))
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			value := chi.URLParam(r, "value")
			require.Equal(t, testValStr, value)
			cFound = true
		}
	}

	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
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
	repos := repository.NewRepositories()

	// Prepare data and set to repos.
	gaugeData, err := types.ParseGauge(testValStr)
	require.NoError(t, err)
	counterData, err := types.ParseCounter(testValStr)
	require.NoError(t, err)

	repos[repository.Gauge].Set("gauge_var", types.GaugeToBytes(gaugeData))
	repos[repository.Counter].Set("counter_var", types.CounterToBytes(counterData))

	sendGaugeData(context.Background(), srv.URL, repos[repository.Gauge], client)
	sendCounterData(context.Background(), srv.URL, repos[repository.Counter], client)

	require.Eventually(t, func() bool { return gFound && cFound }, 200*time.Millisecond, 50*time.Millisecond)
}
