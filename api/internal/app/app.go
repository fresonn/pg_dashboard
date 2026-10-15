package app

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/config"
	"dashboard/api/internal/helper"
	infraBadger "dashboard/api/internal/infra/badger"
	"dashboard/api/internal/infra/hub"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/infra/postgres"
	"dashboard/api/internal/service/cluster"
	clusterCache "dashboard/api/internal/service/cluster/repo/cache"
	clusterPostgresRepo "dashboard/api/internal/service/cluster/repo/postgres"
	"dashboard/api/internal/service/database"
	databaseCache "dashboard/api/internal/service/database/repo/cache"
	databasePostgresRepo "dashboard/api/internal/service/database/repo/postgres"
	databaseStorageRepo "dashboard/api/internal/service/database/repo/storage"
	"dashboard/api/internal/service/metrics"
	metricsPostgresRepo "dashboard/api/internal/service/metrics/repo/postgres"
	metricsStorageRepo "dashboard/api/internal/service/metrics/repo/storage"
	"dashboard/api/internal/service/roles"
	rolesPostgresRepo "dashboard/api/internal/service/roles/repo/postgres"
	sharedbuffers "dashboard/api/internal/service/shared-buffers"
	sharedBuffersPostgresRepo "dashboard/api/internal/service/shared-buffers/repo/postgres"
	httpTransport "dashboard/api/internal/transport/http"
	"dashboard/api/internal/usecase"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config  config.AppConfig
	router  http.Handler
	badger  *badger.DB
	log     logger.Logger
	metrics *metrics.Service
}

func New(cfg config.AppConfig) *App {

	slogLogger := logger.New(cfg)

	pgManager := postgres.New(cfg, slogLogger)
	badgerDB := infraBadger.New()

	clusterLogger := logger.WithScopeLogger(slogLogger, "cluster")
	clusterPostgres := clusterPostgresRepo.New(cfg, clusterLogger, pgManager)
	clusterCache := clusterCache.New(cfg, clusterLogger)

	clusterService := cluster.New(cluster.Options{
		Config:          cfg,
		PostgresManager: pgManager,
		Logger:          clusterLogger,
		PostgresRepo:    clusterPostgres,
		Cache:           clusterCache,
	})

	rolesLogger := logger.WithScopeLogger(slogLogger, "role")
	rolesPostgres := rolesPostgresRepo.New(cfg, rolesLogger, pgManager)

	rolesService := roles.New(roles.Options{
		Config:          cfg,
		Logger:          rolesLogger,
		PostgresManager: pgManager,
		PostgresRepo:    rolesPostgres,
	})

	databaseLogger := logger.WithScopeLogger(slogLogger, "database")
	databasePostgres := databasePostgresRepo.New(cfg, databaseLogger, pgManager)
	databaseCache := databaseCache.New(cfg, databaseLogger)
	databaseStorage := databaseStorageRepo.New(cfg, databaseLogger, badgerDB)

	databaseService := database.New(database.Options{
		Config:          cfg,
		Logger:          databaseLogger,
		PostgresManager: pgManager,
		PostgresRepo:    databasePostgres,
		Cache:           databaseCache,
		Storage:         databaseStorage,
	})

	sharedBufferLogger := logger.WithScopeLogger(slogLogger, "shared-buffers")
	sharedBufferPostgres := sharedBuffersPostgresRepo.New(cfg, sharedBufferLogger, pgManager)

	sharedBuffersService := sharedbuffers.New(sharedbuffers.Options{
		Config:          cfg,
		Logger:          sharedBufferLogger,
		PostgresManager: pgManager,
		PostgresRepo:    sharedBufferPostgres,
	})

	metricsLogger := logger.WithScopeLogger(slogLogger, "metrics")
	metricsHub := hub.New(hub.Options{
		Buffer: cfg.Env.Metrics.SubscriberBuffer,
		Logger: metricsLogger,
	})
	metricsPostgres := metricsPostgresRepo.New(cfg, metricsLogger, pgManager)
	metricsStorage := metricsStorageRepo.New(cfg, metricsLogger, badgerDB)

	metricsService := metrics.New(metrics.Options{
		Config:        cfg,
		Logger:        metricsLogger,
		PgManager:     pgManager,
		PostgresRepo:  metricsPostgres,
		SharedBuffers: sharedBuffersService,
		Hub:           metricsHub,
		Storage:       metricsStorage,
	})

	useCases := usecase.New(usecase.Options{
		ClusterService:  clusterService,
		DatabaseService: databaseService,
		PgManager:       pgManager,
	})

	r := chi.NewRouter()
	r.Use(httpTransport.RequestIDMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // todo: make it dynamic
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))

	restHandler := httpTransport.New(useCases, clusterService, rolesService, databaseService, metricsService)

	strictHandler := openapi.NewStrictHandler(restHandler, nil)

	handler := openapi.HandlerFromMuxWithBaseURL(strictHandler, r, "/api")

	// SSE is registered next to the generated router: a strict OpenAPI handler
	// returns a complete body, while the metrics stream is endless.
	metricsStreamHandler := httpTransport.NewMetricsStreamHandler(metricsHub, cfg.Env.Metrics, metricsLogger)
	r.Get(httpTransport.MetricsStreamPath, metricsStreamHandler.ServeHTTP)

	return &App{
		config:  cfg,
		log:     slogLogger,
		badger:  badgerDB,
		router:  handler,
		metrics: metricsService,
	}
}

func (a *App) Run() {

	ctx, cancel := context.WithCancel(context.Background())

	server := &http.Server{
		Addr:    ":" + helper.IntToString(a.config.Env.Port),
		Handler: a.router,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if a.config.Env.IsDev {
		a.log.Info(fmt.Sprintf("http server is running http://localhost%s/api", server.Addr))
	}

	defer cancel()

	go func() {
		<-done
		cancel()
	}()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	})

	// Collectors live as long as the application does, independent of the
	// number of connected clients.
	g.Go(func() error {
		return a.metrics.Run(gCtx)
	})

	if err := g.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) &&
			!errors.Is(err, http.ErrServerClosed) {
			a.log.Error("shutdown with error", "error", err)
		}
	}

	a.log.Info("✅ server shutdown gracefully")

	if err := a.badger.Close(); err != nil {
		a.log.Error("failed to close badger", "error", err)
	} else {
		a.log.Info("✅ badger shutdown gracefully")
	}
}
