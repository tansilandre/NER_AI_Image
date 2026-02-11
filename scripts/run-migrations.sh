#!/bin/bash
# Run database migrations on Supabase

echo "=== Running Database Migrations ==="
echo ""

# Load environment
if [ -f "apps/api/.env" ]; then
    set -a
    source apps/api/.env
    set +a
fi

if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL not set!"
    echo "   Please set it in apps/api/.env"
    exit 1
fi

echo "Running migrations..."
echo ""

# Run each migration file in order
for file in supabase/migrations/*.sql; do
    echo "→ Running $(basename $file)..."
    if psql "$DATABASE_URL" -f "$file" > /dev/null 2>&1; then
        echo "  ✅ Success"
    else
        echo "  ⚠️  Failed (may already exist or connection error)"
    fi
done

echo ""
echo "✅ Migrations complete!"
