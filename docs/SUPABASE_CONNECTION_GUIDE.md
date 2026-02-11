# Supabase Connection Guide

## Problem: "Tenant or user not found"

This error means the **connection pooler is not enabled** in your Supabase project.

## Solution 1: Enable Connection Pooling (Recommended)

### Step 1: Enable in Dashboard
1. Go to https://supabase.com/dashboard/project/rdkuodxdgcmcszogibdu
2. Click **Database** in left sidebar
3. Scroll down to **Connection Pooling**
4. Click **Enable Connection Pooling**
5. Wait 2-3 minutes for it to activate

### Step 2: Get Correct Connection String
After enabling, copy the **Session pooler** connection string:
```
postgresql://postgres.[PROJECT_REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:5432/postgres
```

### Step 3: Update .env
```env
DATABASE_URL=postgresql://postgres.rdkuodxdgcmcszogibdu:YOUR_PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres
```

## Solution 2: Use Supabase CLI (Local Dev)

If pooler can't be enabled, use Supabase CLI:

```bash
# Install
npm install -g supabase

# Login
supabase login

# Link project
supabase link --project-ref rdkuodxdgcmcszogibdu

# Start local database
supabase start

# Update .env to use local DB
DATABASE_URL=postgresql://postgres:postgres@localhost:54322/postgres
```

## Solution 3: Refactor to Use Supabase Client

Since Supabase Go client already works, we can modify the repository to use it instead of pgx:

```go
// Instead of pgx pool
client, _ := supabase.NewClient(url, key, nil)
result, _, err := client.From("organizations").Select("*", "exact", false).Execute()
```

This requires changing the repository layer but works immediately.

## Current Status

| Method | Works? | Issue |
|--------|--------|-------|
| Supabase Go Client | ✅ Yes | Tables don't exist (need migrations) |
| PostgreSQL Pooler | ❌ No | Pooler not enabled |
| Direct PostgreSQL | ❌ No | Network routing blocked |

## Quick Fix: Run Migrations via Dashboard

1. Go to https://supabase.com/dashboard/project/rdkuodxdgcmcszogibdu
2. Open **SQL Editor**
3. Run each migration file from `supabase/migrations/`
4. Tables will be created
5. Supabase Go Client will work

Then we can decide on PostgreSQL pooler vs Supabase Client approach.
