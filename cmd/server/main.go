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

func main() {
	address, logLevel, path := parseFlags()

	Log, err := log.AddLoggerWith(logLevel, path)
	if err != nil {
		panic(err)
	}
	defer Log.Sync()

	Log.Fatal("Failed to start server:" +
		zap.String("err:", http.ListenAndServe(address, handlers.WithLogging(newRouter(), Log)).Error()).String)
}

func newRouter() http.Handler {
	repos := map[string]handlers.Repository{
		handlers.Gauge:   storage.New[string, []byte](),
		handlers.Counter: storage.New[string, []byte](),
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.JSONPostHandler(repos, handlers.SetDataTo))
			r.Post("/gauge/{name}/{value}", handlers.GaugePostHandler(repos[handlers.Gauge]))
			r.Post("/counter/{name}/{value}", handlers.CounterPostHandler(repos[handlers.Counter]))
			r.Post("/{type}/{name}/{value}", handlers.FailurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.JSONPostHandler(repos, handlers.GetDataFrom))
			r.Get("/gauge/{name}", handlers.GaugeGetHandler(repos[handlers.Gauge]))
			r.Get("/counter/{name}", handlers.CounterGetHandler(repos[handlers.Counter]))
			r.Get("/{type}/{name}", handlers.FailureGetHandler())
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/", handlers.AllGetHandler(templates.PrepareTemplate(), repos))
	})
	return r
}
