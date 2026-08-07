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

func (s *Store) CreateBinding(ctx context.Context, b models.ModelBinding) (models.ModelBinding, error) {
	id := uuid.NewString()
	encrypted, err := s.box.Encrypt(b.APIKey)
	if err != nil {
		return models.ModelBinding{}, fmt.Errorf("encrypt api key: %w", err)
	}
	if b.Config == nil {
		b.Config = map[string]any{}
	}
	cfgJSON, err := json.Marshal(b.Config)
	if err != nil {
		return models.ModelBinding{}, fmt.Errorf("encode config: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO model_bindings (id, provider, model_name, api_key_encrypted, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, b.Provider, b.ModelName, encrypted, cfgJSON, now)
	if err != nil {
		return models.ModelBinding{}, fmt.Errorf("insert model binding: %w", err)
	}

	b.ID = id
	b.CreatedAt = now
	b.APIKey = ""
	return b, nil
}

func (s *Store) ListBindings(ctx context.Context) ([]models.ModelBinding, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, provider, model_name, config, created_at FROM model_bindings ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query model bindings: %w", err)
	}
	defer rows.Close()

	out := []models.ModelBinding{}
	for rows.Next() {
		var b models.ModelBinding
		var cfgJSON []byte
		if err := rows.Scan(&b.ID, &b.Provider, &b.ModelName, &cfgJSON, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan model binding: %w", err)
		}
		if err := json.Unmarshal(cfgJSON, &b.Config); err != nil {
			return nil, fmt.Errorf("decode config: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// UpdateBinding updates provider, model name, and config. The API key is only
// re-encrypted and overwritten when b.APIKey is non-empty, so callers can edit
// other fields without having to resupply the secret.
func (s *Store) UpdateBinding(ctx context.Context, id string, b models.ModelBinding) (models.ModelBinding, error) {
	if b.Config == nil {
		b.Config = map[string]any{}
	}
	cfgJSON, err := json.Marshal(b.Config)
	if err != nil {
		return models.ModelBinding{}, fmt.Errorf("encode config: %w", err)
	}

	var row *sql.Row
	if b.APIKey != "" {
		encrypted, err := s.box.Encrypt(b.APIKey)
		if err != nil {
			return models.ModelBinding{}, fmt.Errorf("encrypt api key: %w", err)
		}
		row = s.db.QueryRowContext(ctx, `
			UPDATE model_bindings SET provider = $1, model_name = $2, config = $3, api_key_encrypted = $4
			WHERE id = $5
			RETURNING created_at
		`, b.Provider, b.ModelName, cfgJSON, encrypted, id)
	} else {
		row = s.db.QueryRowContext(ctx, `
			UPDATE model_bindings SET provider = $1, model_name = $2, config = $3
			WHERE id = $4
			RETURNING created_at
		`, b.Provider, b.ModelName, cfgJSON, id)
	}

	if err := row.Scan(&b.CreatedAt); errors.Is(err, sql.ErrNoRows) {
		return models.ModelBinding{}, ErrNotFound
	} else if err != nil {
		return models.ModelBinding{}, fmt.Errorf("update model binding: %w", err)
	}

	b.ID = id
	b.APIKey = ""
	return b, nil
}

func (s *Store) DeleteBinding(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM model_bindings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete model binding: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ResolveCredentials returns the provider name, model name, decrypted API key,
// and config (e.g. base_url) for a binding.
func (s *Store) ResolveCredentials(ctx context.Context, bindingID string) (provider, modelName, apiKey string, config map[string]any, err error) {
	var encrypted []byte
	var cfgJSON []byte
	row := s.db.QueryRowContext(ctx, `SELECT provider, model_name, api_key_encrypted, config FROM model_bindings WHERE id = $1`, bindingID)
	if err := row.Scan(&provider, &modelName, &encrypted, &cfgJSON); err != nil {
		return "", "", "", nil, fmt.Errorf("query model binding: %w", err)
	}
	apiKey, err = s.box.Decrypt(encrypted)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("decrypt api key: %w", err)
	}
	if err := json.Unmarshal(cfgJSON, &config); err != nil {
		return "", "", "", nil, fmt.Errorf("decode config: %w", err)
	}
	return provider, modelName, apiKey, config, nil
}
