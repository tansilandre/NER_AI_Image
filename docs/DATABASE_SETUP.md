# Database Setup Guide

## Supabase Configuration

### 1. Create Supabase Project
1. Go to https://supabase.com
2. Create new project
3. Note down the database connection string

### 2. Get Connection String
1. In Supabase Dashboard → Settings → Database
2. Copy "Connection string" (URI format)
3. Format: `postgresql://postgres:[PASSWORD]@db.[PROJECT_ID].supabase.co:5432/postgres`

### 3. Configure Environment
```bash
# apps/api/.env
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.xxx.supabase.co:5432/postgres
```

### 4. Network Configuration
Supabase requires IP whitelisting for direct connections:

1. Go to Supabase Dashboard → Settings → Database
2. Under "IPv4", add your IP address
3. Or temporarily disable IP restrictions (not recommended for production)

### 5. Run Migrations

Option A: Using psql
```bash
# Install psql if needed
# macOS: brew install libpq
# Ubuntu: sudo apt-get install postgresql-client

# Run migrations
./scripts/run-migrations.sh
```

Option B: Using Supabase CLI
```bash
# Install Supabase CLI
npm install -g supabase

# Login
supabase login

# Link to project
supabase link --project-ref YOUR_PROJECT_ID

# Push migrations
supabase db push
```

Option C: Manual (Supabase Dashboard)
```bash
# Go to Supabase Dashboard → SQL Editor
# Copy contents of each migration file and execute
```

### 6. Seed Data
```bash
# Run seed file
psql $DATABASE_URL -f supabase/migrations/010_seed_providers.sql
```

### 7. Verify Setup
```bash
# Test connection
./scripts/test-db.sh

# Or manually
cd apps/api
go run cmd/server/main.go
```

You should see:
```
✅ Database connection successful!
Server starting on port 5005
```

## Troubleshooting

### "no route to host" or "connection refused"
- Your IP is not whitelisted in Supabase
- Solution: Add your IP in Supabase Dashboard → Settings → Database → IPv4

### "password authentication failed"
- Wrong password in DATABASE_URL
- Solution: Reset password in Supabase Dashboard

### "database does not exist"
- Wrong database name
- Solution: Use `postgres` as database name for Supabase

### SSL/TLS errors
Add `?sslmode=require` to DATABASE_URL:
```
postgresql://.../postgres?sslmode=require
```

## Alternative: Supabase Local Development

For local development without cloud connection:

```bash
# Install Supabase CLI
npm install -g supabase

# Start local Supabase
supabase start

# DATABASE_URL will be:
# postgresql://postgres:postgres@localhost:54322/postgres

# Stop when done
supabase stop
```

## Connection Pooling (Recommended for Production)

For production, use Supabase's connection pooler (PgBouncer):

1. In Supabase Dashboard → Database → Connection Pooling
2. Copy "Connection string" from Connection Pooler section
3. Use port 6543 instead of 5432
4. Update DATABASE_URL accordingly

This prevents "too many connections" errors under load.
