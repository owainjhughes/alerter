package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validSpec() Spec {
	return Spec{
		Name:     "api uptime",
		Type:     CheckHTTP,
		Target:   "https://example.com/health",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Enabled:  true,
	}
}

func TestNewValid(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	m, err := New(uuid.New(), now, validSpec())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ExpectedStatus != 200 {
		t.Errorf("expected default status 200, got %d", m.ExpectedStatus)
	}
	if !m.CreatedAt.Equal(now) || !m.UpdatedAt.Equal(now) {
		t.Errorf("timestamps not set to now")
	}
}

func TestNewInvalid(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Spec)
		field  string
	}{
		{"empty name", func(s *Spec) { s.Name = "  " }, "name"},
		{"bad type", func(s *Spec) { s.Type = "ping" }, "type"},
		{"empty target", func(s *Spec) { s.Target = "" }, "target"},
		{"non-url http target", func(s *Spec) { s.Target = "example.com" }, "target"},
		{"tcp without port", func(s *Spec) { s.Type = CheckTCP; s.Target = "example.com" }, "target"},
		{"interval too small", func(s *Spec) { s.Interval = time.Second }, "interval"},
		{"timeout exceeds interval", func(s *Spec) { s.Timeout = 60 * time.Second }, "timeout"},
		{"status on non-http", func(s *Spec) { s.Type = CheckTCP; s.Target = "example.com:443"; s.ExpectedStatus = 200 }, "expectedStatus"},
		{"status out of range", func(s *Spec) { s.ExpectedStatus = 99 }, "expectedStatus"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validSpec()
			tc.mutate(&s)
			_, err := New(uuid.New(), time.Now(), s)

			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if ve.Field != tc.field {
				t.Errorf("field = %q, want %q", ve.Field, tc.field)
			}
		})
	}
}

func TestUpdateValidatesAndBumpsTimestamp(t *testing.T) {
	created := time.Unix(1700000000, 0).UTC()
	m, _ := New(uuid.New(), created, validSpec())

	updated := created.Add(time.Hour)
	s := validSpec()
	s.Name = "renamed"
	if err := m.Update(updated, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "renamed" {
		t.Errorf("name not updated")
	}
	if !m.CreatedAt.Equal(created) {
		t.Errorf("createdAt changed on update")
	}
	if !m.UpdatedAt.Equal(updated) {
		t.Errorf("updatedAt = %v, want %v", m.UpdatedAt, updated)
	}

	bad := validSpec()
	bad.Name = ""
	if err := m.Update(updated.Add(time.Hour), bad); err == nil {
		t.Fatalf("expected validation error")
	}
	if m.Name != "renamed" {
		t.Errorf("failed update mutated the aggregate")
	}
}
