package events

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
)

const (
	TypeCheckRequested = "io.alerter.check.requested"
	TypeCheckCompleted = "io.alerter.check.completed"
	TypeAlertRaised    = "io.alerter.alert.raised"
	TypeAlertResolved  = "io.alerter.alert.resolved"
)

const contentTypeCloudEvent = "application/cloudevents+json"

type MonitorSnapshot struct {
	MonitorID           string `json:"monitorId"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Target              string `json:"target"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	ExpectedStatus      int    `json:"expectedStatus"`
	ConsecutiveFailures int    `json:"consecutiveFailures"`
}

type CheckResult struct {
	Monitor    MonitorSnapshot `json:"monitor"`
	Status     string          `json:"status"`
	LatencyMs  int64           `json:"latencyMs"`
	StatusCode int             `json:"statusCode,omitempty"`
	Error      string          `json:"error,omitempty"`
	ObservedAt string          `json:"observedAt"`
}

type AlertRaised struct {
	MonitorID string `json:"monitorId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	At        string `json:"at"`
}

func New(eventType, source string, data any) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.NewString())
	e.SetType(eventType)
	e.SetSource(source)
	e.SetTime(time.Now().UTC())
	if err := e.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return e, err
	}
	return e, nil
}

func FromRequest(r *http.Request) (*cloudevents.Event, error) {
	return cehttp.NewEventFromHTTPRequest(r)
}

func Send(ctx context.Context, sinkURL string, e cloudevents.Event) error {
	body, err := e.MarshalJSON()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sinkURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeCloudEvent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sink %s returned status %d", sinkURL, resp.StatusCode)
	}
	return nil
}
