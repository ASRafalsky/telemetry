package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/cmd/agent/poller"
	"github.com/ASRafalsky/telemetry/cmd/agent/reporter"
	"github.com/ASRafalsky/telemetry/cmd/agent/repository"
)

func TestAgent(t *testing.T) {
	var (
		gFound, cFound bool
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

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
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

	t.Log(srv.URL)

	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())

	repos := repository.NewRepositories()

	go poller.Poll(ctx, 20*time.Millisecond, repos)
	go reporter.Send(ctx, srv.URL, 100*time.Millisecond, client, repos)

	require.Eventually(t, func() bool { return gFound && cFound }, 200*time.Millisecond, 50*time.Millisecond)
	cancel()
}
