package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ASRafalsky/telemetry/internal/backup"
	"github.com/ASRafalsky/telemetry/internal/handlers"
	"github.com/ASRafalsky/telemetry/internal/middleware"
	"github.com/ASRafalsky/telemetry/internal/repository"
	"github.com/ASRafalsky/telemetry/internal/storage"
	"github.com/ASRafalsky/telemetry/internal/templates"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

func main() {
	cfg, err := updateCfg()
	if err != nil {
		panic(err)
	}

	Log, err := log.AddLoggerWith(cfg.LogLevel, cfg.LogPath)
	if err != nil {
		panic(err)
	}
	defer Log.Sync()

	repo := repository.NewExtendedRepository(storage.New[string, []byte]())

	if cfg.Restore {
		if err = backup.RestoreRepo(cfg.DumpPath, repo); err != nil {
			Log.Error("Failed to restore from the dump file:", cfg.DumpPath, err.Error())
		}
	}

	ctx := context.Background()
	go backup.BackupRepo(ctx, repo, cfg.StorePeriod, cfg.DumpPath, *Log)

	Log.Fatal("Failed to start server:" +
		zap.String("err:",
			http.ListenAndServe(cfg.Addr, middleware.WithLogging(newRouter(repo, Log), Log)).Error()).String)
}

func newRouter(repo dataRepository, logger *log.Logger) http.Handler {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/", middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.SetDataTo), logger))
			r.Post("/gauge/{name}/{value}", handlers.GaugePostHandler(repo))
			r.Post("/counter/{name}/{value}", handlers.CounterPostHandler(repo))
			r.Post("/{type}/{name}/{value}", handlers.FailurePostHandler())
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.GetDataFrom), logger))
			r.Get("/gauge/{name}", handlers.GaugeGetHandler(repo))
			r.Get("/counter/{name}", handlers.CounterGetHandler(repo))
			r.Get("/{type}/{name}", handlers.FailureGetHandler())
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/", middleware.WithCompress(handlers.AllGetHandler(templates.PrepareTemplate(), repo), logger))
	})
	return r
}

type dataRepository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}
