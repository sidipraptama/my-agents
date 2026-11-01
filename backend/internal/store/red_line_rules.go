package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"myagents/internal/models"
)

func (s *Store) CreateRedLineRule(ctx context.Context, rule models.RedLineRule) (models.RedLineRule, error) {
	id := uuid.NewString()
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO red_line_rules (id, agent_id, pattern, created_at)
		VALUES ($1, $2, $3, $4)
	`, id, rule.AgentID, rule.Pattern, now)
	if err != nil {
		return models.RedLineRule{}, fmt.Errorf("insert red line rule: %w", err)
	}
	rule.ID = id
	rule.CreatedAt = now
	return rule, nil
}

func (s *Store) ListRedLineRulesForAgent(ctx context.Context, agentID string) ([]models.RedLineRule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, pattern, created_at FROM red_line_rules WHERE agent_id = $1 ORDER BY created_at ASC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("query red line rules: %w", err)
	}
	defer rows.Close()

	out := []models.RedLineRule{}
	for rows.Next() {
		var r models.RedLineRule
		if err := rows.Scan(&r.ID, &r.AgentID, &r.Pattern, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan red line rule: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) DeleteRedLineRule(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM red_line_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete red line rule: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
