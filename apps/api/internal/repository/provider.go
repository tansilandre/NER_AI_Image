package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/model"
)

// GetProvider retrieves a provider by ID
func (r *Repository) GetProvider(ctx context.Context, id uuid.UUID) (*model.Provider, error) {
	query := `
		SELECT id, slug, name, category, api_key, base_url, model, config,
			priority, is_active, cost_per_use, created_at, updated_at
		FROM providers
		WHERE id = $1
	`

	var provider model.Provider
	var configJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&provider.ID,
		&provider.Slug,
		&provider.Name,
		&provider.Category,
		&provider.APIKey,
		&provider.BaseURL,
		&provider.Model,
		&configJSON,
		&provider.Priority,
		&provider.IsActive,
		&provider.CostPerUse,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &provider.Config)
	}

	return &provider, nil
}

// GetProviderBySlug retrieves a provider by slug
func (r *Repository) GetProviderBySlug(ctx context.Context, slug string) (*model.Provider, error) {
	query := `
		SELECT id, slug, name, category, api_key, base_url, model, config,
			priority, is_active, cost_per_use, created_at, updated_at
		FROM providers
		WHERE slug = $1
	`

	var provider model.Provider
	var configJSON []byte
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&provider.ID,
		&provider.Slug,
		&provider.Name,
		&provider.Category,
		&provider.APIKey,
		&provider.BaseURL,
		&provider.Model,
		&configJSON,
		&provider.Priority,
		&provider.IsActive,
		&provider.CostPerUse,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &provider.Config)
	}

	return &provider, nil
}

// ListProviders lists providers by category
func (r *Repository) ListProviders(ctx context.Context, category string, onlyActive bool) ([]*model.Provider, error) {
	query := `
		SELECT id, slug, name, category, api_key, base_url, model, config,
			priority, is_active, cost_per_use, created_at, updated_at
		FROM providers
		WHERE ($1 = '' OR category = $1)
		AND ($2 = false OR is_active = true)
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, category, onlyActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*model.Provider
	for rows.Next() {
		var provider model.Provider
		var configJSON []byte
		err := rows.Scan(
			&provider.ID,
			&provider.Slug,
			&provider.Name,
			&provider.Category,
			&provider.APIKey,
			&provider.BaseURL,
			&provider.Model,
			&configJSON,
			&provider.Priority,
			&provider.IsActive,
			&provider.CostPerUse,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &provider.Config)
		}
		providers = append(providers, &provider)
	}

	return providers, nil
}

// CreateProvider creates a new provider
func (r *Repository) CreateProvider(ctx context.Context, provider *model.Provider) error {
	configJSON, _ := json.Marshal(provider.Config)

	query := `
		INSERT INTO providers (
			slug, name, category, api_key, base_url, model, config,
			priority, is_active, cost_per_use, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		provider.Slug,
		provider.Name,
		provider.Category,
		provider.APIKey,
		provider.BaseURL,
		provider.Model,
		configJSON,
		provider.Priority,
		provider.IsActive,
		provider.CostPerUse,
	).Scan(&provider.ID, &provider.CreatedAt, &provider.UpdatedAt)
}

// UpdateProvider updates a provider
func (r *Repository) UpdateProvider(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic query
	query := `UPDATE providers SET `
	args := []interface{}{}
	argCount := 0

	for field, value := range updates {
		if field == "config" {
			configJSON, _ := json.Marshal(value)
			argCount++
			query += fmt.Sprintf("%s = $%d, ", field, argCount)
			args = append(args, configJSON)
		} else {
			argCount++
			query += fmt.Sprintf("%s = $%d, ", field, argCount)
			args = append(args, value)
		}
	}

	argCount++
	query += fmt.Sprintf("updated_at = NOW() WHERE id = $%d", argCount)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// DeleteProvider deletes a provider
func (r *Repository) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM providers WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
