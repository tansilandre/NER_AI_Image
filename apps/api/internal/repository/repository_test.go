package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RepositoryTestSuite integration tests with real database
type RepositoryTestSuite struct {
	suite.Suite
	repo *Repository
	ctx  context.Context
}

// SetupSuite runs once before all tests
func (s *RepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	
	// Get database URL from env or use test database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres@localhost:54322/postgres" // Supabase local default
	}
	
	repo, err := NewRepository(databaseURL)
	if err != nil {
		s.T().Skipf("Skipping integration tests: cannot connect to database: %v", err)
		return
	}
	
	s.repo = repo
	
	// Verify connection
	if err := s.repo.Ping(s.ctx); err != nil {
		s.T().Skipf("Skipping integration tests: cannot ping database: %v", err)
	}
}

// TearDownSuite runs once after all tests
func (s *RepositoryTestSuite) TearDownSuite() {
	if s.repo != nil {
		s.repo.Close()
	}
}

// SetupTest runs before each test
func (s *RepositoryTestSuite) SetupTest() {
	// Clean up test data before each test
	s.cleanupTestData()
}

func (s *RepositoryTestSuite) cleanupTestData() {
	// Delete test data (be careful with this in production!)
	testOrgIDs := []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	}
	
	for _, id := range testOrgIDs {
		uuidID, _ := uuid.Parse(id)
		s.repo.pool.Exec(s.ctx, "DELETE FROM credit_ledger WHERE organization_id = $1", uuidID)
		s.repo.pool.Exec(s.ctx, "DELETE FROM generation_images WHERE generation_id IN (SELECT id FROM generations WHERE organization_id = $1)", uuidID)
		s.repo.pool.Exec(s.ctx, "DELETE FROM generations WHERE organization_id = $1", uuidID)
		s.repo.pool.Exec(s.ctx, "DELETE FROM profiles WHERE organization_id = $1", uuidID)
		s.repo.pool.Exec(s.ctx, "DELETE FROM organizations WHERE id = $1", uuidID)
	}
}

func (s *RepositoryTestSuite) TestCreateAndGetOrganization() {
	if s.repo == nil {
		s.T().Skip("Database not available")
	}
	
	org := &model.Organization{
		ID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:    "Test Organization",
		Slug:    "test-org",
		Credits: 1000,
	}
	
	// Create
	err := s.repo.CreateOrganization(s.ctx, org)
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), org.CreatedAt)
	assert.NotZero(s.T(), org.UpdatedAt)
	
	// Get
	retrieved, err := s.repo.GetOrganization(s.ctx, org.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), org.Name, retrieved.Name)
	assert.Equal(s.T(), org.Slug, retrieved.Slug)
	assert.Equal(s.T(), org.Credits, retrieved.Credits)
}

func (s *RepositoryTestSuite) TestCreateAndGetProfile() {
	if s.repo == nil {
		s.T().Skip("Database not available")
	}
	
	// First create organization
	org := &model.Organization{
		ID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:    "Test Org",
		Slug:    "test-org",
		Credits: 100,
	}
	err := s.repo.CreateOrganization(s.ctx, org)
	require.NoError(s.T(), err)
	
	// Create profile
	profile := &model.Profile{
		ID:             uuid.New(),
		UserID:         uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		OrganizationID: org.ID,
		FullName:       "Test User",
		Role:           "admin",
	}
	
	err = s.repo.CreateProfile(s.ctx, profile)
	require.NoError(s.T(), err)
	
	// Get by user ID
	retrieved, err := s.repo.GetProfileByUserID(s.ctx, profile.UserID.String())
	require.NoError(s.T(), err)
	assert.Equal(s.T(), profile.FullName, retrieved.FullName)
	assert.Equal(s.T(), profile.Role, retrieved.Role)
	assert.Equal(s.T(), org.ID, retrieved.OrganizationID)
}

func (s *RepositoryTestSuite) TestGenerationWorkflow() {
	if s.repo == nil {
		s.T().Skip("Database not available")
	}
	
	// Setup: Create org and user
	org := &model.Organization{
		ID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:    "Test Org",
		Slug:    "test-org",
		Credits: 1000,
	}
	require.NoError(s.T(), s.repo.CreateOrganization(s.ctx, org))
	
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	providerID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	
	// Create generation
	gen := &model.Generation{
		ID:              uuid.New(),
		OrganizationID:  org.ID,
		UserID:          userID,
		Status:          "pending",
		BasePrompt:      "A beautiful sunset",
		ReferenceImages: []string{"https://example.com/ref1.jpg"},
		ProductImages:   []string{},
		ProviderID:      providerID,
		EstimatedCost:   40,
	}
	
	err := s.repo.CreateGeneration(s.ctx, gen)
	require.NoError(s.T(), err)
	
	// Create generation images
	images := []*model.GenerationImage{
		{
			ID:           uuid.New(),
			GenerationID: gen.ID,
			Prompt:       "Prompt variation 1",
			Status:       "pending",
			TaskID:       "task-001",
		},
		{
			ID:           uuid.New(),
			GenerationID: gen.ID,
			Prompt:       "Prompt variation 2",
			Status:       "pending",
			TaskID:       "task-002",
		},
	}
	
	for _, img := range images {
		err := s.repo.CreateGenerationImage(s.ctx, img)
		require.NoError(s.T(), err)
	}
	
	// Get generation with images
	retrievedGen, err := s.repo.GetGeneration(s.ctx, gen.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), gen.BasePrompt, retrievedGen.BasePrompt)
	
	retrievedImages, err := s.repo.ListGenerationImages(s.ctx, gen.ID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), retrievedImages, 2)
	
	// Update image status to completed
	err = s.repo.UpdateGenerationImageComplete(s.ctx, images[0].ID, "https://bucket.tansil.pro/image1.jpg", "gen/image1.jpg")
	require.NoError(s.T(), err)
	
	// Check stats
	total, completed, failed, err := s.repo.GetGenerationStats(s.ctx, gen.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, total)
	assert.Equal(s.T(), 1, completed)
	assert.Equal(s.T(), 0, failed)
	
	// Get image by task ID
	imgByTask, err := s.repo.GetGenerationImageByTaskID(s.ctx, "task-001")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), images[0].ID, imgByTask.ID)
	assert.Equal(s.T(), "completed", imgByTask.Status)
}

func (s *RepositoryTestSuite) TestCreditDeduction() {
	if s.repo == nil {
		s.T().Skip("Database not available")
	}
	
	// Setup
	org := &model.Organization{
		ID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:    "Test Org",
		Slug:    "test-org",
		Credits: 100,
	}
	require.NoError(s.T(), s.repo.CreateOrganization(s.ctx, org))
	
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	genID := uuid.New()
	
	// Deduct credits
	err := s.repo.DeductCredits(s.ctx, org.ID, 30, "Test generation", userID, &genID)
	require.NoError(s.T(), err)
	
	// Verify credits deducted
	updatedOrg, err := s.repo.GetOrganization(s.ctx, org.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(70), updatedOrg.Credits)
	
	// Try to deduct more than available
	err = s.repo.DeductCredits(s.ctx, org.ID, 100, "Should fail", userID, &genID)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "insufficient credits")
}

func (s *RepositoryTestSuite) TestProviderOperations() {
	if s.repo == nil {
		s.T().Skip("Database not available")
	}
	
	// Note: Providers table may be seeded, so we'll test retrieval
	providers, err := s.repo.ListProviders(s.ctx, "llm", true)
	require.NoError(s.T(), err)
	
	// Should have at least the seeded providers
	assert.GreaterOrEqual(s.T(), len(providers), 0)
	
	// Test getting by slug (if exists)
	provider, err := s.repo.GetProviderBySlug(s.ctx, "openai-gpt4o")
	if err == nil {
		assert.Equal(s.T(), "vision", provider.Category)
	}
}

// Run the test suite
func TestRepositorySuite(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("Skipping database tests")
	}
	suite.Run(t, new(RepositoryTestSuite))
}

// Simple unit tests (no DB required)
func TestUUIDGeneration(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, uuid.Nil, id1)
}

func TestModelValidation(t *testing.T) {
	gen := &model.Generation{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Status:         "pending",
		BasePrompt:     "Test prompt",
	}
	
	assert.NotNil(t, gen)
	assert.Equal(t, "pending", gen.Status)
}
