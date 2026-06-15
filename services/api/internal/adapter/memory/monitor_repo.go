package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/yourtory/alerter/services/api/internal/monitor"
)

type MonitorRepo struct {
	mu    sync.RWMutex
	items map[uuid.UUID]monitor.Monitor
}

func NewMonitorRepo() *MonitorRepo {
	return &MonitorRepo{items: make(map[uuid.UUID]monitor.Monitor)}
}

func (r *MonitorRepo) Create(_ context.Context, m *monitor.Monitor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[m.ID] = *m
	return nil
}

func (r *MonitorRepo) Get(_ context.Context, id uuid.UUID) (*monitor.Monitor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.items[id]
	if !ok {
		return nil, monitor.ErrNotFound
	}
	return &m, nil
}

func (r *MonitorRepo) List(_ context.Context) ([]*monitor.Monitor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*monitor.Monitor, 0, len(r.items))
	for _, m := range r.items {
		out = append(out, &m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *MonitorRepo) Update(_ context.Context, m *monitor.Monitor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[m.ID]; !ok {
		return monitor.ErrNotFound
	}
	r.items[m.ID] = *m
	return nil
}

func (r *MonitorRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return monitor.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
