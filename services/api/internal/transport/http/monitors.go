package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/yourtory/alerter/services/api/internal/monitor"
)

type monitorHandler struct {
	svc *monitor.Service
	log *slog.Logger
}

func (h *monitorHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/monitors", h.create)
	mux.HandleFunc("GET /v1/monitors", h.list)
	mux.HandleFunc("GET /v1/monitors/{id}", h.get)
	mux.HandleFunc("PUT /v1/monitors/{id}", h.update)
	mux.HandleFunc("DELETE /v1/monitors/{id}", h.remove)
}

type monitorRequest struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Target          string `json:"target"`
	IntervalSeconds int    `json:"intervalSeconds"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
	ExpectedStatus  int    `json:"expectedStatus"`
	Enabled         bool   `json:"enabled"`
}

func (r monitorRequest) toSpec() monitor.Spec {
	return monitor.Spec{
		Name:           r.Name,
		Type:           monitor.CheckType(r.Type),
		Target:         r.Target,
		Interval:       time.Duration(r.IntervalSeconds) * time.Second,
		Timeout:        time.Duration(r.TimeoutSeconds) * time.Second,
		ExpectedStatus: r.ExpectedStatus,
		Enabled:        r.Enabled,
	}
}

type monitorResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Target          string `json:"target"`
	IntervalSeconds int    `json:"intervalSeconds"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
	ExpectedStatus  int    `json:"expectedStatus,omitempty"`
	Enabled         bool   `json:"enabled"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

func toResponse(m *monitor.Monitor) monitorResponse {
	return monitorResponse{
		ID:              m.ID.String(),
		Name:            m.Name,
		Type:            string(m.Type),
		Target:          m.Target,
		IntervalSeconds: int(m.Interval / time.Second),
		TimeoutSeconds:  int(m.Timeout / time.Second),
		ExpectedStatus:  m.ExpectedStatus,
		Enabled:         m.Enabled,
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *monitorHandler) create(w http.ResponseWriter, r *http.Request) {
	var req monitorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid json body")
		return
	}
	m, err := h.svc.Create(r.Context(), req.toSpec())
	if err != nil {
		h.fail(w, err)
		return
	}
	w.Header().Set("Location", "/v1/monitors/"+m.ID.String())
	writeJSON(w, h.log, http.StatusCreated, toResponse(m))
}

func (h *monitorHandler) list(w http.ResponseWriter, r *http.Request) {
	ms, err := h.svc.List(r.Context())
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]monitorResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, toResponse(m))
	}
	writeJSON(w, h.log, http.StatusOK, out)
}

func (h *monitorHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, h.log, r)
	if !ok {
		return
	}
	m, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, h.log, http.StatusOK, toResponse(m))
}

func (h *monitorHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, h.log, r)
	if !ok {
		return
	}
	var req monitorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, h.log, http.StatusBadRequest, "invalid json body")
		return
	}
	m, err := h.svc.Update(r.Context(), id, req.toSpec())
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, h.log, http.StatusOK, toResponse(m))
}

func (h *monitorHandler) remove(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, h.log, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *monitorHandler) fail(w http.ResponseWriter, err error) {
	var ve *monitor.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, h.log, http.StatusBadRequest, ve.Error())
	case errors.Is(err, monitor.ErrNotFound):
		writeError(w, h.log, http.StatusNotFound, "monitor not found")
	default:
		h.log.Error("monitor handler", "err", err)
		writeError(w, h.log, http.StatusInternalServerError, "internal error")
	}
}

func parseID(w http.ResponseWriter, log *slog.Logger, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, log, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeError(w http.ResponseWriter, log *slog.Logger, status int, msg string) {
	writeJSON(w, log, status, map[string]string{"error": msg})
}
