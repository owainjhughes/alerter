package checker

import (
	"time"

	"github.com/owainjhughes/alerter/internal/events"
)

const (
	StatusUp       = "up"
	StatusDown     = "down"
	StatusDegraded = "degraded"
)

// Classify turns a raw probe Outcome into a CheckResult status, judged against
// the monitor's success criteria carried in the snapshot.
func Classify(snap events.MonitorSnapshot, o Outcome) string {
	if o.Err != nil {
		return StatusDown
	}

	if snap.Type == "http" {
		expected := snap.ExpectedStatus
		if expected == 0 {
			expected = 200
		}
		if o.StatusCode != expected {
			return StatusDown
		}
	}

	budget := time.Duration(snap.TimeoutSeconds) * time.Second
	if budget > 0 && o.Latency > budget/2 {
		return StatusDegraded
	}
	return StatusUp
}
