package checker

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/owainjhughes/alerter/internal/events"
)

type Handler struct {
	log    *slog.Logger
	source string
	sink   string
}

func NewHandler(log *slog.Logger, source, sink string) *Handler {
	return &Handler{log: log, source: source, sink: sink}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	evt, err := events.FromRequest(r)
	if err != nil {
		h.log.Warn("not a cloudevent", "err", err)
		http.Error(w, "expected a cloudevent", http.StatusBadRequest)
		return
	}

	var snap events.MonitorSnapshot
	if err := evt.DataAs(&snap); err != nil {
		h.log.Warn("bad event data", "err", err)
		http.Error(w, "invalid event data", http.StatusBadRequest)
		return
	}

	outcome := Probe(r.Context(), snap)
	status := Classify(snap, outcome)

	result := events.CheckResult{
		Monitor:    snap,
		Status:     status,
		LatencyMs:  outcome.Latency.Milliseconds(),
		StatusCode: outcome.StatusCode,
		ObservedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if outcome.Err != nil {
		result.Error = outcome.Err.Error()
	}

	h.log.Info("probed",
		"monitor", snap.Name,
		"target", snap.Target,
		"status", status,
		"latencyMs", result.LatencyMs,
	)

	out, err := events.New(events.TypeCheckCompleted, h.source, result)
	if err != nil {
		h.log.Error("build event", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if h.sink != "" {
		if err := events.Send(r.Context(), h.sink, out); err != nil {
			h.log.Error("emit check.completed", "err", err)
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
