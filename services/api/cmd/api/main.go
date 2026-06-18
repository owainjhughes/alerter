package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/owainjhughes/alerter/internal/config"
	"github.com/owainjhughes/alerter/internal/health"
	"github.com/owainjhughes/alerter/internal/httpserver"
	applog "github.com/owainjhughes/alerter/internal/log"
	"github.com/owainjhughes/alerter/services/api/internal/adapter/memory"
	"github.com/owainjhughes/alerter/services/api/internal/adapter/postgres"
	"github.com/owainjhughes/alerter/services/api/internal/monitor"
	httpapi "github.com/owainjhughes/alerter/services/api/internal/transport/http"
)

func main() {
	cfg := config.Load("alerter-api")
	logger := applog.New(cfg.LogLevel, cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting service",
		"service", cfg.ServiceName,
		"env", cfg.Env,
		"addr", cfg.HTTPAddr,
	)

	repo, ready, cleanup, err := buildRepo(ctx, cfg, logger)
	if err != nil {
		logger.Error("init repository", "err", err)
		os.Exit(1)
	}
	defer cleanup()

	svc := monitor.NewService(repo)
	router := httpapi.NewRouter(logger, svc, ready)

	if err := httpserver.Run(ctx, cfg.HTTPAddr, router, logger); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
	logger.Info("service stopped cleanly")
}

func buildRepo(ctx context.Context, cfg config.Config, logger *slog.Logger) (monitor.Repository, health.ReadyFunc, func(), error) {
	if cfg.DBDSN == "" {
		logger.Warn("ALERTER_DB_DSN not set, using in-memory repository (data is not persisted)")
		return memory.NewMonitorRepo(), nil, func() {}, nil
	}

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, nil, nil, err
	}
	logger.Info("using postgres repository, migrations applied")

	ready := func() error {
		pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return pool.Ping(pingCtx)
	}
	return postgres.NewMonitorRepo(pool), ready, pool.Close, nil
}
