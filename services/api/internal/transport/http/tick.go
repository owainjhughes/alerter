package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/owainjhughes/alerter/internal/events"
	"github.com/owainjhughes/alerter/services/api/internal/monitor"
)

// The api doubles as the scheduler; split into its own service only if config and scheduling ever need to scale separately.
type tickHandler struct {
	svc    *monitor.Service
	log    *slog.Logger
	source string
	sink   string
}

func (h *tickHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.sink == "" {
		http.Error(w, "no sink configured", http.StatusServiceUnavailable)
		return
	}

	ms, err := h.svc.List(r.Context())
	if err != nil {
		h.log.Error("tick: list monitors", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	emitted := 0
	for _, m := range ms {
		if !m.Enabled {
			continue
		}
		evt, err := events.New(events.TypeCheckRequested, h.source, snapshot(m))
		if err != nil {
			h.log.Error("tick: build event", "err", err)
			continue
		}
		if err := events.Send(r.Context(), h.sink, evt); err != nil {
			h.log.Error("tick: emit", "monitor", m.Name, "err", err)
			continue
		}
		emitted++
	}

	h.log.Info("tick", "monitors", len(ms), "emitted", emitted)
	w.WriteHeader(http.StatusAccepted)
}

func snapshot(m *monitor.Monitor) events.MonitorSnapshot {
	return events.MonitorSnapshot{
		MonitorID:           m.ID.String(),
		Name:                m.Name,
		Type:                string(m.Type),
		Target:              m.Target,
		TimeoutSeconds:      int(m.Timeout / time.Second),
		ExpectedStatus:      m.ExpectedStatus,
		ConsecutiveFailures: 1, // fixed rule; store per-monitor if rules become configurable
	}
}
