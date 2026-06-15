//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourtory/alerter/services/api/internal/adapter/postgres"
	"github.com/yourtory/alerter/services/api/internal/monitor"
)

func TestMonitorRepoPostgres(t *testing.T) {
	dsn := os.Getenv("ALERTER_TEST_DB_DSN")
	if dsn == "" {
		t.Skip("ALERTER_TEST_DB_DSN not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE monitors"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	repo := postgres.NewMonitorRepo(pool)

	m, err := monitor.New(uuid.New(), time.Now().UTC().Truncate(time.Second), monitor.Spec{
		Name:     "pg",
		Type:     monitor.CheckHTTP,
		Target:   "https://example.com",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Enabled:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.Create(ctx, m); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.Get(ctx, m.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "pg" || got.ExpectedStatus != 200 || !got.Enabled {
		t.Errorf("round-trip mismatch: %+v", got)
	}

	m.Name = "pg2"
	m.Enabled = false
	if err := repo.Update(ctx, m); err != nil {
		t.Fatalf("update: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Name != "pg2" || list[0].Enabled {
		t.Errorf("list mismatch: %+v", list)
	}

	if err := repo.Delete(ctx, m.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.Get(ctx, m.ID); !errors.Is(err, monitor.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
