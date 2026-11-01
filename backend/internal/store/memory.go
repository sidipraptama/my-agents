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

func (s *Store) CreateMemoryEntry(ctx context.Context, m models.MemoryEntry) (models.MemoryEntry, error) {
	id := uuid.NewString()
	if m.EntryType == "" {
		m.EntryType = "fact"
	}
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO memory_entries (id, agent_id, entry_type, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, m.AgentID, m.EntryType, m.Content, now, now)
	if err != nil {
		return models.MemoryEntry{}, fmt.Errorf("insert memory entry: %w", err)
	}
	m.ID = id
	m.CreatedAt = now
	m.UpdatedAt = now
	return m, nil
}

// ListMemoryForAgent returns every memory entry for an agent, most recent first.
func (s *Store) ListMemoryForAgent(ctx context.Context, agentID string) ([]models.MemoryEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, entry_type, content, created_at, updated_at
		FROM memory_entries WHERE agent_id = $1 ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("query memory entries: %w", err)
	}
	defer rows.Close()

	out := []models.MemoryEntry{}
	for rows.Next() {
		var m models.MemoryEntry
		if err := rows.Scan(&m.ID, &m.AgentID, &m.EntryType, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan memory entry: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) UpdateMemoryEntry(ctx context.Context, id, entryType, content string) (models.MemoryEntry, error) {
	now := time.Now()
	row := s.db.QueryRowContext(ctx, `
		UPDATE memory_entries SET entry_type = $1, content = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, agent_id, entry_type, content, created_at, updated_at
	`, entryType, content, now, id)

	var m models.MemoryEntry
	err := row.Scan(&m.ID, &m.AgentID, &m.EntryType, &m.Content, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.MemoryEntry{}, ErrNotFound
	}
	if err != nil {
		return models.MemoryEntry{}, fmt.Errorf("update memory entry: %w", err)
	}
	return m, nil
}

func (s *Store) DeleteMemoryEntry(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM memory_entries WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete memory entry: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
