package reporter

import (
	"bytes"
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

	"github.com/ASRafalsky/telemetry/internal/cache"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
	"github.com/ASRafalsky/telemetry/internal/utils"
)

const (
	testValStr   = "1234"
	testValInt64 = int64(1234)
	testValFloat = float64(1234)
	secretKey    = "secret"
)

func TestSend(t *testing.T) {
	var (
		gFound, cFound, gJSONFound, cJSONFound bool
	)

	// Add handlers and router.
	gaugeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Contains(t, []string{"_var1", "_var2"}, chi.URLParam(r, "name"))
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			value := chi.URLParam(r, "value")
			require.Equal(t, testValStr, value)
			gFound = true
		}
	}
	counterHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Contains(t, []string{"_var3", "_var4"}, chi.URLParam(r, "name"))
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
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			utils.SignCheck(t, body, []byte(secretKey), r.Header.Get("HashSHA256"))
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
					require.Contains(t, []string{"_var3", "_var4"}, m.ID)
					require.NotNil(t, m.Delta)
					require.Equal(t, testValInt64, *m.Delta)
					require.Nil(t, m.Value)
					cJSONFound = true
				case gauge:
					require.Contains(t, []string{"_var1", "_var2"}, m.ID)
					require.NotNil(t, m.Value)
					require.Equal(t, testValFloat, *m.Value)
					require.Nil(t, m.Delta)
					gJSONFound = true
				default:
					require.Fail(t, "unknown metric type")
				}
			}
		}
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/", jsonHandler())
			r.Post("/gauge/{name}/{value}", gaugeHandler())
			r.Post("/counter/{name}/{value}", counterHandler())
			r.Post("/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
				panic("wrong request")
			})
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", jsonHandler())
		})
	})

	// Create test server.
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create a new HTTP client with a default timeout
	timeout := 1000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// Init repository.
	repo := cache.New[string, []byte]()

	// Prepare data and set to repos.
	gaugeData, err := types.ParseGauge(testValStr)
	require.NoError(t, err)
	counterData, err := types.ParseCounter(testValStr)
	require.NoError(t, err)

	repo.Set(gauge+"_var1", types.GaugeToBytes(gaugeData))
	repo.Set(gauge+"_var2", types.GaugeToBytes(gaugeData))
	repo.Set(counter+"_var3", types.CounterToBytes(counterData))
	repo.Set(counter+"_var4", types.CounterToBytes(counterData))

	require.NoError(t, sendGaugeData(context.Background(), srv.URL, repo, client))
	require.Eventually(t,
		func() bool {
			return gFound
		},
		200*time.Millisecond, 50*time.Millisecond)

	require.NoError(t, sendCounterData(context.Background(), srv.URL, repo, client))
	require.Eventually(t,
		func() bool {
			return cFound
		},
		200*time.Millisecond, 50*time.Millisecond)

	cfg := Config{
		Address: srv.URL,
		Key:     secretKey,
	}

	header, data := prepareData(t, repo, counter, secretKey)
	require.NoError(t, sendDataTo("/updates/", cfg, header, data, client))
	require.Eventually(t,
		func() bool {
			return cJSONFound
		},
		200*time.Millisecond, 50*time.Millisecond)

	repo.Set(gauge+"_var1", types.GaugeToBytes(gaugeData))
	repo.Set(gauge+"_var2", types.GaugeToBytes(gaugeData))

	header, data = prepareData(t, repo, gauge, secretKey)
	require.NoError(t, sendDataTo("/updates/", cfg, header, data, client))
	require.Eventually(t,
		func() bool {
			return gJSONFound
		},
		200*time.Millisecond, 50*time.Millisecond)

	cJSONFound, gJSONFound = false, false

	// After sending gauge entries have been dropped.
	header, data = prepareData(t, repo, "", secretKey)
	require.NoError(t, sendDataTo("/updates/", cfg, header, data, client))

	require.Eventually(t,
		func() bool {
			return !gJSONFound && cJSONFound
		},
		200*time.Millisecond, 50*time.Millisecond)
}

func prepareData(t *testing.T, repo repository, mType, key string) (http.Header, io.Reader) {
	t.Helper()
	var bufToSend = bytes.NewBuffer(nil)

	zw := gzip.NewWriter(bufToSend)
	require.NoError(t, serializeMetrics(context.Background(), mType, repo, zw))
	require.NoError(t, zw.Close())
	header := http.Header{
		"Content-Type":     []string{"application/json"},
		"Content-Encoding": []string{"gzip"},
	}

	signature, err := sign(bufToSend.Bytes(), []byte(key))
	require.NoError(t, err)
	header.Set("HashSHA256", signature)
	return header, bufToSend
}
