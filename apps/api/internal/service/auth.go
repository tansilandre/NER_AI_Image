package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/model"
	"github.com/ner-studio/api/internal/repository"
)

// AuthService handles authentication and user management
type AuthService struct {
	repo *repository.Repository
}

// NewAuthService creates a new auth service
func NewAuthService(repo *repository.Repository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

// GetOrCreateUser retrieves or creates a user after OAuth
func (s *AuthService) GetOrCreateUser(ctx context.Context, userID, email string) (*model.Profile, error) {
	// Try to get existing profile
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err == nil {
		return profile, nil
	}

	// Profile doesn't exist - user needs onboarding
	return nil, nil
}

// CreateOrganizationAndProfile creates a new org and admin profile
func (s *AuthService) CreateOrganizationAndProfile(ctx context.Context, userID, email, fullName, orgName string) (*model.Organization, *model.Profile, error) {
	// Create organization
	org := &model.Organization{
		ID:      uuid.New(),
		Name:    orgName,
		Slug:    generateSlug(orgName),
		Credits: 0,
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Create profile as admin
	profile := &model.Profile{
		ID:             uuid.New(),
		UserID:         uuid.MustParse(userID),
		OrganizationID: org.ID,
		FullName:       fullName,
		Role:           "admin",
	}

	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return nil, nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return org, profile, nil
}

// JoinOrganization joins user to existing org via invitation
func (s *AuthService) JoinOrganization(ctx context.Context, userID, fullName, inviteToken string) (*model.Organization, *model.Profile, error) {
	// TODO: Validate invitation token
	// TODO: Create profile in organization
	return nil, nil, fmt.Errorf("not implemented")
}

func generateSlug(name string) string {
	// Simple slug generation - replace with proper implementation
	slug := ""
	for _, c := range name {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c - 'A' + 'a')
		} else if c == ' ' || c == '-' {
			slug += "-"
		}
	}
	return slug
}
