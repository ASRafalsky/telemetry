package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/pkg/services/handlers"
	"github.com/ASRafalsky/telemetry/pkg/services/templates"
)

var Log = log.NewEmpty()

func main() {
	address, logLevel, path := parseFlags()

	var err error
	if Log, err = log.AddLoggerWith(logLevel, path); err != nil {
		panic(err)
	}
	defer Log.Sync()

	Log.Fatal("Failed to start server:", zap.String("err:", http.ListenAndServe(address, handlers.WithLogging(newRouter(), Log)).Error()))
}

func newRouter() http.Handler {
	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/gauge/{name}/{value}", handlers.GaugePostHandler(gaugeRepo))
			r.Post("/counter/{name}/{value}", handlers.CounterPostHandler(counterRepo))
			r.Post("/{type}/{name}/{value}", handlers.FailurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/gauge/{name}", handlers.GaugeGetHandler(gaugeRepo))
			r.Get("/counter/{name}", handlers.CounterGetHandler(counterRepo))
			r.Get("/{type}/{name}", handlers.FailureGetHandler())
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/", handlers.AllGetHandler(templates.PrepareTemplate(), gaugeRepo, counterRepo))
	})
	return r
}
