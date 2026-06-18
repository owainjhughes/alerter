package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/owainjhughes/alerter/services/api/internal/monitor"
)

type MonitorRepo struct {
	pool *pgxpool.Pool
}

func NewMonitorRepo(pool *pgxpool.Pool) *MonitorRepo {
	return &MonitorRepo{pool: pool}
}

const monitorColumns = "id, name, type, target, interval_seconds, timeout_seconds, expected_status, enabled, created_at, updated_at"

func (r *MonitorRepo) Create(ctx context.Context, m *monitor.Monitor) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO monitors (`+monitorColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		m.ID, m.Name, string(m.Type), m.Target,
		seconds(m.Interval), seconds(m.Timeout), m.ExpectedStatus,
		m.Enabled, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (r *MonitorRepo) Get(ctx context.Context, id uuid.UUID) (*monitor.Monitor, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+monitorColumns+` FROM monitors WHERE id = $1`, id)
	m, err := scanMonitor(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, monitor.ErrNotFound
	}
	return m, err
}

func (r *MonitorRepo) List(ctx context.Context) ([]*monitor.Monitor, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+monitorColumns+` FROM monitors ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*monitor.Monitor, 0)
	for rows.Next() {
		m, err := scanMonitor(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MonitorRepo) Update(ctx context.Context, m *monitor.Monitor) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE monitors SET name=$2, type=$3, target=$4, interval_seconds=$5,
		 timeout_seconds=$6, expected_status=$7, enabled=$8, updated_at=$9 WHERE id=$1`,
		m.ID, m.Name, string(m.Type), m.Target,
		seconds(m.Interval), seconds(m.Timeout), m.ExpectedStatus,
		m.Enabled, m.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return monitor.ErrNotFound
	}
	return nil
}

func (r *MonitorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM monitors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return monitor.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMonitor(s rowScanner) (*monitor.Monitor, error) {
	var (
		m         monitor.Monitor
		typ       string
		intervalS int
		timeoutS  int
	)
	if err := s.Scan(&m.ID, &m.Name, &typ, &m.Target, &intervalS, &timeoutS,
		&m.ExpectedStatus, &m.Enabled, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.Type = monitor.CheckType(typ)
	m.Interval = time.Duration(intervalS) * time.Second
	m.Timeout = time.Duration(timeoutS) * time.Second
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return &m, nil
}

func seconds(d time.Duration) int {
	return int(d / time.Second)
}
