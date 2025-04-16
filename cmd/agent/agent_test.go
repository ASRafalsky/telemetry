package main

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/pkg/services/poller"
	"github.com/ASRafalsky/telemetry/pkg/services/reporter"
)

func TestAgent(t *testing.T) {
	var (
		gFound, cFound, cJSONFound, gJSONFound bool
	)

	// Add handlers and router.
	gaugeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			if chi.URLParam(r, "name") == "RandomValue" {
				gFound = true
			}
		}
	}
	counterHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			if chi.URLParam(r, "name") == "PollCount" {
				cFound = true
			}
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
			m := transport.Metrics{}
			require.NoError(t, easyjson.Unmarshal(buf, &m))
			switch m.MType {
			case counter:
				require.NotNil(t, m.Delta)
				require.Nil(t, m.Value)
				cJSONFound = true
			case gauge:
				require.NotNil(t, m.Value)
				require.Nil(t, m.Delta)
				gJSONFound = true
			default:
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
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			panic("wrong request")
		})
	})

	// Create test server.
	srv := httptest.NewServer(r)
	defer srv.Close()

	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())

	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	log, err := log.AddLoggerWith("info", "")
	require.NoError(t, err)

	go poller.Poll(ctx, 20*time.Millisecond, map[string]poller.Repository{
		gauge:   gaugeRepo,
		counter: counterRepo,
	}, log)
	go reporter.Send(ctx, srv.URL, 100*time.Millisecond, client, map[string]reporter.Repository{
		gauge:   gaugeRepo,
		counter: counterRepo,
	}, log)

	require.Eventually(t,
		func() bool {
			return !gFound && !cFound && gJSONFound && cJSONFound
		},
		200*time.Millisecond, 50*time.Millisecond)
	cancel()
}
