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

var ErrNotFound = errors.New("not found")

func normalizePersona(w *models.Agent) {
	if w.Persona.Constraints == nil {
		w.Persona.Constraints = []string{}
	}
	if w.Persona.Knowledge == nil {
		w.Persona.Knowledge = []string{}
	}
}

func (s *Store) CreateAgent(ctx context.Context, w models.Agent) (models.Agent, error) {
	normalizePersona(&w)
	id := uuid.NewString()
	personaJSON, err := json.Marshal(w.Persona)
	if err != nil {
		return models.Agent{}, fmt.Errorf("encode persona: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO agents (id, name, role, persona, model_binding_id, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, w.Name, w.Role, personaJSON, w.ModelBindingID, w.Enabled, now, now)
	if err != nil {
		return models.Agent{}, fmt.Errorf("insert agent: %w", err)
	}

	w.ID = id
	w.CreatedAt = now
	w.UpdatedAt = now
	return w, nil
}

func (s *Store) ListAgents(ctx context.Context) ([]models.Agent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, role, persona, model_binding_id, enabled, created_at, updated_at
		FROM agents ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	out := []models.Agent{}
	for rows.Next() {
		w, err := scanAgent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) GetAgent(ctx context.Context, id string) (models.Agent, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, role, persona, model_binding_id, enabled, created_at, updated_at
		FROM agents WHERE id = $1
	`, id)
	w, err := scanAgent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Agent{}, ErrNotFound
	}
	return w, err
}

func (s *Store) UpdateAgent(ctx context.Context, w models.Agent) (models.Agent, error) {
	normalizePersona(&w)
	personaJSON, err := json.Marshal(w.Persona)
	if err != nil {
		return models.Agent{}, fmt.Errorf("encode persona: %w", err)
	}
	now := time.Now()

	res, err := s.db.ExecContext(ctx, `
		UPDATE agents SET name = $1, role = $2, persona = $3, model_binding_id = $4, enabled = $5, updated_at = $6
		WHERE id = $7
	`, w.Name, w.Role, personaJSON, w.ModelBindingID, w.Enabled, now, w.ID)
	if err != nil {
		return models.Agent{}, fmt.Errorf("update agent: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Agent{}, ErrNotFound
	}
	w.UpdatedAt = now
	return w, nil
}

func (s *Store) DeleteAgent(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM agents WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAgent(row rowScanner) (models.Agent, error) {
	var w models.Agent
	var personaJSON []byte
	if err := row.Scan(&w.ID, &w.Name, &w.Role, &personaJSON, &w.ModelBindingID, &w.Enabled, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return models.Agent{}, err
	}
	if err := json.Unmarshal(personaJSON, &w.Persona); err != nil {
		return models.Agent{}, fmt.Errorf("decode persona: %w", err)
	}
	if w.Persona.Constraints == nil {
		w.Persona.Constraints = []string{}
	}
	if w.Persona.Knowledge == nil {
		w.Persona.Knowledge = []string{}
	}
	return w, nil
}
