package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"myagents/internal/models"
)

func (s *Store) CreateSchedule(ctx context.Context, sc models.Schedule) (models.Schedule, error) {
	id := uuid.NewString()
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO schedules (id, agent_id, cron_expr, input, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, sc.AgentID, sc.CronExpr, sc.Input, sc.Enabled, now, now)
	if err != nil {
		return models.Schedule{}, fmt.Errorf("insert schedule: %w", err)
	}
	sc.ID = id
	sc.CreatedAt = now
	sc.UpdatedAt = now
	return sc, nil
}

func (s *Store) ListSchedulesForAgent(ctx context.Context, agentID string) ([]models.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, cron_expr, input, enabled, created_at, updated_at
		FROM schedules WHERE agent_id = $1 ORDER BY created_at ASC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()
	return scanSchedules(rows)
}

// ListEnabledSchedules returns every enabled schedule, for the scheduler to
// register on startup.
func (s *Store) ListEnabledSchedules(ctx context.Context) ([]models.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, cron_expr, input, enabled, created_at, updated_at
		FROM schedules WHERE enabled = true
	`)
	if err != nil {
		return nil, fmt.Errorf("query enabled schedules: %w", err)
	}
	defer rows.Close()
	return scanSchedules(rows)
}

func (s *Store) GetSchedule(ctx context.Context, id string) (models.Schedule, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, agent_id, cron_expr, input, enabled, created_at, updated_at
		FROM schedules WHERE id = $1
	`, id)
	sc, err := scanSchedule(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Schedule{}, ErrNotFound
	}
	return sc, err
}

func (s *Store) UpdateSchedule(ctx context.Context, id string, sc models.Schedule) (models.Schedule, error) {
	now := time.Now()
	res, err := s.db.ExecContext(ctx, `
		UPDATE schedules SET cron_expr = $1, input = $2, enabled = $3, updated_at = $4
		WHERE id = $5
	`, sc.CronExpr, sc.Input, sc.Enabled, now, id)
	if err != nil {
		return models.Schedule{}, fmt.Errorf("update schedule: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Schedule{}, ErrNotFound
	}
	sc.ID = id
	sc.UpdatedAt = now
	return sc, nil
}

func (s *Store) DeleteSchedule(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanSchedules(rows *sql.Rows) ([]models.Schedule, error) {
	out := []models.Schedule{}
	for rows.Next() {
		sc, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func scanSchedule(row rowScanner) (models.Schedule, error) {
	var sc models.Schedule
	if err := row.Scan(&sc.ID, &sc.AgentID, &sc.CronExpr, &sc.Input, &sc.Enabled, &sc.CreatedAt, &sc.UpdatedAt); err != nil {
		return models.Schedule{}, err
	}
	return sc, nil
}
