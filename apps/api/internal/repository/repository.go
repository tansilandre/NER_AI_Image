package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ner-studio/api/internal/model"
)

// Repository provides data access layer
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new repository
func NewRepository(databaseURL string) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &Repository{pool: pool}, nil
}

// Close closes the connection pool
func (r *Repository) Close() {
	r.pool.Close()
}

// Ping checks database connectivity
func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// WithTx executes a function within a transaction
func (r *Repository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetProfileByUserID retrieves a profile by user ID
func (r *Repository) GetProfileByUserID(ctx context.Context, userID string) (*model.Profile, error) {
	query := `
		SELECT id, user_id, organization_id, full_name, avatar_url, role, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`
	
	var profile model.Profile
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.OrganizationID,
		&profile.FullName,
		&profile.AvatarURL,
		&profile.Role,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// CreateOrganization creates a new organization
func (r *Repository) CreateOrganization(ctx context.Context, org *model.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, credits, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	
	return r.pool.QueryRow(ctx, query, org.ID, org.Name, org.Slug, org.Credits).Scan(
		&org.CreatedAt,
		&org.UpdatedAt,
	)
}

// CreateProfile creates a new profile
func (r *Repository) CreateProfile(ctx context.Context, profile *model.Profile) error {
	query := `
		INSERT INTO profiles (id, user_id, organization_id, full_name, avatar_url, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	
	return r.pool.QueryRow(ctx, query,
		profile.ID,
		profile.UserID,
		profile.OrganizationID,
		profile.FullName,
		profile.AvatarURL,
		profile.Role,
	).Scan(&profile.CreatedAt, &profile.UpdatedAt)
}

// GetOrganization retrieves an organization by ID
func (r *Repository) GetOrganization(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	query := `
		SELECT id, name, slug, credits, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	
	var org model.Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Credits,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &org, nil
}

// DeductCredits atomically deducts credits from an organization
func (r *Repository) DeductCredits(ctx context.Context, orgID uuid.UUID, amount int64, description string, userID uuid.UUID, generationID *uuid.UUID) error {
	return r.WithTx(ctx, func(tx pgx.Tx) error {
		// Lock the organization row
		var currentCredits int64
		err := tx.QueryRow(ctx, 
			"SELECT credits FROM organizations WHERE id = $1 FOR UPDATE",
			orgID,
		).Scan(&currentCredits)
		if err != nil {
			return fmt.Errorf("failed to lock organization: %w", err)
		}

		if currentCredits < amount {
			return fmt.Errorf("insufficient credits: have %d, need %d", currentCredits, amount)
		}

		// Deduct credits
		_, err = tx.Exec(ctx,
			"UPDATE organizations SET credits = credits - $1, updated_at = NOW() WHERE id = $2",
			amount, orgID,
		)
		if err != nil {
			return fmt.Errorf("failed to deduct credits: %w", err)
		}

		// Record in ledger
		ledgerID := uuid.New()
		_, err = tx.Exec(ctx,
			`INSERT INTO credit_ledger (id, organization_id, user_id, amount, type, description, generation_id, created_at)
			 VALUES ($1, $2, $3, $4, 'generation', $5, $6, NOW())`,
			ledgerID, orgID, userID, -amount, description, generationID,
		)
		if err != nil {
			return fmt.Errorf("failed to record ledger entry: %w", err)
		}

		return nil
	})
}
