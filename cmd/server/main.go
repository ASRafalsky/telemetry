package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ASRafalsky/telemetry/internal/log"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/pkg/services/handlers"
	"github.com/ASRafalsky/telemetry/pkg/templates"
)

func main() {
	address, logLevel, path, dump, storePeriod, restore := parseFlags()

	Log, err := log.AddLoggerWith(logLevel, path)
	if err != nil {
		panic(err)
	}
	defer Log.Sync()

	repo := storage.New[string, []byte]()

	if restore {
		if err = restoreRepo(dump, repo); err != nil {
			Log.Error("Failed to restore from the dump file:", dump, err.Error())
		}
	}

	ctx := context.Background()

	go backupRepo(ctx, repo, storePeriod, dump, *Log)

	Log.Fatal("Failed to start server:" +
		zap.String("err:", http.ListenAndServe(address, handlers.WithLogging(newRouter(repo), Log)).Error()).String)
}

func newRouter(repo repository) http.Handler {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.WithCompress(handlers.JSONPostHandler(repo, handlers.SetDataTo)))
			r.Post("/gauge/{name}/{value}", handlers.GaugePostHandler(repo))
			r.Post("/counter/{name}/{value}", handlers.CounterPostHandler(repo))
			r.Post("/{type}/{name}/{value}", handlers.FailurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.WithCompress(handlers.JSONPostHandler(repo, handlers.GetDataFrom)))
			r.Get("/gauge/{name}", handlers.GaugeGetHandler(repo))
			r.Get("/counter/{name}", handlers.CounterGetHandler(repo))
			r.Get("/{type}/{name}", handlers.FailureGetHandler())
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/", handlers.WithCompress(handlers.AllGetHandler(templates.PrepareTemplate(), repo)))
	})
	return r
}

type repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}
