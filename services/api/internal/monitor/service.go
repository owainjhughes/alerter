package monitor

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
	now  func() time.Time
	id   func() uuid.UUID
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
		id:   uuid.New,
	}
}

func (s *Service) Create(ctx context.Context, spec Spec) (*Monitor, error) {
	m, err := New(s.id(), s.now(), spec)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Monitor, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Monitor, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, spec Spec) (*Monitor, error) {
	m, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := m.Update(s.now(), spec); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
