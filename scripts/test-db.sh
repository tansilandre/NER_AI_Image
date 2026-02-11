#!/bin/bash
# Test database connection script

echo "=== NER Studio Database Connection Test ==="
echo ""

# Check if .env exists
if [ ! -f "apps/api/.env" ]; then
    echo "❌ .env file not found!"
    echo "   Please create apps/api/.env from apps/api/.env.example"
    exit 1
fi

# Load environment
set -a
source apps/api/.env
set +a

echo "Database URL: ${DATABASE_URL:0:50}..."
echo ""

# Test connection using psql
echo "Testing PostgreSQL connection..."
if command -v psql &> /dev/null; then
    if psql "$DATABASE_URL" -c "SELECT version();" > /dev/null 2>&1; then
        echo "✅ Database connection successful!"
    else
        echo "❌ Database connection failed!"
        echo ""
        echo "Troubleshooting:"
        echo "  1. Check if DATABASE_URL is correct"
        echo "  2. Ensure your IP is whitelisted in Supabase"
        echo "  3. Verify network connectivity"
        exit 1
    fi
else
    echo "⚠️  psql not installed, skipping direct test"
fi

echo ""
echo "Done!"
