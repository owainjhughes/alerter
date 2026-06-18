package evaluator

import (
	"fmt"
	"sync"

	"github.com/owainjhughes/alerter/internal/events"
)

type state struct {
	consecutiveDown int
	firing          bool
}

type Engine struct {
	mu     sync.Mutex
	states map[string]*state
}

func NewEngine() *Engine {
	return &Engine{states: make(map[string]*state)}
}

type Decision struct {
	Fire    bool
	Resolve bool
	Reason  string
}

func (e *Engine) Apply(res events.CheckResult) Decision {
	e.mu.Lock()
	defer e.mu.Unlock()

	st := e.states[res.Monitor.MonitorID]
	if st == nil {
		st = &state{}
		e.states[res.Monitor.MonitorID] = st
	}

	threshold := max(res.Monitor.ConsecutiveFailures, 1)

	if res.Status == "down" {
		st.consecutiveDown++
		if st.consecutiveDown >= threshold && !st.firing {
			st.firing = true
			return Decision{Fire: true, Reason: fmt.Sprintf("%d consecutive failures", st.consecutiveDown)}
		}
		return Decision{}
	}

	st.consecutiveDown = 0
	if st.firing {
		st.firing = false
		return Decision{Resolve: true, Reason: "recovered"}
	}
	return Decision{}
}
