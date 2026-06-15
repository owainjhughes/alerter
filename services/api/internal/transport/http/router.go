package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/yourtory/alerter/internal/health"
	spec "github.com/yourtory/alerter/services/api/api"
	"github.com/yourtory/alerter/services/api/internal/monitor"
)

func NewRouter(log *slog.Logger, monitors *monitor.Service, ready health.ReadyFunc) http.Handler {
	mux := http.NewServeMux()

	health.Register(mux, ready)

	mux.HandleFunc("GET /v1/ping", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, log, http.StatusOK, map[string]any{
			"message": "pong",
			"service": "alerter-api",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(spec.OpenAPI)
	})

	(&monitorHandler{svc: monitors, log: log}).register(mux)

	return mux
}

func writeJSON(w http.ResponseWriter, log *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error("encode response", "err", err)
	}
}
