package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/owainjhughes/alerter/services/api/internal/adapter/memory"
	"github.com/owainjhughes/alerter/services/api/internal/monitor"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	svc := monitor.NewService(memory.NewMonitorRepo())
	srv := httptest.NewServer(NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), svc, nil, ""))
	t.Cleanup(srv.Close)
	return srv
}

func TestRouterRoutes(t *testing.T) {
	srv := newTestServer(t)

	cases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"liveness", "/healthz", http.StatusOK},
		{"readiness", "/readyz", http.StatusOK},
		{"ping", "/v1/ping", http.StatusOK},
		{"openapi", "/openapi.yaml", http.StatusOK},
		{"unknown route", "/does-not-exist", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + tc.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tc.path, err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("GET %s: status = %d, want %d", tc.path, resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

func TestPingReturnsJSON(t *testing.T) {
	srv := newTestServer(t)

	resp, err := http.Get(srv.URL + "/v1/ping")
	if err != nil {
		t.Fatalf("GET /v1/ping: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
}
