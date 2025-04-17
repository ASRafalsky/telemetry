package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/pkg/services/backup"
	"github.com/ASRafalsky/telemetry/pkg/services/handlers"
	"github.com/ASRafalsky/telemetry/pkg/services/templates"
)

func main() {
	address, logLevel, path, dump, storePeriod, restore := parseFlags()

	Log, err := log.AddLoggerWith(logLevel, path)
	if err != nil {
		panic(err)
	}
	defer Log.Sync()

	gaugeRepo := storage.New[string, []byte]()
	counterRepo := storage.New[string, []byte]()

	if restore {
		if err = restoreRepo(dump, map[string]backup.Repository{
			handlers.Gauge:   gaugeRepo,
			handlers.Counter: counterRepo,
		}); err != nil {
			Log.Error("Failed to restore from the dump file:", dump, err.Error())
		}
	}

	ctx := context.Background()

	go backupRepo(ctx, map[string]backup.Repository{
		handlers.Gauge:   gaugeRepo,
		handlers.Counter: counterRepo,
	}, time.Duration(storePeriod)*time.Second, dump, *Log)

	Log.Fatal("Failed to start server:" +
		zap.String("err:", http.ListenAndServe(address, handlers.WithLogging(newRouter(map[string]handlers.Repository{
			handlers.Gauge:   gaugeRepo,
			handlers.Counter: counterRepo,
		}), Log)).Error()).String)
}

func newRouter(repos map[string]handlers.Repository) http.Handler {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.WithCompress(handlers.JSONPostHandler(repos, handlers.SetDataTo)))
			r.Post("/gauge/{name}/{value}", handlers.GaugePostHandler(repos[handlers.Gauge]))
			r.Post("/counter/{name}/{value}", handlers.CounterPostHandler(repos[handlers.Counter]))
			r.Post("/{type}/{name}/{value}", handlers.FailurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.WithCompress(handlers.JSONPostHandler(repos, handlers.GetDataFrom)))
			r.Get("/gauge/{name}", handlers.GaugeGetHandler(repos[handlers.Gauge]))
			r.Get("/counter/{name}", handlers.CounterGetHandler(repos[handlers.Counter]))
			r.Get("/{type}/{name}", handlers.FailureGetHandler())
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/", handlers.WithCompress(handlers.AllGetHandler(templates.PrepareTemplate(), repos)))
	})
	return r
}
