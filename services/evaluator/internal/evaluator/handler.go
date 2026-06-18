package evaluator

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
	engine *Engine
}

func NewHandler(log *slog.Logger, source, sink string) *Handler {
	return &Handler{log: log, source: source, sink: sink, engine: NewEngine()}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	evt, err := events.FromRequest(r)
	if err != nil {
		h.log.Warn("not a cloudevent", "err", err)
		http.Error(w, "expected a cloudevent", http.StatusBadRequest)
		return
	}

	var res events.CheckResult
	if err := evt.DataAs(&res); err != nil {
		h.log.Warn("bad event data", "err", err)
		http.Error(w, "invalid event data", http.StatusBadRequest)
		return
	}

	switch dec := h.engine.Apply(res); {
	case dec.Fire:
		h.notify(r, events.TypeAlertRaised, res, dec.Reason, "ALERT RAISED")
	case dec.Resolve:
		h.notify(r, events.TypeAlertResolved, res, dec.Reason, "ALERT RESOLVED")
	default:
		h.log.Info("evaluated", "monitor", res.Monitor.Name, "status", res.Status)
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) notify(r *http.Request, eventType string, res events.CheckResult, reason, banner string) {
	h.log.Warn(banner,
		"monitor", res.Monitor.Name,
		"target", res.Monitor.Target,
		"reason", reason,
	)

	if h.sink == "" {
		return
	}
	alert := events.AlertRaised{
		MonitorID: res.Monitor.MonitorID,
		Name:      res.Monitor.Name,
		Status:    res.Status,
		Reason:    reason,
		At:        time.Now().UTC().Format(time.RFC3339),
	}
	out, err := events.New(eventType, h.source, alert)
	if err != nil {
		h.log.Error("build alert event", "err", err)
		return
	}
	if err := events.Send(r.Context(), h.sink, out); err != nil {
		h.log.Error("emit alert", "err", err)
	}
}
