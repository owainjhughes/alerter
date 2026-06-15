package monitor

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("monitor not found")

type Repository interface {
	Create(ctx context.Context, m *Monitor) error
	Get(ctx context.Context, id uuid.UUID) (*Monitor, error)
	List(ctx context.Context) ([]*Monitor, error)
	Update(ctx context.Context, m *Monitor) error
	Delete(ctx context.Context, id uuid.UUID) error
}
