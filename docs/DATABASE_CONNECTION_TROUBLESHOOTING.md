# Database Connection Troubleshooting

## Issue: Cannot Connect to Supabase

### Symptoms
- Direct connection (port 5432): `no route to host`
- Transaction pooler (port 6543): `Tenant or user not found`
- Session pooler (port 5432): `Tenant or user not found`

### Root Cause
The current environment cannot route to the Supabase database IPs directly.

## Solutions

### Solution 1: Enable Connection Pooler (Recommended)

1. Go to https://supabase.com/dashboard
2. Select your project: `rdkuodxdgcmcszogibdu`
3. Go to **Database** → **Connection Pooling**
4. Click **Enable Connection Pooling**
5. Wait 2-3 minutes for it to activate
6. Copy the "Transaction pooler" connection string
7. Update `apps/api/.env`:

```env
DATABASE_URL=postgresql://postgres.[PROJECT_REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres
```

### Solution 2: Use Supabase Local Development

```bash
# Install Supabase CLI
npm install -g supabase

# Login
supabase login

# Link to your project
supabase link --project-ref rdkuodxdgcmcszogibdu

# Start local database
supabase start

# Update .env to use local database
DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres

# Run migrations
psql $DATABASE_URL -f supabase/migrations/001_create_organizations.sql
# ... run all other migrations

# Start server
make dev-api
```

### Solution 3: Deploy to Server with DB Access

Deploy the backend to a service like Railway, Fly.io, or Render which has direct network access:

```bash
# Railway
railway login
railway init
railway up

# Fly.io
flyctl auth login
flyctl launch
flyctl deploy
```

### Solution 4: Use Supabase PostgREST API

Instead of direct PostgreSQL connection, modify the repository to use Supabase's REST API:

```go
import "github.com/supabase-community/supabase-go"

// Use supabase client instead of pgx
client, _ := supabase.NewClient(supabaseURL, supabaseKey, nil)
```

This requires refactoring the repository layer.

## Testing Connection

```bash
# Test with our utility
cd apps/api
go run cmd/dbtest/main.go

# Or test with psql (if installed)
psql $DATABASE_URL -c "SELECT version();"
```

## Current Configuration

```env
# Session Pooler (attempted)
DATABASE_URL=postgresql://postgres.rdkuodxdgcmcszogibdu:[PASSWORD]@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres

# Transaction Pooler (attempted)  
DATABASE_URL=postgresql://postgres.rdkuodxdgcmcszogibdu:[PASSWORD]@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres

# Direct connection (blocked by network)
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.rdkuodxdgcmcszogibdu.supabase.co:5432/postgres
```

## Next Steps

1. Try **Solution 2 (Supabase Local)** for immediate development
2. Or deploy to **Railway/Fly.io** for production-like environment
3. Once connected, run migrations: `./scripts/run-migrations.sh`
4. Test: `go run cmd/dbtest/main.go`
