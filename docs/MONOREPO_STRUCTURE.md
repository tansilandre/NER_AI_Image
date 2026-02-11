# Monorepo Structure Document

## NER Studio

**Version:** 1.0
**Date:** 2026-02-10

---

## 1. Directory Structure

```
ner-studio/
├── .github/
│   └── workflows/
│       ├── ci.yml                  # Lint + test on PR
│       ├── deploy-api.yml          # Deploy Go backend
│       └── deploy-web.yml          # Deploy React frontend
│
├── apps/
│   ├── web/                        # React + Vite frontend
│   │   ├── public/
│   │   ├── src/
│   │   │   ├── assets/             # Static assets (logos, icons)
│   │   │   ├── components/         # Shared UI components
│   │   │   │   ├── ui/             # Base UI primitives (Button, Input, Card, etc.)
│   │   │   │   ├── layout/         # Layout components (Sidebar, Header, etc.)
│   │   │   │   └── shared/         # Shared business components
│   │   │   ├── features/           # Feature modules
│   │   │   │   ├── auth/           # Login, onboarding
│   │   │   │   ├── generate/       # Image generation page
│   │   │   │   ├── gallery/        # Gallery page
│   │   │   │   └── admin/          # Admin panel (credits, members)
│   │   │   ├── hooks/              # Custom React hooks
│   │   │   ├── lib/                # Utilities, API client, Supabase client
│   │   │   ├── stores/             # Zustand stores
│   │   │   ├── types/              # TypeScript types
│   │   │   ├── App.tsx
│   │   │   ├── main.tsx
│   │   │   └── router.tsx          # React Router config
│   │   ├── index.html
│   │   ├── tailwind.config.ts
│   │   ├── vite.config.ts
│   │   ├── tsconfig.json
│   │   └── package.json
│   │
│   └── api/                        # Go + Fiber backend
│       ├── cmd/
│       │   └── server/
│       │       └── main.go         # Entry point
│       ├── internal/
│       │   ├── config/             # Environment config loading
│       │   │   └── config.go
│       │   ├── middleware/          # Fiber middleware
│       │   │   ├── auth.go         # JWT verification via Supabase
│       │   │   ├── cors.go
│       │   │   └── logger.go
│       │   ├── handler/            # HTTP handlers (controllers)
│       │   │   ├── auth.go
│       │   │   ├── generation.go
│       │   │   ├── gallery.go
│       │   │   ├── upload.go
│       │   │   ├── callback.go     # Generic callback router → provider.ParseCallback()
│       │   │   ├── admin.go
│       │   │   └── admin_provider.go # Provider CRUD + test endpoint
│       │   ├── workflow/            # Workflow modules (one file per workflow)
│       │   │   └── imagegen.go    # Image generation: prompts, orchestration, parsing
│       │   ├── service/            # Shared business logic
│       │   │   ├── auth.go
│       │   │   ├── gallery.go
│       │   │   ├── upload.go
│       │   │   ├── credit.go
│       │   │   └── member.go
│       │   ├── repository/         # Database queries
│       │   │   ├── generation.go
│       │   │   ├── gallery.go
│       │   │   ├── profile.go
│       │   │   ├── organization.go
│       │   │   ├── credit.go
│       │   │   └── provider.go     # Provider CRUD + query by category/priority
│       │   ├── model/              # Data models / structs
│       │   │   ├── generation.go
│       │   │   ├── profile.go
│       │   │   ├── organization.go
│       │   │   └── credit.go
│       │   ├── provider/            # Provider registry + implementations
│       │   │   ├── types.go        # Interfaces: ImageGenerationProvider, LLMProvider, VisionProvider
│       │   │   ├── registry.go     # Registry service (load from DB, cache, resolve)
│       │   │   ├── factory.go      # Slug → concrete provider factory
│       │   │   ├── kieai_image.go  # kie.ai image gen (seedream, nano-banana-pro)
│       │   │   ├── kieai_llm.go    # kie.ai LLM proxy (OpenAI-compatible)
│       │   │   ├── google_llm.go   # Google Gemini direct (native format)
│       │   │   └── openai_vision.go # OpenAI GPT-4o vision
│       │   ├── external/           # Non-provider external clients
│       │   │   └── r2.go           # Cloudflare R2 (S3) client
│       │   └── router/
│       │       └── router.go       # Route definitions
│       ├── go.mod
│       ├── go.sum
│       └── Dockerfile
│
├── packages/                       # Shared code (future)
│   └── shared-types/               # Shared TypeScript types (if needed)
│
├── docs/                           # Project documentation
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   ├── DATABASE_SCHEMA.md
│   ├── API_SPECIFICATION.md
│   ├── FRONTEND_SPEC.md
│   ├── MONOREPO_STRUCTURE.md
│   └── workflow_1.json             # Original n8n workflow (reference)
│
├── supabase/                       # Supabase local dev config
│   ├── migrations/                 # SQL migration files
│   │   ├── 001_create_organizations.sql
│   │   ├── 002_create_profiles.sql
│   │   ├── 003_create_generations.sql
│   │   ├── 004_create_generation_images.sql
│   │   ├── 005_create_credit_ledger.sql
│   │   ├── 006_create_invitations.sql
│   │   ├── 007_create_providers.sql
│   │   ├── 008_create_functions.sql
│   │   ├── 009_create_rls_policies.sql
│   │   └── 010_seed_providers.sql     # Default provider config
│   ├── seed.sql                    # Dev seed data
│   └── config.toml                 # Supabase local config
│
├── .env.example                    # Template for env vars
├── .gitignore
├── Makefile                        # Common commands
└── CLAUDE.md                       # AI assistant instructions
```

---

## 2. Makefile Commands

```makefile
# Development
make dev-web          # Start React dev server (port 5173)
make dev-api          # Start Go server with hot reload (port 8080)
make dev              # Start both concurrently

# Database
make db-migrate       # Run Supabase migrations
make db-reset         # Reset database and re-run migrations
make db-seed          # Seed development data

# Build
make build-web        # Build React for production
make build-api        # Build Go binary

# Test
make test-api         # Run Go tests
make test-web         # Run React tests
make test             # Run all tests

# Lint
make lint-api         # golangci-lint
make lint-web         # eslint
make lint             # Run all linters
```

---

## 3. Key Tooling

| Tool | Purpose |
|------|---------|
| **pnpm** | Node package manager (for web app) |
| **Air** | Go hot-reload for development |
| **Supabase CLI** | Local Supabase + migrations |
| **golangci-lint** | Go linting |
| **ESLint + Prettier** | Frontend linting/formatting |

---

## 4. Git Conventions

| Pattern | Example |
|---------|---------|
| Branch naming | `feat/auth-google-sso`, `fix/credit-deduction` |
| Commit prefix | `feat:`, `fix:`, `docs:`, `refactor:`, `chore:` |
| Scope | `feat(api): add generation endpoint` |
| PR target | Always to `main` |

---

## 5. Deployment Strategy

| Component | Platform | Trigger |
|-----------|----------|---------|
| Frontend (`apps/web`) | Cloudflare Pages or Vercel | Push to `main` |
| Backend (`apps/api`) | Railway or Fly.io | Push to `main` |
| Database | Supabase (managed) | Migration via CI |
| Storage | Cloudflare R2 | Already configured |
