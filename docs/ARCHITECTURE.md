# Technical Architecture Document

## NER Studio

**Version:** 1.0
**Date:** 2026-02-10

---

## 1. System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND                                  │
│                   React + Vite + TailwindCSS                     │
│                     (Bauhaus Design)                             │
│                                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────────────┐  │
│  │  Auth     │ │ Generate │ │ Gallery  │ │  Admin Panel      │  │
│  │  (SSO)   │ │  Page    │ │  Page    │ │  (Credits/Members)│  │
│  └──────────┘ └──────────┘ └──────────┘ └───────────────────┘  │
└─────────────────────────┬───────────────────────────────────────┘
                          │ HTTPS (REST API)
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND                                   │
│                   Go + Fiber Framework                            │
│                                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │  Auth    │ │ Generate │ │ Gallery  │ │  Admin            │  │
│  │ Middleware│ │ Handler  │ │ Handler  │ │  Handler          │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────────────┘  │
│       │             │            │              │                 │
│  ┌────┴─────────────┴────────────┴──────────────┴──────────┐    │
│  │                   Service Layer                          │    │
│  │  AuthService │ GenerationService │ GalleryService │ etc  │    │
│  └──────────────────────┬───────────────────────────────────┘    │
└─────────────────────────┼───────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────────┐
          ▼               ▼                   ▼
┌──────────────┐ ┌──────────────┐   ┌──────────────────┐
│  PostgreSQL  │ │ Cloudflare   │   │  External APIs   │
│  (Database)  │ │  R2 Storage  │   │                  │
│              │ │              │   │  - kie.ai        │
│  - Users     │ │  - Generated │   │  - OpenAI        │
│  - Orgs      │ │    images    │   │  - Gemini        │
│  - Generations│ │  - Uploads   │   │                  │
└──────────────┘ └──────────────┘   └──────────────────┘
```

---

## 2. Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Frontend | React 19 + Vite | Fast dev experience, modern React features |
| Styling | TailwindCSS | Utility-first, easy Bauhaus implementation |
| State | Zustand | Lightweight, minimal boilerplate |
| Backend | Go 1.23 + Fiber v2 | High performance, familiar Express-like API |
| Auth | Simple JWT | Built-in, no external dependency |
| Database | PostgreSQL | Any provider (local, Railway, Neon, AWS) |
| Storage | Cloudflare R2 | S3-compatible, no egress fees |
| Image Gen | **Provider Registry** | Pluggable — see §9 Provider System |
| Image Analysis | **Provider Registry** | Pluggable — see §9 Provider System |
| LLM Prompts | **Provider Registry** | Pluggable — see §9 Provider System |

---

## 3. Authentication Flow

```
Frontend                    Backend (Fiber)              PostgreSQL
   │                            │                           │
   │  1. POST /auth/login       │                           │
   │  {email, password}         │                           │
   │───────────────────────────>│                           │
   │                            │  2. Query user by email   │
   │                            │──────────────────────────>│
   │                            │  3. Return user + hash    │
   │                            │<──────────────────────────│
   │                            │                           │
   │                            │  4. Verify password       │
   │                            │  5. Generate JWT          │
   │                            │                           │
   │  6. Return JWT token       │                           │
   │<───────────────────────────│                           │
   │                            │                           │
   │  7. API request with       │                           │
   │     Bearer token           │                           │
   │───────────────────────────>│                           │
   │                            │  8. Verify JWT locally    │
   │                            │  (no external call)       │
   │                            │                           │
   │  9. Response               │                           │
   │<───────────────────────────│                           │
```

- Simple JWT authentication (no external service)
- Backend validates JWT locally via middleware
- User profile + org association stored in PostgreSQL
- Passwords hashed with bcrypt

---

## 4. Image Generation Pipeline

Ported from n8n workflow_1 logic into Go services. All external API calls go through the **Provider Registry** (§9), making every model and endpoint swappable.

Each pipeline is implemented as a **workflow module** — a single Go file in `internal/workflow/`. The v1 workflow is `imagegen.go`. Future workflows (e.g., video generation) would be separate files reusing the same shared providers and services.

```
Workflow Module: internal/workflow/imagegen.go
  Contains: hardcoded prompts, orchestration logic, prompt splitting/parsing
  Uses: Provider Registry (shared), Credit Service (shared), R2 Client (shared)

Step 1: Receive Request
  POST /api/v1/generations
  Body: { referenceImages[], productImages[], prompt, imageCount,
          aspectRatio?, quality?, imageModel?, llmModel? }

Step 2: Upload Images to R2 (if raw files)
  → Store in R2, get public URLs

Step 3: Analyze Reference Images (parallel)
  → Resolve active "vision" provider from registry
  → For each reference image: call vision provider API
  → Get text descriptions

Step 4: Generate Creative Prompts
  → Build system prompt with reference descriptions (hardcoded in imagegen.go)
  → Resolve active "llm" provider chain from registry (ordered by priority)
  → Walk the chain: try provider 1 → on failure try provider 2 → etc.
  → Parse response into N individual prompts

Step 5: Submit Image Generation Jobs (parallel)
  → Resolve active "image_generation" provider from registry
  → For each prompt: call image generation provider API
  → Include product image URLs, aspect_ratio, quality
  → Set callback URL: /api/v1/callbacks/{provider_slug}

Step 6: Return Task IDs
  → Respond with generation ID + task IDs
  → Frontend polls or listens for updates

Step 7: Receive Callbacks
  POST /api/v1/callbacks/:provider (from external provider)
  → Route to correct provider's callback parser
  → Parse result URLs
  → Download and re-upload to R2 (for permanence)
  → Update generation record in DB
  → Deduct credits from org
  → Notify frontend via SSE or polling
```

---

## 5. Callback Handling Strategy

Image generation providers use async callbacks. Our approach:

1. **Per-provider callback endpoints**: `POST /api/v1/callbacks/:provider_slug` — each provider has its own callback parser since response formats differ
2. **Store results** in DB with status updates
3. **Frontend polling**: Client polls `GET /api/v1/generations/:id` every 5s
4. **Future enhancement**: Replace polling with Server-Sent Events (SSE)

---

## 6. Credit System

```
Credit Flow:
  Admin adds credits to org → org.credits += amount
  User generates images → org.credits -= (imageCount * provider.cost_per_image)
  Each generation logs: user_id, org_id, credits_used, provider_used, timestamp

Cost Model (per-provider, configurable via admin panel):
  Each image generation provider has its own cost_per_image
  Example: kie.ai seedream = 7 credits, kie.ai nano-banana-pro = 10 credits
  Reference analysis = free (absorbed cost)
  LLM prompt generation = free (absorbed cost)
```

---

## 7. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Monorepo | Single repo with `/apps/web`, `/apps/api` | Shared types, atomic commits, simpler CI |
| Auth | Simple JWT (built-in) | No external dependency, full control |
| DB | Standard PostgreSQL | Works with any provider, simple connection |
| Storage | R2 (not Supabase Storage) | Already in use, no egress fees |
| Polling (not WebSocket) | Simpler for v1 | Fewer failure modes, easy to implement |
| Go Fiber (not Echo/Gin) | Express-like API | Familiar patterns, fast performance |
| Provider Registry | DB-driven config (not env vars) | Admins can swap models without redeploy, add new providers at runtime |
| Provider Interface | Go interfaces per category | New providers implement an interface, no core code changes needed |
| Workflow Modules | One `.go` file per workflow | Self-contained pipeline logic, reuses shared providers/services |

---

## 8. Environment Configuration

```
# Backend (.env)
DATABASE_URL=postgresql://user:pass@host:5432/dbname
JWT_SECRET=your-secret-key-min-32-characters

# API keys are stored in the providers table (encrypted),
# but can be overridden via env vars for local development:
KIE_AI_API_KEY=...
OPENAI_API_KEY=...
GEMINI_API_KEY=...

R2_ACCOUNT_ID=...
R2_ACCESS_KEY_ID=...
R2_SECRET_ACCESS_KEY=...
R2_BUCKET_NAME=ner-storage
R2_PUBLIC_URL=https://bucket.tansil.pro

CALLBACK_BASE_URL=https://api.nerstudio.com
PROVIDER_KEY_ENCRYPTION_SECRET=...

# Frontend (.env)
VITE_API_URL=https://api.nerstudio.com
```

---

## 9. Provider System

The provider system is the core abstraction that makes all external AI services pluggable. Providers are configured in the database and managed via the admin panel — no redeployment needed to switch models.

### 9.1 Provider Categories

```
┌───────────────────────────────────────────────────────────┐
│                    PROVIDER REGISTRY                       │
│                    (providers table)                        │
│                                                            │
│  ┌─────────────────┐  ┌────────────┐  ┌────────────────┐ │
│  │ IMAGE GENERATION │  │    LLM     │  │    VISION      │ │
│  │                  │  │            │  │                │ │
│  │ kie.ai seedream  │  │ kie.ai     │  │ OpenAI GPT-4o  │ │
│  │ kie.ai nano-     │  │  Gemini 3  │  │ Gemini Vision  │ │
│  │   banana-pro     │  │ kie.ai     │  │ Claude Vision  │ │
│  │ Replicate FLUX   │  │  Gemini2.5 │  │ (future)       │ │
│  │ (future)         │  │ Google     │  │                │ │
│  │                  │  │  Gemini    │  │                │ │
│  │                  │  │  Direct    │  │                │ │
│  │                  │  │ OpenAI     │  │                │ │
│  │                  │  │  GPT-4o    │  │                │ │
│  │                  │  │ (future)   │  │                │ │
│  └─────────────────┘  └────────────┘  └────────────────┘ │
└───────────────────────────────────────────────────────────┘
```

| Category | Purpose | Interface Contract |
|----------|---------|-------------------|
| `image_generation` | Generate images from prompt + product images | `Submit(prompt, imageURLs, config) → taskID` + callback parser |
| `llm` | Generate creative prompt variants from system prompt + user input | `ChatCompletion(messages, config) → text` |
| `vision` | Analyze reference images to extract text descriptions | `AnalyzeImage(imageURL, systemPrompt) → text` |

### 9.2 Go Interface Design

```go
// internal/provider/types.go

// ImageGenerationProvider generates images from prompts
type ImageGenerationProvider interface {
    Name() string
    Slug() string

    // Submit sends a generation job, returns external task ID
    Submit(ctx context.Context, req ImageGenRequest) (taskID string, err error)

    // ParseCallback parses the provider-specific callback payload
    ParseCallback(body []byte) (*ImageGenResult, error)

    // CallbackPath returns the URL path suffix for this provider's callbacks
    // e.g. "kie" → /api/v1/callbacks/kie
    CallbackPath() string
}

type ImageGenRequest struct {
    Prompt      string
    ImageURLs   []string  // product image URLs
    AspectRatio string    // e.g. "9:16", "1:1", "16:9"
    Quality     string    // e.g. "high", "standard"
    Model       string    // e.g. "seedream/4.5-edit", "nano-banana-pro"
    CallbackURL string
}

type ImageGenResult struct {
    TaskID      string
    Success     bool
    ImageURLs   []string  // result image URLs (may be temporary)
    CostTime    int       // seconds
    Error       string
}

// LLMProvider generates text completions
type LLMProvider interface {
    Name() string
    Slug() string

    // ChatCompletion sends messages and returns the response text
    ChatCompletion(ctx context.Context, req LLMRequest) (string, error)
}

type LLMRequest struct {
    SystemPrompt string
    UserContent  []ContentPart   // text + image_url parts
    Model        string          // optional override
    MaxTokens    int
}

type ContentPart struct {
    Type     string  // "text" or "image_url"
    Text     string
    ImageURL string
}

// VisionProvider analyzes images
type VisionProvider interface {
    Name() string
    Slug() string

    // AnalyzeImage returns a text description of the image
    AnalyzeImage(ctx context.Context, imageURL string, prompt string) (string, error)
}
```

### 9.3 Provider Registry Service

```go
// internal/provider/registry.go

type Registry struct {
    db          *sql.DB
    imgProviders map[string]ImageGenerationProvider  // slug → provider
    llmProviders map[string]LLMProvider
    visProviders map[string]VisionProvider
}

// GetImageGenProvider returns the active image generation provider
// (or a specific one by slug)
func (r *Registry) GetImageGenProvider(slug string) (ImageGenerationProvider, error)

// GetLLMChain returns LLM providers ordered by priority (for fallback)
func (r *Registry) GetLLMChain() ([]LLMProvider, error)

// GetVisionProvider returns the active vision provider
func (r *Registry) GetVisionProvider() (VisionProvider, error)

// Reload re-reads provider configs from DB (call after admin changes)
func (r *Registry) Reload(ctx context.Context) error
```

### 9.4 Concrete Provider Implementations (v1)

**Image Generation:**

| Slug | Provider | Models | Notes |
|------|----------|--------|-------|
| `kieai` | kie.ai | `seedream/4.5-edit`, `nano-banana-pro` | Async callback-based. Same API, different `model` field |

**LLM (Prompt Generation):**

| Slug | Provider | Models | Request Format | Notes |
|------|----------|--------|----------------|-------|
| `kieai-gemini3` | kie.ai proxy | Gemini 3 Pro | OpenAI-compatible | `image_url` in user content |
| `kieai-gemini25` | kie.ai proxy | Gemini 2.5 Pro | OpenAI-compatible | `image_url` in user content |
| `google-gemini` | Google direct | Gemini 3 Pro Preview | Gemini native | Requires base64 image conversion |

**Vision (Reference Analysis):**

| Slug | Provider | Model | Notes |
|------|----------|-------|-------|
| `openai-vision` | OpenAI | GPT-4o | Default. Sends `image_url` directly |

### 9.5 How Provider Config Flows

```
Admin Panel                    Backend                    Database
     │                            │                          │
     │  1. Admin configures        │                          │
     │  provider in Settings UI    │                          │
     │  (model, API key, priority, │                          │
     │   cost_per_image, etc.)     │                          │
     │────────────────────────────>│                          │
     │                             │  2. Write to providers   │
     │                             │  table (encrypted keys)  │
     │                             │─────────────────────────>│
     │                             │                          │
     │                             │  3. Reload registry      │
     │                             │  (in-memory cache)       │
     │                             │                          │
     │                             │                          │
     │  User clicks "Generate"     │                          │
     │────────────────────────────>│                          │
     │                             │  4. Registry resolves    │
     │                             │  active providers:       │
     │                             │  - vision: openai-vision │
     │                             │  - llm: [kieai-gemini3,  │
     │                             │    kieai-gemini25,        │
     │                             │    google-gemini]         │
     │                             │  - image: kieai           │
     │                             │    (nano-banana-pro)      │
     │                             │                          │
     │                             │  5. Execute pipeline     │
     │                             │  using resolved providers│
```

### 9.6 Adding a New Provider (Developer Guide)

To add a new provider (e.g., Replicate for image generation):

1. **Create implementation**: `internal/provider/replicate.go`
   - Implement the `ImageGenerationProvider` interface
   - Handle the provider's specific API format, auth, and callback parsing

2. **Register in factory**: `internal/provider/factory.go`
   - Add `"replicate"` slug to the provider factory
   - Map it to the new implementation

3. **No core pipeline changes needed** — the generation service calls the interface, not the concrete provider

4. **Admin configures via UI**: Add the provider record to `providers` table with API key, model, cost, etc.

```
That's it. No env var changes, no config file edits, no redeployment needed
(unless adding the Go code for a brand-new provider type).
```

### 9.7 Adding a New Workflow (Developer Guide)

To add a new workflow (e.g., video generation):

1. **Create workflow file**: `internal/workflow/videogen.go`
   - Define hardcoded prompts specific to this workflow
   - Implement the orchestration logic (which providers to call, in what order)
   - Reuse shared providers via `Registry.GetLLMChain()`, `Registry.GetVisionProvider()`, etc.

2. **Create handler**: `internal/handler/video.go`
   - Wire the HTTP endpoint (e.g., `POST /api/v1/videos`)
   - Call the workflow module

3. **Reuse everything else**: providers, credits, R2 upload, auth middleware — all shared

```
Each workflow is a single file. It owns its prompts and pipeline.
It borrows providers, credits, and storage from shared infrastructure.
```
