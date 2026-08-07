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

func (s *Store) CreateWebhook(ctx context.Context, wh models.Webhook) (models.Webhook, error) {
	id := uuid.NewString()
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO webhooks (id, agent_id, token, created_at)
		VALUES ($1, $2, $3, $4)
	`, id, wh.AgentID, wh.Token, now)
	if err != nil {
		return models.Webhook{}, fmt.Errorf("insert webhook: %w", err)
	}
	wh.ID = id
	wh.CreatedAt = now
	return wh, nil
}

func (s *Store) ListWebhooksForAgent(ctx context.Context, agentID string) ([]models.Webhook, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, agent_id, token, created_at FROM webhooks WHERE agent_id = $1 ORDER BY created_at ASC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("query webhooks: %w", err)
	}
	defer rows.Close()

	out := []models.Webhook{}
	for rows.Next() {
		wh, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, wh)
	}
	return out, rows.Err()
}

func (s *Store) GetWebhookByToken(ctx context.Context, token string) (models.Webhook, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, agent_id, token, created_at FROM webhooks WHERE token = $1`, token)
	wh, err := scanWebhook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Webhook{}, ErrNotFound
	}
	return wh, err
}

func (s *Store) DeleteWebhook(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanWebhook(row rowScanner) (models.Webhook, error) {
	var wh models.Webhook
	if err := row.Scan(&wh.ID, &wh.AgentID, &wh.Token, &wh.CreatedAt); err != nil {
		return models.Webhook{}, err
	}
	return wh, nil
}
