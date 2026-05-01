package app

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/config"
	"dashboard/api/internal/postgres"
	"dashboard/api/internal/service/cluster"
	clusterCache "dashboard/api/internal/service/cluster/repo/cache"
	clusterRepo "dashboard/api/internal/service/cluster/repo/storage"
	"dashboard/api/internal/service/database"
	databaseRepo "dashboard/api/internal/service/database/repo/storage"
	"dashboard/api/internal/service/roles"
	rolesRepo "dashboard/api/internal/service/roles/repo/storage"
	httpTransport "dashboard/api/internal/transport/http"
	"dashboard/api/internal/utils"
	"errors"
	"fmt"
	"log/slog"
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
	log    *slog.Logger
}

func New(cfg config.AppConfig, logger *slog.Logger) *App {

	pgManager := postgres.New(cfg, logger)

	clusterStorage := clusterRepo.New(cfg, logger, pgManager)
	clusterCache := clusterCache.New(&cfg, logger)

	clusterService := cluster.New(cluster.Options{
		Config:          cfg,
		Logger:          logger,
		PostgresManager: pgManager,
		Storage:         clusterStorage,
		Cache:           clusterCache,
	})

	rolesStorage := rolesRepo.New(cfg, logger, pgManager)

	rolesService := roles.New(roles.Options{
		Config:          cfg,
		Logger:          logger,
		PostgresManager: pgManager,
		Storage:         rolesStorage,
	})

	databaseStorage := databaseRepo.New(cfg, logger, pgManager)

	databaseService := database.New(database.Options{
		Config:          cfg,
		Logger:          logger,
		PostgresManager: pgManager,
		Storage:         databaseStorage,
	})

	r := chi.NewRouter()
	r.Use(requestIDMiddleware)
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
		log:    logger,
		router: handler,
	}
}

func (a *App) Run() {

	ctx, cancel := context.WithCancel(context.Background())

	server := &http.Server{
		Addr:    ":" + utils.IntToString(a.config.Env.Port),
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
