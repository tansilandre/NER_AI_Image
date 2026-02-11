# NER Studio — AI Image Generation Platform

**Last Updated:** 2026-02-11

---

## 1. Project Overview

NER Studio is a collaborative AI image generation platform for creative teams. Users upload reference images (for brand/style direction) and product images, write creative prompts, and the system generates multiple high-quality AI images.

### Key Features
- **Organization-based**: Billing and credits managed at org level, usage tracked per user
- **Provider-agnostic AI services**: Swappable vision, LLM, and image generation providers
- **Multi-tier LLM fallback**: Automatic failover between LLM providers if one fails
- **Credit-based billing**: Per-model pricing, real-time credit estimation

### Roles
- **Admin**: Manages organization, members, credits, views all usage
- **Member**: Uploads images, writes prompts, generates images, views personal gallery

---

## 2. Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | React 19 + Vite | Web UI with Bauhaus design system |
| **Styling** | TailwindCSS 4 | Utility-first CSS |
| **State** | Zustand | Frontend state management |
| **Backend** | Go 1.23 + Fiber v2 | High-performance REST API |
| **Auth** | Supabase Auth | Google SSO, JWT tokens |
| **Database** | Supabase PostgreSQL | Managed Postgres with RLS |
| **Storage** | Cloudflare R2 | S3-compatible image storage |
| **AI Providers** | Provider Registry | Pluggable kie.ai, OpenAI, Gemini |

### External Services
| Service | Purpose |
|---------|---------|
| **kie.ai API** | Image generation (Seedream, Nano Banana Pro) + LLM proxy (Gemini) |
| **OpenAI API** | Vision analysis (GPT-4o) |
| **Google Gemini API** | Backup LLM (direct API) |

---

## 3. Project Structure

```
ner-studio/
├── apps/
│   ├── web/                        # React + Vite frontend
│   │   ├── src/
│   │   │   ├── components/         # UI components (Bauhaus style)
│   │   │   ├── features/           # auth, generate, gallery, admin
│   │   │   ├── stores/             # Zustand stores
│   │   │   └── types/              # TypeScript types
│   │   ├── package.json
│   │   └── vite.config.ts
│   │
│   └── api/                        # Go + Fiber backend
│       ├── cmd/server/main.go      # Entry point
│       ├── internal/
│       │   ├── config/             # Environment config
│       │   ├── middleware/         # Auth, CORS, logging
│       │   ├── handler/            # HTTP handlers
│       │   ├── workflow/           # Workflow modules (imagegen.go)
│       │   ├── service/            # Business logic
│       │   ├── repository/         # Database queries
│       │   ├── model/              # Data structs
│       │   ├── provider/           # Provider registry + implementations
│       │   └── external/r2.go      # R2 storage client
│       ├── go.mod
│       └── Dockerfile
│
├── supabase/
│   └── migrations/                 # SQL migration files
│       ├── 001_create_organizations.sql
│       ├── 002_create_profiles.sql
│       ├── 003_create_generations.sql
│       ├── 004_create_generation_images.sql
│       ├── 005_create_credit_ledger.sql
│       ├── 006_create_invitations.sql
│       ├── 007_create_providers.sql
│       ├── 008_create_functions.sql
│       ├── 009_create_rls_policies.sql
│       └── 010_seed_providers.sql
│
├── docs/                           # Project documentation
│   ├── PRD.md                      # Product Requirements
│   ├── ARCHITECTURE.md             # System architecture
│   ├── API_SPECIFICATION.md        # REST API docs
│   ├── FRONTEND_SPEC.md            # Frontend spec
│   ├── DATABASE_SCHEMA.md          # DB schema
│   ├── MONOREPO_STRUCTURE.md       # Repo structure
│   └── FLOW.md                     # Complex flows
│
├── workflow_1.json                 # Original n8n workflow (image generation)
├── workflow_2.json                 # Original n8n workflow (upload/status)
├── Makefile                        # Common commands
└── AGENTS.md                       # This file
```

---

## 4. Build and Development Commands

### Prerequisites
- Node.js 20+ with pnpm
- Go 1.23+
- Supabase CLI
- Air (Go hot-reload): `go install github.com/air-verse/air@latest`

### Development
```bash
# Start frontend dev server (port 5173)
make dev-web

# Start backend with hot reload (port 8080)
make dev-api

# Start both concurrently
make dev
```

### Database
```bash
# Run Supabase migrations
make db-migrate

# Reset database and re-run migrations
make db-reset

# Seed development data
make db-seed
```

### Build
```bash
# Build React for production
make build-web

# Build Go binary
make build-api
```

### Test
```bash
# Run Go tests
make test-api

# Run React tests
make test-web

# Run all tests
make test
```

### Lint
```bash
# Lint Go code (golangci-lint)
make lint-api

# Lint frontend (ESLint)
make lint-web

# Run all linters
make lint
```

---

## 5. Environment Configuration

### Backend (.env)
```bash
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_SERVICE_ROLE_KEY=eyJ...
DATABASE_URL=postgresql://...

# Provider API keys (can also be in DB)
KIE_AI_API_KEY=...
OPENAI_API_KEY=...
GEMINI_API_KEY=...

# R2 Storage
R2_ACCOUNT_ID=...
R2_ACCESS_KEY_ID=...
R2_SECRET_ACCESS_KEY=...
R2_BUCKET_NAME=ner-storage
R2_PUBLIC_URL=https://bucket.tansil.pro

CALLBACK_BASE_URL=https://api.nerstudio.com
PROVIDER_KEY_ENCRYPTION_SECRET=...  # AES key for encrypting API keys in DB
```

### Frontend (.env)
```bash
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=eyJ...
VITE_API_URL=https://api.nerstudio.com
```

---

## 6. Provider System

The provider system makes all external AI services pluggable. Providers are configured in the database and managed via the admin panel.

### Provider Categories

| Category | Purpose | Interface |
|----------|---------|-----------|
| `image_generation` | Generate images from prompts | `ImageGenerationProvider` |
| `llm` | Generate creative prompt variants | `LLMProvider` |
| `vision` | Analyze reference images | `VisionProvider` |

### Adding a New Provider

1. **Create implementation**: `internal/provider/{name}.go`
   - Implement the appropriate interface
   - Handle API-specific format, auth, and callbacks

2. **Register in factory**: `internal/provider/factory.go`
   - Add slug mapping to the new implementation

3. **Admin configures via UI**: Add record to `providers` table with API key, model, cost

No env var changes or redeployment needed for configuration changes.

---

## 7. Key Workflows

### Image Generation Pipeline
```
POST /api/v1/generations
  → Validate & check credits
  → Create DB records
  → Analyze reference images (parallel, vision provider)
  → Build LLM messages
  → Generate prompts (LLM chain with fallback)
  → Split prompts
  → Submit image gen jobs (parallel, image provider)
  → Return task IDs
  → [Async] Receive callbacks, store images, deduct credits
```

### LLM Fallback Chain
The system tries LLM providers in priority order:
1. `kieai-gemini3` (priority 0)
2. `kieai-gemini25` (priority 1)
3. `google-gemini` (priority 2, direct API)

If a provider returns `error_code_for_fallback`, the next provider is attempted.

### Callback Handling
- Each provider has its own callback endpoint: `POST /api/v1/callbacks/:provider_slug`
- Provider's `ParseCallback()` handles format-specific parsing
- Images are downloaded from temp URLs and re-uploaded to R2
- Credits are deducted atomically after successful generation

---

## 8. Database Schema

### Core Tables
- `organizations` — Billing entity with credit balance
- `profiles` — User profiles linked to orgs (roles: admin/member)
- `providers` — AI provider registry (global config)
- `generations` — Generation requests with status
- `generation_images` — Individual generated images
- `credit_ledger` — Immutable credit transaction log
- `invitations` — Pending org invitations

### RLS Policies
All tables have Row Level Security enabled. Users can only access data from their own organization.

### Key Functions
- `deduct_credits()` — Atomically deduct credits with row locking
- `handle_new_user()` — Auto-create profile on auth signup

---

## 9. API Endpoints

### Auth
- `POST /api/v1/auth/callback` — Post-OAuth callback
- `POST /api/v1/auth/onboarding` — Create/join org

### Generations
- `POST /api/v1/generations` — Start generation (202 Accepted)
- `GET /api/v1/generations` — List generations
- `GET /api/v1/generations/:id` — Get generation status (polling)

### Gallery
- `GET /api/v1/gallery` — List completed images

### Upload
- `POST /api/v1/uploads` — Upload image to R2

### Callbacks (No auth)
- `POST /api/v1/callbacks/:provider_slug` — Provider callbacks

### Admin (Admin only)
- `GET /api/v1/admin/organization` — Org details
- `POST /api/v1/admin/credits` — Add credits
- `GET /api/v1/admin/credits/history` — Credit transactions
- `GET /api/v1/admin/members` — List members
- `POST /api/v1/admin/members/invite` — Invite member
- `GET/POST/PATCH/DELETE /api/v1/admin/providers` — Provider CRUD
- `POST /api/v1/admin/providers/:slug/test` — Test provider connection

### Providers (All authenticated users)
- `GET /api/v1/providers?category=...` — List active providers

---

## 10. Frontend Pages

| Route | Purpose |
|-------|---------|
| `/login` | Google SSO login |
| `/onboarding` | Create/join org (first-time) |
| `/generate` | Image generation workspace (main) |
| `/gallery` | User's generated images |
| `/gallery/:id` | Generation detail view |
| `/admin/credits` | Credit management + stats |
| `/admin/members` | Member management |
| `/admin/providers` | AI provider configuration |
| `/admin/gallery` | Org-wide gallery |

### State Management (Zustand)
- `auth.ts` — User, org, login/logout
- `generation.ts` — Images, prompt, provider selection, generation flow
- `gallery.ts` — Images, filters, pagination

---

## 11. Code Style Guidelines

### Go
- Use standard Go formatting (`gofmt`)
- Follow `golangci-lint` rules
- Group imports: stdlib, external, internal
- Interface definitions in `internal/provider/types.go`
- One workflow per file in `internal/workflow/`

### TypeScript/React
- ESLint + Prettier configuration
- Functional components with hooks
- Zustand for state management
- Tailwind for styling (Bauhaus design system)

### Git Conventions
| Pattern | Example |
|---------|---------|
| Branch naming | `feat/auth-google-sso`, `fix/credit-deduction` |
| Commit prefix | `feat:`, `fix:`, `docs:`, `refactor:`, `chore:` |
| Scope | `feat(api): add generation endpoint` |
| PR target | Always to `main` |

---

## 12. Testing Strategy

- **Unit tests**: Service layer logic
- **Integration tests**: API endpoints with test database
- **Provider tests**: Mock external APIs
- Run with `make test-api` and `make test-web`

---

## 13. Deployment

| Component | Platform | Trigger |
|-----------|----------|---------|
| Frontend (`apps/web`) | Cloudflare Pages or Vercel | Push to `main` |
| Backend (`apps/api`) | Railway or Fly.io | Push to `main` |
| Database | Supabase (managed) | Migration via CI |
| Storage | Cloudflare R2 | Already configured |

---

## 14. Legacy Reference

The `workflow_1.json` and `workflow_2.json` files contain the original n8n workflows that this platform replaces. They serve as reference for:
- Image generation pipeline logic
- Prompt templates and system prompts
- LLM request/response formats
- Callback parsing logic
- Error handling patterns

Key learnings from n8n workflows:
- **3-tier LLM fallback** is critical for reliability
- **Reference image analysis** uses GPT-4o with a specific system prompt
- **Prompt splitting** has multiple fallback strategies (double newline, semicolon)
- **Async callbacks** require task ID tracking and R2 re-upload

---

## 15. Security Considerations

1. **API keys** are encrypted in the database using `PROVIDER_KEY_ENCRYPTION_SECRET`
2. **JWT tokens** from Supabase are verified on every request
3. **RLS policies** ensure data isolation between organizations
4. **Credit deduction** uses atomic DB operations with row locking
5. **Callbacks** are verified by task_id lookup (no auth header from providers)
6. **CORS** is configured for specific origins only

---

## 16. Common Tasks

### Add a new LLM provider to the fallback chain
1. Admin creates provider via API with `category: "llm"`
2. Set `priority` to control position in chain
3. Configure `config.error_code_for_fallback` for fallback triggers
4. No code changes needed

### Add a new image generation model
1. Create provider record with `category: "image_generation"`
2. Set `config.model` to the provider's model identifier
3. Configure `cost_per_use` for credit pricing
4. If using same API (e.g., kie.ai), can reuse existing callback parser

### Add a new workflow (e.g., video generation)
1. Create `internal/workflow/videogen.go`
2. Define prompts and orchestration logic
3. Create handler in `internal/handler/video.go`
4. Reuse shared providers, credits, R2 client

---

## 17. Resources

- **Supabase Docs**: https://supabase.com/docs
- **Fiber Docs**: https://docs.gofiber.io
- **Tailwind Docs**: https://tailwindcss.com/docs
- **kie.ai API Docs**: Refer to provider documentation
- **Design Reference**: Bauhaus principles — geometric forms, primary colors, grid-based layouts
