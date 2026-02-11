package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/model"
	"github.com/ner-studio/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
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

// CreateUser creates a new user with hashed password
func (s *AuthService) CreateUser(ctx context.Context, email, password string) (*model.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.CreateUser(ctx, email, string(hashedPassword))
}

// AuthenticateUser verifies email/password and returns user
func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (*model.User, error) {
	// Get user from database
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	_ = s.repo.UpdateUserLastLogin(ctx, user.ID)

	return user, nil
}

// GetUserProfile retrieves user profile
func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*model.Profile, error) {
	return s.repo.GetProfileByUserID(ctx, userID)
}

// CreateOrganizationAndProfile creates a new org and admin profile
func (s *AuthService) CreateOrganizationAndProfile(ctx context.Context, email, fullName, orgName, password string) (*model.Organization, *model.Profile, error) {
	// Create user first
	user, err := s.CreateUser(ctx, email, password)
	if err != nil {
		return nil, nil, err
	}

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
		UserID:         user.ID,
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
	// Simple slug generation
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
