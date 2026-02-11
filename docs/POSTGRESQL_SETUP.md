# PostgreSQL Setup Guide

Complete guide for setting up PostgreSQL for NER Studio.

## Why PostgreSQL (Not Supabase)?

- ✅ **Simple**: Direct connection, no pooler config
- ✅ **Flexible**: Works with any provider (local, Railway, Neon, AWS, etc.)
- ✅ **Standard**: Standard pgx driver, no special handling
- ✅ **Portable**: Easy to migrate between providers

## Recommended Providers

### 1. Railway (Easiest for Production)

```bash
# 1. Go to https://railway.app
# 2. New Project → Provision PostgreSQL
# 3. Copy connection string from "Connect" tab

# Example connection string:
DATABASE_URL=postgresql://postgres:password@containers.railway.app:5432/railway
```

### 2. Neon (Best Free Tier)

```bash
# 1. Go to https://neon.tech
# 2. Create project
# 3. Copy connection string from dashboard

# Example:
DATABASE_URL=postgresql://user:pass@ep-xxx.us-east-1.aws.neon.tech/neondb
```

### 3. Local Development

```bash
# macOS
brew install postgresql
brew services start postgresql

# Ubuntu
sudo apt-get install postgresql
sudo service postgresql start

# Create database
sudo -u postgres createdb ner_studio

# Connection string
DATABASE_URL=postgresql://postgres:password@localhost:5432/ner_studio
```

### 4. Docker

```bash
docker run -d \
  --name ner-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=ner_studio \
  -p 5432:5432 \
  postgres:15

# Connection string
DATABASE_URL=postgresql://postgres:password@localhost:5432/ner_studio
```

## Setup Steps

### 1. Create Database

```bash
# If using local PostgreSQL
createdb ner_studio

# Or use your provider's dashboard/UI
```

### 2. Configure .env

```bash
cp apps/api/.env.example apps/api/.env
# Edit apps/api/.env
```

Add your connection string:
```env
DATABASE_URL=postgresql://user:password@host:5432/ner_studio
JWT_SECRET=your-secret-key-min-32-characters
```

### 3. Run Migrations

```bash
cd apps/api
go run cmd/migrate/main.go
```

This creates all 8 tables:
- users
- organizations
- profiles
- generations
- generation_images
- credit_ledger
- invitations
- providers

### 4. Verify

```bash
# Test connection
go run cmd/dbtest/main.go

# Expected output:
# ✅ PostgreSQL connection successful!
# 📦 PostgreSQL version: PostgreSQL 16.x
# Tables found:
#   ✓ users
#   ✓ organizations
#   ...
```

### 5. Start Server

```bash
make dev-api
```

## Common Issues

### "password authentication failed"

Check your DATABASE_URL:
```bash
# Test connection manually
psql $DATABASE_URL -c "SELECT version();"
```

### "database does not exist"

Create the database:
```bash
# Local PostgreSQL
psql -c "CREATE DATABASE ner_studio;"

# Or via your provider's dashboard
```

### "permission denied for schema public"

Grant permissions:
```sql
GRANT CREATE ON SCHEMA public TO your_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_user;
```

## Migration Utility

We provide a migration tool at `apps/api/cmd/migrate/main.go`:

```bash
cd apps/api
go run cmd/migrate/main.go
```

Features:
- Connects to your database
- Runs all migration files in order
- Skips already-applied migrations
- Shows progress and verification

## Connection Testing

Test your connection anytime:

```bash
cd apps/api
go run cmd/dbtest/main.go
```

## Next Steps

1. ✅ Database connected
2. ✅ Tables created
3. ✅ Server running
4. → Start developing or deploy!

See `docs/DATABASE_SCHEMA.md` for table details.
