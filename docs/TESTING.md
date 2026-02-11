# Testing Guide

## Test Types

### 1. Unit Tests (Fast, Isolated)
- **No database required**
- Tests business logic in isolation
- Uses mocks for dependencies
- Runs in milliseconds

```bash
cd apps/api
go test ./... -short -v
```

### 2. Integration Tests (Requires Database)
- **Needs PostgreSQL**
- Tests database operations
- Tests repository layer
- Slower but more comprehensive

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
export DATABASE_URL=postgres://postgres:postgres@localhost:54322/ner_studio_test?sslmode=disable
go test ./internal/repository -v

# Stop test database
docker-compose -f docker-compose.test.yml down
```

### 3. E2E Tests (Full Application)
- Tests HTTP endpoints
- Requires running server
- Tests full request/response cycle
- Slowest but most realistic

```bash
# Start server
make dev-api &

# Run API tests (using curl, Postman, or similar)
./scripts/api-tests.sh
```

## Running Tests

### All Tests
```bash
# Unit tests only
make test-api

# With coverage
cd apps/api && go test ./... -cover

# Specific package
cd apps/api && go test ./internal/provider -v
```

### Database Setup for Testing

1. **Using Docker (Recommended)**
```bash
# Start test DB
docker-compose -f docker-compose.test.yml up -d

# Run migrations
psql $DATABASE_URL -f supabase/migrations/001_create_organizations.sql
# ... run all migrations

# Run tests
go test ./internal/repository -v

# Cleanup
docker-compose -f docker-compose.test.yml down
```

2. **Using Local PostgreSQL**
```bash
# Start PostgreSQL
brew services start postgresql  # macOS
# or
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15

# Create test database
createdb ner_studio_test

# Run tests
export DATABASE_URL=postgresql://postgres:postgres@localhost:5432/ner_studio_test
go test ./internal/repository -v
```

3. **Using Cloud PostgreSQL (Use with caution!)**
```bash
# Only for staging/development databases
export DATABASE_URL=your_postgresql_connection_string
go test ./internal/repository -v -run TestOrganization  # Run specific test
```

## Test Structure

```
apps/api/internal/
├── provider/
│   └── provider_test.go      # Provider factory tests
├── service/
│   ├── generation_test.go    # Generation logic tests
│   └── upload_test.go        # Upload validation tests
├── handler/
│   └── handler_test.go       # HTTP endpoint tests
├── middleware/
│   └── middleware_test.go    # Auth, CORS tests
├── external/
│   └── r2_test.go            # R2 client tests
└── repository/
    └── repository_test.go    # Database integration tests
```

## Writing Tests

### Unit Test Example
```go
func TestSplitPrompts(t *testing.T) {
    tests := []struct {
        input    string
        expected []string
    }{
        {"Prompt 1\n\nPrompt 2", []string{"Prompt 1", "Prompt 2"}},
    }
    
    for _, tt := range tests {
        result := splitPrompts(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}
```

### Integration Test Example
```go
func (s *RepositoryTestSuite) TestCreateOrganization() {
    org := &model.Organization{
        ID:   uuid.New(),
        Name: "Test Org",
        Slug: "test-org",
    }
    
    err := s.repo.CreateOrganization(s.ctx, org)
    require.NoError(s.T(), err)
    
    retrieved, err := s.repo.GetOrganization(s.ctx, org.ID)
    assert.Equal(s.T(), org.Name, retrieved.Name)
}
```

## CI/CD Testing

Tests run automatically on GitHub Actions:

1. **On every push**: Unit tests
2. **On pull request**: Unit + Integration tests
3. **On merge to main**: All tests + Docker build

See `.github/workflows/test.yml` for configuration.

## Test Coverage

Generate coverage report:
```bash
cd apps/api
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

View coverage report:
```bash
open apps/api/coverage.html
```

## Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Name tests descriptively**: `TestCreateGeneration_WithInvalidProvider`
3. **Use `require` for fatal errors**, `assert` for non-fatal
4. **Clean up test data** in integration tests
5. **Skip tests gracefully** when dependencies unavailable
