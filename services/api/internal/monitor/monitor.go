package monitor

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CheckType string

const (
	CheckHTTP CheckType = "http"
	CheckTCP  CheckType = "tcp"
	CheckDNS  CheckType = "dns"
)

func (t CheckType) valid() bool {
	switch t {
	case CheckHTTP, CheckTCP, CheckDNS:
		return true
	default:
		return false
	}
}

type Monitor struct {
	ID             uuid.UUID
	Name           string
	Type           CheckType
	Target         string
	Interval       time.Duration
	Timeout        time.Duration
	ExpectedStatus int
	Enabled        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Spec struct {
	Name           string
	Type           CheckType
	Target         string
	Interval       time.Duration
	Timeout        time.Duration
	ExpectedStatus int
	Enabled        bool
}

const (
	minInterval = 5 * time.Second
	minTimeout  = 1 * time.Second
	maxNameLen  = 200
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func New(id uuid.UUID, now time.Time, s Spec) (*Monitor, error) {
	m := &Monitor{ID: id, CreatedAt: now, UpdatedAt: now}
	if err := m.apply(s); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Monitor) Update(now time.Time, s Spec) error {
	if err := m.apply(s); err != nil {
		return err
	}
	m.UpdatedAt = now
	return nil
}

func (m *Monitor) apply(s Spec) error {
	name := strings.TrimSpace(s.Name)
	switch {
	case name == "":
		return &ValidationError{"name", "must not be empty"}
	case len(name) > maxNameLen:
		return &ValidationError{"name", fmt.Sprintf("must be at most %d characters", maxNameLen)}
	}

	if !s.Type.valid() {
		return &ValidationError{"type", "must be one of http, tcp, dns"}
	}
	if err := validateTarget(s.Type, s.Target); err != nil {
		return err
	}

	switch {
	case s.Interval < minInterval:
		return &ValidationError{"interval", "must be at least 5s"}
	case s.Timeout < minTimeout:
		return &ValidationError{"timeout", "must be at least 1s"}
	case s.Timeout > s.Interval:
		return &ValidationError{"timeout", "must not exceed interval"}
	}

	status := s.ExpectedStatus
	if s.Type == CheckHTTP {
		if status == 0 {
			status = 200
		}
		if status < 100 || status > 599 {
			return &ValidationError{"expectedStatus", "must be between 100 and 599"}
		}
	} else if status != 0 {
		return &ValidationError{"expectedStatus", "only valid for http monitors"}
	}

	m.Name = name
	m.Type = s.Type
	m.Target = strings.TrimSpace(s.Target)
	m.Interval = s.Interval
	m.Timeout = s.Timeout
	m.ExpectedStatus = status
	m.Enabled = s.Enabled
	return nil
}

func validateTarget(t CheckType, target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return &ValidationError{"target", "must not be empty"}
	}
	switch t {
	case CheckHTTP:
		u, err := url.Parse(target)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return &ValidationError{"target", "must be a valid http(s) URL"}
		}
	case CheckTCP:
		host, port, err := net.SplitHostPort(target)
		if err != nil || host == "" || port == "" {
			return &ValidationError{"target", "must be host:port"}
		}
	case CheckDNS:
		if strings.ContainsAny(target, "/: ") {
			return &ValidationError{"target", "must be a hostname"}
		}
	}
	return nil
}
