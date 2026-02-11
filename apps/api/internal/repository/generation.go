package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ner-studio/api/internal/model"
)

// CreateGeneration creates a new generation record
func (r *Repository) CreateGeneration(ctx context.Context, gen *model.Generation) error {
	query := `
		INSERT INTO generations (
			id, organization_id, user_id, status, base_prompt, 
			reference_images, product_images, provider_id, 
			estimated_cost, actual_cost, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		gen.ID,
		gen.OrganizationID,
		gen.UserID,
		gen.Status,
		gen.BasePrompt,
		gen.ReferenceImages,
		gen.ProductImages,
		gen.ProviderID,
		gen.EstimatedCost,
		gen.ActualCost,
	).Scan(&gen.CreatedAt, &gen.UpdatedAt)
}

// GetGeneration retrieves a generation by ID
func (r *Repository) GetGeneration(ctx context.Context, id uuid.UUID) (*model.Generation, error) {
	query := `
		SELECT id, organization_id, user_id, status, base_prompt,
			reference_images, product_images, provider_id,
			estimated_cost, actual_cost, error_message,
			created_at, updated_at, completed_at
		FROM generations
		WHERE id = $1
	`

	var gen model.Generation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&gen.ID,
		&gen.OrganizationID,
		&gen.UserID,
		&gen.Status,
		&gen.BasePrompt,
		&gen.ReferenceImages,
		&gen.ProductImages,
		&gen.ProviderID,
		&gen.EstimatedCost,
		&gen.ActualCost,
		&gen.ErrorMessage,
		&gen.CreatedAt,
		&gen.UpdatedAt,
		&gen.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &gen, nil
}

// ListGenerations lists generations for an organization
func (r *Repository) ListGenerations(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*model.Generation, error) {
	query := `
		SELECT id, organization_id, user_id, status, base_prompt,
			reference_images, product_images, provider_id,
			estimated_cost, actual_cost, error_message,
			created_at, updated_at, completed_at
		FROM generations
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var generations []*model.Generation
	for rows.Next() {
		var gen model.Generation
		err := rows.Scan(
			&gen.ID,
			&gen.OrganizationID,
			&gen.UserID,
			&gen.Status,
			&gen.BasePrompt,
			&gen.ReferenceImages,
			&gen.ProductImages,
			&gen.ProviderID,
			&gen.EstimatedCost,
			&gen.ActualCost,
			&gen.ErrorMessage,
			&gen.CreatedAt,
			&gen.UpdatedAt,
			&gen.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		generations = append(generations, &gen)
	}

	return generations, nil
}

// UpdateGenerationStatus updates generation status
func (r *Repository) UpdateGenerationStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	var completedAt interface{}
	if status == "completed" || status == "failed" {
		completedAt = "NOW()"
	} else {
		completedAt = nil
	}

	query := `
		UPDATE generations
		SET status = $2, error_message = $3, completed_at = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, status, errorMsg, completedAt)
	return err
}

// CreateGenerationImage creates a generation image record
func (r *Repository) CreateGenerationImage(ctx context.Context, img *model.GenerationImage) error {
	query := `
		INSERT INTO generation_images (
			id, generation_id, prompt, status, task_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		img.ID,
		img.GenerationID,
		img.Prompt,
		img.Status,
		img.TaskID,
	).Scan(&img.CreatedAt, &img.UpdatedAt)
}

// GetGenerationImageByTaskID retrieves an image by task ID
func (r *Repository) GetGenerationImageByTaskID(ctx context.Context, taskID string) (*model.GenerationImage, error) {
	query := `
		SELECT id, generation_id, prompt, image_url, r2_key, status, task_id,
			error_message, created_at, updated_at, completed_at
		FROM generation_images
		WHERE task_id = $1
	`

	var img model.GenerationImage
	err := r.pool.QueryRow(ctx, query, taskID).Scan(
		&img.ID,
		&img.GenerationID,
		&img.Prompt,
		&img.ImageURL,
		&img.R2Key,
		&img.Status,
		&img.TaskID,
		&img.ErrorMessage,
		&img.CreatedAt,
		&img.UpdatedAt,
		&img.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &img, nil
}

// UpdateGenerationImageComplete updates image on callback
func (r *Repository) UpdateGenerationImageComplete(ctx context.Context, id uuid.UUID, imageURL, r2Key string) error {
	query := `
		UPDATE generation_images
		SET status = 'completed', image_url = $2, r2_key = $3,
			completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, imageURL, r2Key)
	return err
}

// UpdateGenerationImageFailed marks image as failed
func (r *Repository) UpdateGenerationImageFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	query := `
		UPDATE generation_images
		SET status = 'failed', error_message = $2,
			completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, errorMsg)
	return err
}

// ListGenerationImages retrieves images for a generation
func (r *Repository) ListGenerationImages(ctx context.Context, generationID uuid.UUID) ([]*model.GenerationImage, error) {
	query := `
		SELECT id, generation_id, prompt, image_url, r2_key, status, task_id,
			error_message, created_at, updated_at, completed_at
		FROM generation_images
		WHERE generation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, generationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*model.GenerationImage
	for rows.Next() {
		var img model.GenerationImage
		err := rows.Scan(
			&img.ID,
			&img.GenerationID,
			&img.Prompt,
			&img.ImageURL,
			&img.R2Key,
			&img.Status,
			&img.TaskID,
			&img.ErrorMessage,
			&img.CreatedAt,
			&img.UpdatedAt,
			&img.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, &img)
	}

	return images, nil
}

// CountCompletedImages counts completed images for a generation
func (r *Repository) CountCompletedImages(ctx context.Context, generationID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM generation_images
		WHERE generation_id = $1 AND status = 'completed'
	`
	err := r.pool.QueryRow(ctx, query, generationID).Scan(&count)
	return count, err
}

// GetGenerationStats retrieves stats for updating generation status
func (r *Repository) GetGenerationStats(ctx context.Context, generationID uuid.UUID) (total, completed, failed int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM generation_images
		WHERE generation_id = $1
	`
	err = r.pool.QueryRow(ctx, query, generationID).Scan(&total, &completed, &failed)
	return
}

// UpdateGenerationActualCost updates the actual cost
func (r *Repository) UpdateGenerationActualCost(ctx context.Context, id uuid.UUID, cost int64) error {
	query := `UPDATE generations SET actual_cost = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, cost)
	return err
}

// Transaction wrapper for generation creation
func (r *Repository) CreateGenerationWithImages(ctx context.Context, gen *model.Generation, images []*model.GenerationImage) error {
	return r.WithTx(ctx, func(tx pgx.Tx) error {
		// Insert generation
		query := `
			INSERT INTO generations (
				id, organization_id, user_id, status, base_prompt,
				reference_images, product_images, provider_id,
				estimated_cost, actual_cost, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		`
		_, err := tx.Exec(ctx, query,
			gen.ID, gen.OrganizationID, gen.UserID, gen.Status,
			gen.BasePrompt, gen.ReferenceImages, gen.ProductImages,
			gen.ProviderID, gen.EstimatedCost, gen.ActualCost,
		)
		if err != nil {
			return fmt.Errorf("failed to insert generation: %w", err)
		}

		// Insert images
		for _, img := range images {
			query := `
				INSERT INTO generation_images (
					id, generation_id, prompt, status, task_id, created_at, updated_at
				)
				VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			`
			_, err := tx.Exec(ctx, query,
				img.ID, img.GenerationID, img.Prompt, img.Status, img.TaskID,
			)
			if err != nil {
				return fmt.Errorf("failed to insert image: %w", err)
			}
		}

		return nil
	})
}
