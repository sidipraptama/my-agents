package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"myagents/internal/models"
)

func (s *Store) CreateTask(ctx context.Context, t models.Task) (models.Task, error) {
	id := uuid.NewString()
	if t.Transcript == nil {
		t.Transcript = []models.TranscriptMessage{}
	}
	transcriptJSON, err := json.Marshal(t.Transcript)
	if err != nil {
		return models.Task{}, fmt.Errorf("encode transcript: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO tasks (id, agent_id, trigger_type, status, input, transcript, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, t.AgentID, t.TriggerType, t.Status, t.Input, transcriptJSON, now, now)
	if err != nil {
		return models.Task{}, fmt.Errorf("insert task: %w", err)
	}

	t.ID = id
	t.CreatedAt = now
	t.UpdatedAt = now
	return t, nil
}

func (s *Store) UpdateTaskResult(ctx context.Context, id string, status models.TaskStatus, transcript []models.TranscriptMessage) error {
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		return fmt.Errorf("encode transcript: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE tasks SET status = $1, transcript = $2, updated_at = $3 WHERE id = $4
	`, status, transcriptJSON, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (s *Store) ListTasksForAgent(ctx context.Context, agentID string) ([]models.Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, trigger_type, status, input, transcript, created_at, updated_at
		FROM tasks WHERE agent_id = $1 ORDER BY created_at ASC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	out := []models.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListAllTasks returns the most recent tasks across every Agent, for the task
// board. Capped at 200 rows — this is a dashboard view, not an export.
func (s *Store) ListAllTasks(ctx context.Context) ([]models.TaskWithAgent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tasks.id, tasks.agent_id, tasks.trigger_type, tasks.status, tasks.input,
		       tasks.transcript, tasks.created_at, tasks.updated_at, agents.name
		FROM tasks
		JOIN agents ON agents.id = tasks.agent_id
		ORDER BY tasks.created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, fmt.Errorf("query all tasks: %w", err)
	}
	defer rows.Close()

	out := []models.TaskWithAgent{}
	for rows.Next() {
		var t models.TaskWithAgent
		var transcriptJSON []byte
		if err := rows.Scan(&t.ID, &t.AgentID, &t.TriggerType, &t.Status, &t.Input,
			&transcriptJSON, &t.CreatedAt, &t.UpdatedAt, &t.AgentName); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		if err := json.Unmarshal(transcriptJSON, &t.Transcript); err != nil {
			return nil, fmt.Errorf("decode transcript: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetTask(ctx context.Context, id string) (models.Task, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, agent_id, trigger_type, status, input, transcript, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Task{}, ErrNotFound
	}
	return t, err
}

func scanTask(row rowScanner) (models.Task, error) {
	var t models.Task
	var transcriptJSON []byte
	if err := row.Scan(&t.ID, &t.AgentID, &t.TriggerType, &t.Status, &t.Input, &transcriptJSON, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return models.Task{}, err
	}
	if err := json.Unmarshal(transcriptJSON, &t.Transcript); err != nil {
		return models.Task{}, fmt.Errorf("decode transcript: %w", err)
	}
	return t, nil
}
