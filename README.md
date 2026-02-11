# NER Studio

AI Image Generation Platform for Creative Teams

## Quick Start

```bash
# Install dependencies
make deps

# Start Supabase local
make db-start

# Run migrations
make db-migrate

# Start development servers
make dev
```

## Project Structure

```
ner-studio/
├── apps/
│   ├── api/          # Go + Fiber backend
│   └── web/          # React + Vite frontend
├── supabase/
│   └── migrations/   # Database migrations
├── docs/             # Documentation
└── Makefile          # Common commands
```

## Backend (API)

The backend is built with Go and Fiber, featuring:

- **Provider System**: Pluggable AI providers (OpenAI, Kie.ai, Gemini)
- **Multi-tier LLM Fallback**: Automatic failover between providers
- **Credit-based Billing**: Per-model pricing with atomic deductions
- **Async Workflows**: Image generation pipeline with callbacks
- **RLS Security**: Row-level security for multi-tenant data

### Key Files

```
apps/api/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/                  # Environment config
│   ├── middleware/              # Auth, CORS, logging
│   ├── handler/                 # HTTP handlers
│   ├── service/                 # Business logic
│   ├── repository/              # Database layer
│   ├── model/                   # Data structs
│   ├── provider/                # Provider implementations
│   ├── workflow/                # Image generation workflow
│   └── external/                # R2 storage client
└── go.mod
```

### Environment Variables

See `apps/api/.env.example` for required environment variables.

## Frontend (Web)

To be implemented with React 19 + Vite + Tailwind 4.

## Database Schema

### Core Tables

- `organizations` — Billing entity with credit balance
- `profiles` — User profiles linked to orgs
- `providers` — AI provider registry
- `generations` — Generation requests
- `generation_images` — Individual generated images
- `credit_ledger` — Immutable credit transaction log
- `invitations` — Pending org invitations

## Available Commands

```bash
make dev          # Start both frontend and backend
make dev-api      # Start backend with hot reload
make dev-web      # Start frontend dev server
make build        # Build for production
make test         # Run all tests
make lint         # Run all linters
make db-migrate   # Run database migrations
make db-reset     # Reset database
```

## Documentation

See `docs/` folder for detailed documentation:

- `PRD.md` — Product Requirements
- `ARCHITECTURE.md` — System Architecture
- `API_SPECIFICATION.md` — REST API Docs
- `DATABASE_SCHEMA.md` — DB Schema Details
- `FRONTEND_SPEC.md` — Frontend Specification

## License

Private — All rights reserved.
