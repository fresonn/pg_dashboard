package app

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/config"
	"dashboard/api/internal/helper"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/infra/postgres"
	"dashboard/api/internal/service/cluster"
	clusterCache "dashboard/api/internal/service/cluster/repo/cache"
	clusterPostgresRepo "dashboard/api/internal/service/cluster/repo/postgres"
	"dashboard/api/internal/service/database"
	databasePostgresRepo "dashboard/api/internal/service/database/repo/postgres"
	"dashboard/api/internal/service/roles"
	rolesPostgresRepo "dashboard/api/internal/service/roles/repo/postgres"
	httpTransport "dashboard/api/internal/transport/http"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config config.AppConfig
	router http.Handler
	log    logger.Logger
}

func New(cfg config.AppConfig) *App {

	slogLogger := logger.New(cfg)

	pgManager := postgres.New(cfg, slogLogger)

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

	databaseService := database.New(database.Options{
		Config:          cfg,
		Logger:          databaseLogger,
		PostgresManager: pgManager,
		PostgresRepo:    databasePostgres,
	})

	r := chi.NewRouter()
	r.Use(httpTransport.RequestIDMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // todo: make it dynamic
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))

	restHandler := httpTransport.New(clusterService, rolesService, databaseService)

	strictHandler := openapi.NewStrictHandler(restHandler, nil)

	handler := openapi.HandlerFromMuxWithBaseURL(strictHandler, r, "/api")

	return &App{
		config: cfg,
		log:    slogLogger,
		router: handler,
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

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("shutdown with error: %v", err)
	} else {
		fmt.Println("✅ server shutdown gracefully")
	}
}
