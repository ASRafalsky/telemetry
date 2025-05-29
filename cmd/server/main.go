package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ASRafalsky/telemetry/internal/cache"
	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/handlers"
	"github.com/ASRafalsky/telemetry/internal/middleware"
	"github.com/ASRafalsky/telemetry/internal/repository"
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

	Log.Info("Starting telemetry server", cfg.Addr, cfg.LogLevel, cfg.DumpPath)
	repo := repository.NewExtendedRepository(cache.New[string, []byte]())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if db, err := initDB(ctx, cfg.DB, *Log); err == nil {
		repo.UseDB(db)
		defer func() {
			if err = db.Close(); err != nil {
				Log.Error("Failed to close db:", err.Error())
			}
		}()
	}

	repo.Maintain(ctx, cfg, *Log)

	runServer(ctx, cancel, repo, cfg, Log)
}

func runServer(ctx context.Context, cancel context.CancelFunc, repo dataRepository, cfg config.Server, logger *log.Logger) {
	logger.Info("Starting server", cfg.Addr)

	srv := http.Server{
		Addr:    cfg.Addr,
		Handler: middleware.WithLogging(newRouter(repo, cfg, logger), logger),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal("Failed to start server:", err.Error())
			cancel()
		}
	}()

	<-ctx.Done()
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer srvCancel()
	if err := srv.Shutdown(srvCtx); err != nil {
		logger.Fatal("Failed to shutdown server:", err.Error())
	}
	logger.Info("Server shutdown completed")
}

func newRouter(repo dataRepository, cfg config.Server, logger *log.Logger) http.Handler {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/",
				middleware.WithSign(middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.SetDataTo), logger),
					[]byte(cfg.Key), logger))
			r.Post("/gauge/{name}/{value}",
				middleware.WithSign(handlers.GaugePostHandler(repo), []byte(cfg.Key), logger))
			r.Post("/counter/{name}/{value}",
				middleware.WithSign(handlers.CounterPostHandler(repo), []byte(cfg.Key), logger))
			r.Post("/{type}/{name}/{value}",
				middleware.WithSign(handlers.FailurePostHandler(), []byte(cfg.Key), logger))
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/",
				middleware.WithSign(middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.GetDataFrom), logger),
					[]byte(cfg.Key), logger))
			r.Get("/gauge/{name}", middleware.WithSign(handlers.GaugeGetHandler(repo), []byte(cfg.Key), logger))
			r.Get("/counter/{name}", middleware.WithSign(handlers.CounterGetHandler(repo), []byte(cfg.Key), logger))
			r.Get("/{type}/{name}", middleware.WithSign(handlers.FailureGetHandler(), []byte(cfg.Key), logger))
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", middleware.WithSign(handlers.DBPingHandler(repo), []byte(cfg.Key), logger))
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/",
				middleware.WithSign(middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.SetDataTo), logger),
					[]byte(cfg.Key), logger))
		})
		r.Post("/", middleware.WithSign(handlers.FailurePostHandler(), []byte(cfg.Key), logger))
		r.Get("/",
			middleware.WithSign(middleware.WithCompress(handlers.AllGetHandler(templates.PrepareTemplate(), repo), logger),
				[]byte(cfg.Key), logger))
	})
	return r
}

type dataRepository interface {
	Set(k string, v []byte)
	Get(ctx context.Context, k string) ([]byte, error)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Delete(ctx context.Context, k string) error
	Ping(ctx context.Context) error
	Size() (int, error)
}
