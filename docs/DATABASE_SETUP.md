# Database Setup Guide

NER Studio uses **standard PostgreSQL** (any provider works).

## Quick Start

### 1. Get PostgreSQL Database

Choose any PostgreSQL provider:

#### Option A: Local PostgreSQL
```bash
# macOS
brew install postgresql
brew services start postgresql

# Create database
createdb ner_studio

# Connection string
DATABASE_URL=postgresql://localhost:5432/ner_studio
```

#### Option B: Railway (Easiest)
1. Go to https://railway.app
2. Create project → Add PostgreSQL
3. Copy connection string from "Connect" tab

#### Option C: Neon (Serverless)
1. Go to https://neon.tech
2. Create project
3. Copy connection string

#### Option D: AWS RDS / DigitalOcean / Any Provider
Any PostgreSQL 14+ works. Just get the connection string.

### 2. Configure Environment

```bash
# apps/api/.env
DATABASE_URL=postgresql://user:password@host:5432/dbname
```

### 3. Run Migrations

```bash
cd apps/api
go run cmd/migrate/main.go
```

You should see:
```
✅ Connected!
Found 10 migration files
→ Running 000_create_users.sql... ✅ Success
→ Running 001_create_organizations.sql... ✅ Success
...
✅ Migration Complete!
Total tables: 8
```

### 4. Verify Setup

```bash
# Test connection
cd apps/api
go run cmd/dbtest/main.go

# Start server
make dev-api
```

You should see:
```
✅ Database connected successfully!
Server starting on port 5005
```

## Environment Variables

```env
# Required
DATABASE_URL=postgresql://user:password@host:5432/dbname

# Optional (for security)
JWT_SECRET=your-secret-key-here
```

## Connection String Examples

| Provider | Connection String Format |
|----------|-------------------------|
| Local | `postgresql://localhost:5432/ner_studio` |
| Railway | `postgresql://user:pass@containers-xx.railway.app:5432/railway` |
| Neon | `postgresql://user:pass@ep-xxx.us-east-1.aws.neon.tech/dbname` |
| AWS RDS | `postgresql://user:pass@xxx.us-east-1.rds.amazonaws.com:5432/dbname` |

## Troubleshooting

### "connection refused"
- PostgreSQL is not running
- Wrong host/port in connection string
- Firewall blocking connection

### "password authentication failed"
- Wrong username/password
- Check credentials in DATABASE_URL

### "database does not exist"
- Database hasn't been created yet
- Run: `createdb ner_studio` (local) or create via provider dashboard

### SSL/TLS errors
Add `?sslmode=require` to DATABASE_URL for providers that require SSL:
```
postgresql://.../dbname?sslmode=require
```

## Migration Files

All migrations are in `supabase/migrations/` (name kept for historical reasons, but works with any PostgreSQL):

1. `000_create_users.sql` - Users table
2. `001_create_organizations.sql` - Organizations table
3. `002_create_profiles.sql` - User profiles
4. `003_create_generations.sql` - Generation requests
5. `004_create_generation_images.sql` - Generated images
6. `005_create_credit_ledger.sql` - Credit transactions
7. `006_create_invitations.sql` - Org invitations
8. `007_create_providers.sql` - AI providers config
9. `008_create_rls_policies.sql` - Row level security (optional)
10. `010_seed_providers.sql` - Default providers

## Database Schema

See `docs/DATABASE_SCHEMA.md` for full schema documentation.
