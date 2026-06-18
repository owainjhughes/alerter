package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/owainjhughes/alerter/internal/config"
	"github.com/owainjhughes/alerter/internal/health"
	"github.com/owainjhughes/alerter/internal/httpserver"
	applog "github.com/owainjhughes/alerter/internal/log"
	"github.com/owainjhughes/alerter/services/evaluator/internal/evaluator"
)

func main() {
	cfg := config.Load("alerter-evaluator")
	logger := applog.New(cfg.LogLevel, cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	health.Register(mux, nil)
	mux.Handle("POST /", evaluator.NewHandler(logger, cfg.ServiceName, cfg.SinkURL))

	logger.Info("starting service",
		"service", cfg.ServiceName,
		"addr", cfg.HTTPAddr,
		"sink", cfg.SinkURL,
	)
	if err := httpserver.Run(ctx, cfg.HTTPAddr, mux, logger); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}
