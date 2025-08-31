package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	_ "google.golang.org/grpc/encoding/gzip"

	"github.com/ASRafalsky/telemetry/internal/interceptor"
	pb "github.com/ASRafalsky/telemetry/proto"

	"github.com/ASRafalsky/telemetry/internal/config"
	"github.com/ASRafalsky/telemetry/internal/diagnostic"
	"github.com/ASRafalsky/telemetry/internal/handlers"
	"github.com/ASRafalsky/telemetry/internal/middleware"
	"github.com/ASRafalsky/telemetry/internal/repository"
	"github.com/ASRafalsky/telemetry/internal/templates"
	"github.com/ASRafalsky/telemetry/internal/utils"
	"github.com/ASRafalsky/telemetry/pkg/cache"
	"github.com/ASRafalsky/telemetry/pkg/log"
)

var (
	buildVersion, buildDate, buildCommit string
)

func main() {
	utils.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	cfg, err := updateCfg()
	if err != nil {
		panic(fmt.Errorf("failed to update config: %w", err))
	}

	Log, err := log.AddLoggerWith(cfg.LogLevel, cfg.LogPath)
	if err != nil {
		panic(fmt.Errorf("failed to add logger: %w", err))
	}
	defer Log.Sync()

	Log.Info("Starting telemetry server", cfg.Addr, cfg.LogLevel, cfg.DumpPath)
	repo := repository.NewExtendedRepository(cache.New[string, []byte]())

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
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

	go func() {
		diagCfg := newDiagnosticCfg(cfg)
		runServer(ctx, cancel, diagnosticRouter(ctx, diagCfg.Diag), diagCfg, Log, repo)
	}()
	const srcIPInHeader = true
	runServer(ctx, cancel, middleware.WithLogging(middleware.CheckCIDR(middleware.WithSign(
		middleware.Decrypt(
			newRouter(repo, Log), cfg.PrivateKey), []byte(cfg.Key), Log), srcIPInHeader, cfg.CIDR), Log), cfg, Log, repo)

	Log.Info("Telemetry Server stopped.")
}

func runServer(
	ctx context.Context,
	cancel context.CancelFunc,
	handler http.Handler,
	cfg config.Server,
	logger *log.Logger,
	repo dataRepository,
) {
	logger.Info("Starting server", cfg.Addr)

	var srv interface{}
	if cfg.GRPC {
		srv = grpc.NewServer(
			grpc.MaxConcurrentStreams(100500),
			grpc.UnaryInterceptor(interceptor.NewInterceptorsChain(
				true, cfg.CIDR,
				[]byte(cfg.Key),
				cfg.PrivateKey,
				logger,
			)))
	} else {
		srv = &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		}
	}

	go func() {
		if cfg.GRPC {
			listen, err := net.Listen("tcp", cfg.Addr)
			if err != nil {
				logger.Fatal(err.Error())
			}
			s := srv.(*grpc.Server)
			pb.RegisterBDServer(s, handlers.DBServer{Repo: repo})
			pb.RegisterMetricsServer(s, handlers.MetricsServer{Repo: repo})
			pb.RegisterMetricsMultiServer(s, handlers.MetricsMultiServer{Repo: repo})
			if err := s.Serve(listen); err != nil {
				logger.Fatal(err.Error())
			}
		} else {
			if err := srv.(*http.Server).ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("Failed to start server:", err.Error())
			}
		}
		cancel()
	}()

	<-ctx.Done()
	if ctxErr := ctx.Err(); ctxErr != nil {
		logger.Info("Stopping server by cause:", ctxErr.Error())
	}

	srvCtx, srvCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer srvCancel()

	if cfg.GRPC {
		srv.(*grpc.Server).GracefulStop()
	} else {
		if err := srv.(*http.Server).Shutdown(srvCtx); err != nil {
			logger.Fatal("Failed to shutdown server:", err.Error())
		}
	}

	logger.Info("Server shutdown completed")
}

func newRouter(repo dataRepository, logger *log.Logger) http.Handler {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/",
				middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.SetDataTo), logger))
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
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.DBPingHandler(repo))
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", middleware.WithCompress(handlers.JSONPostHandler(repo, handlers.SetDataTo), logger))
		})
		r.Post("/", handlers.FailurePostHandler())
		r.Get("/",
			middleware.WithCompress(handlers.AllGetHandler(templates.PrepareTemplate(), repo), logger))
	})
	return r
}

func diagnosticRouter(ctx context.Context, cfg config.Diagnostics) http.Handler {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/profile", func(r chi.Router) {
			r.Post("/", diagnostic.ProfilePostHandler(ctx, cfg.Path))
		})
	})
	return r
}

type dataRepository interface {
	Set(k string, v []byte)
	Merge(src map[string][]byte)
	Get(ctx context.Context, k string) ([]byte, error)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Delete(ctx context.Context, k string) error
	Ping(ctx context.Context) error
	Size() (int, error)
}
