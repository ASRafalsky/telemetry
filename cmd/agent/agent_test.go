package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/poller"
	"github.com/ASRafalsky/telemetry/internal/reporter"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/cache"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

func TestAgent(t *testing.T) {
	key := "really_secret_key"
	var (
		gFound, cFound, cJSONFound, gJSONFound, psMemFound atomic.Bool
		psCPUCnt, gSendCnt, cSendCnt                       atomic.Int64
	)

	privKey, err := utils.ParsePrivateKey("./testdata/private.key")
	require.NoError(t, err)

	// Add handlers and router.
	gaugeHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			if chi.URLParam(r, "name") == "RandomValue" {
				gFound.Store(true)
			}
		}
	}
	counterHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.Header.Get("Content-Type"), "text/plain")
			if chi.URLParam(r, "name") == "PollCount" {
				cFound.Store(true)
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
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			utils.SignCheck(t, body, []byte(key), r.Header.Get("HashSHA256"))

			data, err := privKey.Decrypt(nil, body, nil)
			require.NoError(t, err)

			r.Body = io.NopCloser(bytes.NewBuffer(data))
			defer require.NoError(t, r.Body.Close())

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
			metricList, err := transport.DeserializeMetrics(buf)
			require.NoError(t, err)
			require.NotEmpty(t, metricList)

			for _, m := range metricList {
				switch m.MType {
				case counter:
					require.NotNil(t, m.Delta)
					require.Nil(t, m.Value)
					cJSONFound.Store(true)
					cSendCnt.Add(1)
				case gauge:
					require.NotNil(t, m.Value)
					require.Nil(t, m.Delta)
					gJSONFound.Store(true)
					if strings.Contains(m.ID, "Memory") {
						psMemFound.Store(true)
					}
					if strings.Contains(m.ID, "CPUutilization") {
						psCPUCnt.Add(1)
					}
					gSendCnt.Add(1)
				default:
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
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			panic("wrong request")
		})
	})

	// Create test server.
	srv := httptest.NewServer(r)
	defer srv.Close()

	client := newClient()
	ctx, cancel := context.WithCancel(context.Background())

	gaugeRepo := cache.New[string, []byte]()
	counterRepo := cache.New[string, []byte]()

	logeer, err := log.AddLoggerWith("info", "")
	require.NoError(t, err)

	pollCfg := poller.Config{
		Interval: time.Duration(10) * time.Millisecond,
	}

	pollerModule := poller.New(pollCfg)
	pollerModule.Run(ctx, poller.GetGaugeMetrics, gaugeRepo, logeer)
	pollerModule.Run(ctx, poller.GetPSMemMetrics, gaugeRepo, logeer)
	pollerModule.Run(ctx, poller.GetPSCPUMetrics, gaugeRepo, logeer)
	pollerModule.Run(ctx, poller.GetCounterMetrics, counterRepo, logeer)

	senderCfg1 := newSenderCfg(config.Agent{
		CommonFields: config.CommonFields{
			Key:    key,
			Crypto: "./testdata/public.key",
		},
		RateLimit: 1,
	})
	senderCfg1.Interval = time.Duration(100) * time.Millisecond
	senderCfg1.Address = srv.URL
	senderCfg2 := newSenderCfg(config.Agent{
		CommonFields: config.CommonFields{
			Addr:   srv.URL,
			Key:    key,
			Crypto: "./testdata/public.key",
		},
		RateLimit: 4,
	})
	senderCfg2.Interval = time.Duration(100) * time.Millisecond
	senderCfg2.Address = srv.URL
	go reporter.Send(ctx, gauge, senderCfg1, client, gaugeRepo, logeer)
	go reporter.Send(ctx, counter, senderCfg2, client, counterRepo, logeer)

	require.Eventually(t,
		func() bool {
			return !gFound.Load() && !cFound.Load() && gJSONFound.Load() && cJSONFound.Load() && psMemFound.Load() &&
				psCPUCnt.Load() == int64(runtime.NumCPU())
		},
		200*time.Millisecond, 50*time.Millisecond)

	time.Sleep(890 * time.Millisecond)

	require.Equal(t, int64(gaugeRepo.Size()), gSendCnt.Load())
	require.Equal(t, int64(counterRepo.Size())*4, cSendCnt.Load())
	cancel()
}
