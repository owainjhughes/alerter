package monitor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yourtory/alerter/services/api/internal/adapter/memory"
	"github.com/yourtory/alerter/services/api/internal/monitor"
)

func spec() monitor.Spec {
	return monitor.Spec{
		Name:     "api uptime",
		Type:     monitor.CheckHTTP,
		Target:   "https://example.com",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Enabled:  true,
	}
}

func TestServiceCreateAndGet(t *testing.T) {
	svc := monitor.NewService(memory.NewMonitorRepo())
	ctx := context.Background()

	created, err := svc.Create(ctx, spec())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("expected generated id")
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "api uptime" {
		t.Errorf("name = %q", got.Name)
	}
}

func TestServiceCreateRejectsInvalid(t *testing.T) {
	svc := monitor.NewService(memory.NewMonitorRepo())
	bad := spec()
	bad.Name = ""
	if _, err := svc.Create(context.Background(), bad); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestServiceUpdateAndDelete(t *testing.T) {
	svc := monitor.NewService(memory.NewMonitorRepo())
	ctx := context.Background()

	m, _ := svc.Create(ctx, spec())

	upd := spec()
	upd.Enabled = false
	upd.Name = "renamed"
	got, err := svc.Update(ctx, m.ID, upd)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Enabled || got.Name != "renamed" {
		t.Errorf("update not applied: %+v", got)
	}

	if err := svc.Delete(ctx, m.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.Get(ctx, m.ID); !errors.Is(err, monitor.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestServiceUpdateMissing(t *testing.T) {
	svc := monitor.NewService(memory.NewMonitorRepo())
	if _, err := svc.Update(context.Background(), uuid.New(), spec()); !errors.Is(err, monitor.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
