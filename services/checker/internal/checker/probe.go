package checker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/owainjhughes/alerter/internal/events"
)

type Outcome struct {
	StatusCode int
	Latency    time.Duration
	Err        error
}

func Probe(ctx context.Context, snap events.MonitorSnapshot) Outcome {
	timeout := time.Duration(snap.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	switch snap.Type {
	case "http":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, snap.Target, nil)
		if err != nil {
			return Outcome{Latency: time.Since(start), Err: err}
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return Outcome{Latency: time.Since(start), Err: err}
		}
		defer func() { _ = resp.Body.Close() }()
		return Outcome{StatusCode: resp.StatusCode, Latency: time.Since(start)}

	case "tcp":
		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", snap.Target)
		if err != nil {
			return Outcome{Latency: time.Since(start), Err: err}
		}
		_ = conn.Close()
		return Outcome{Latency: time.Since(start)}

	case "dns":
		_, err := net.DefaultResolver.LookupHost(ctx, snap.Target)
		return Outcome{Latency: time.Since(start), Err: err}

	default:
		return Outcome{Err: fmt.Errorf("unknown monitor type %q", snap.Type)}
	}
}
