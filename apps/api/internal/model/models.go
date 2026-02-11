package model

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a billing entity
type Organization struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Slug          string    `json:"slug" db:"slug"`
	Credits       int64     `json:"credits" db:"credits"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Profile represents a user within an organization
type Profile struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	FullName       string    `json:"full_name" db:"full_name"`
	AvatarURL      string    `json:"avatar_url" db:"avatar_url"`
	Role           string    `json:"role" db:"role"` // admin, member
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Provider represents an AI service configuration
type Provider struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Slug        string          `json:"slug" db:"slug"`
	Name        string          `json:"name" db:"name"`
	Category    string          `json:"category" db:"category"` // image_generation, llm, vision
	APIKey      string          `json:"-" db:"api_key"`         // encrypted
	BaseURL     string          `json:"base_url" db:"base_url"`
	Model       string          `json:"model" db:"model"`
	Config      ProviderConfig  `json:"config" db:"config"`
	Priority    int             `json:"priority" db:"priority"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	CostPerUse  int64           `json:"cost_per_use" db:"cost_per_use"` // credits per use
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// ProviderConfig holds provider-specific settings
type ProviderConfig struct {
	TimeoutMs            int      `json:"timeout_ms,omitempty"`
	MaxRetries           int      `json:"max_retries,omitempty"`
	ErrorCodeForFallback []string `json:"error_code_for_fallback,omitempty"`
	Headers              map[string]string `json:"headers,omitempty"`
}

// Generation represents an image generation request
type Generation struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Status          string    `json:"status" db:"status"` // pending, processing, completed, failed
	BasePrompt      string    `json:"base_prompt" db:"base_prompt"`
	ReferenceImages []string  `json:"reference_images" db:"reference_images"`
	ProductImages   []string  `json:"product_images" db:"product_images"`
	ProviderID      uuid.UUID `json:"provider_id" db:"provider_id"`
	EstimatedCost   int64     `json:"estimated_cost" db:"estimated_cost"`
	ActualCost      int64     `json:"actual_cost" db:"actual_cost"`
	ErrorMessage    string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// GenerationImage represents a single generated image
type GenerationImage struct {
	ID            uuid.UUID `json:"id" db:"id"`
	GenerationID  uuid.UUID `json:"generation_id" db:"generation_id"`
	Prompt        string    `json:"prompt" db:"prompt"`
	ImageURL      string    `json:"image_url" db:"image_url"`
	R2Key         string    `json:"r2_key" db:"r2_key"`
	Status        string    `json:"status" db:"status"` // pending, processing, completed, failed
	TaskID        string    `json:"task_id" db:"task_id"` // provider task ID
	ErrorMessage  string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// CreditLedger tracks all credit transactions
type CreditLedger struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Amount         int64     `json:"amount" db:"amount"` // positive = add, negative = deduct
	Type           string    `json:"type" db:"type"`     // generation, refund, purchase
	Description    string    `json:"description" db:"description"`
	GenerationID   *uuid.UUID `json:"generation_id,omitempty" db:"generation_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Invitation represents a pending org invitation
type Invitation struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Email          string    `json:"email" db:"email"`
	Role           string    `json:"role" db:"role"`
	InvitedBy      uuid.UUID `json:"invited_by" db:"invited_by"`
	Token          string    `json:"-" db:"token"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	UsedAt         *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// VisionAnalysisResult holds the output from vision analysis
type VisionAnalysisResult struct {
	Description string `json:"description"`
	StyleNotes  string `json:"style_notes"`
}
