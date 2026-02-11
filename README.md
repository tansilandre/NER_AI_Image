# NER Studio

AI Image Generation Platform for Creative Teams

## Quick Start

```bash
# Install dependencies
make deps

# Setup environment
cp apps/api/.env.example apps/api/.env
# Edit apps/api/.env with your database credentials

# Run migrations
cd apps/api && go run cmd/migrate/main.go

# Start development server
make dev-api
```

## Project Structure

```
ner-studio/
├── apps/
│   ├── api/          # Go + Fiber backend
│   └── web/          # React + Vite frontend (coming soon)
├── supabase/
│   └── migrations/   # Database migrations (for any PostgreSQL)
├── docs/             # Documentation
└── Makefile          # Common commands
```

## Backend (API)

The backend is built with Go and Fiber, featuring:

- **Simple JWT Auth**: No external auth provider needed
- **Provider System**: Pluggable AI providers (OpenAI, Kie.ai, Gemini)
- **Multi-tier LLM Fallback**: Automatic failover between providers
- **Credit-based Billing**: Per-model pricing with atomic deductions
- **Async Workflows**: Image generation pipeline with callbacks

### Requirements

- Go 1.23+
- PostgreSQL 14+ (local, Neon, Railway, AWS RDS, etc.)
- (Optional) Cloudflare R2 for image storage

### Environment Variables

See `apps/api/.env.example` for required environment variables.

Key variables:
```env
DATABASE_URL=postgresql://user:pass@host:5432/dbname
JWT_SECRET=your-secret-key
KIE_AI_API_KEY=...
OPENAI_API_KEY=...
```

## Database

Any PostgreSQL database works:

- **Local**: `postgresql://user:pass@localhost:5432/ner_studio`
- **Neon**: `postgresql://user:pass@ep-xxx.us-east-1.aws.neon.tech/dbname`
- **Railway**: `postgresql://user:pass@containers-xx.railway.app:5432/railway`
- **AWS RDS**: `postgresql://user:pass@xxx.us-east-1.rds.amazonaws.com:5432/dbname`

### Run Migrations

```bash
cd apps/api
go run cmd/migrate/main.go
```

## Available Commands

```bash
make dev-api      # Start backend (port 8080)
make dev-web      # Start frontend (port 5173)
make build        # Build for production
make test         # Run all tests
make db-migrate   # Run database migrations
```

## Documentation

See `docs/` folder for detailed documentation:

- `DATABASE_SETUP.md` - Database setup guide
- `TESTING.md` - Testing guide
- `API_SPECIFICATION.md` - REST API docs

## License

Private — All rights reserved.
