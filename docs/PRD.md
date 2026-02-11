# Product Requirements Document (PRD)

## NER Studio — AI Image Generation Platform

**Version:** 1.0
**Date:** 2026-02-10
**Status:** Draft

---

## 1. Overview

NER Studio is a collaborative AI image generation platform for creative teams. Users upload reference images (for brand/style direction) and product images, write a creative prompt, and the system generates multiple high-quality AI images using those inputs.

The platform is organization-based: billing and credits are managed at the org level, while usage is tracked per individual user.

---

## 2. Problem Statement

Creative teams need to generate on-brand product imagery at scale. The current MVP (built on n8n workflows) validates the core generation loop but lacks:

- User authentication and access control
- Organization-based billing and credit management
- Persistent gallery and prompt history
- Member management for team collaboration
- A production-grade, maintainable codebase

---

## 3. Target Users

| Role | Description |
|------|-------------|
| **Admin** | Manages organization, members, credits, and views all usage across the org |
| **Member** | Uploads images, writes prompts, generates images, views personal gallery |

---

## 4. Core Features

### 4.1 Authentication (P0)

- Google SSO via Supabase Auth
- Session management with JWT tokens
- Users are associated with one organization
- First user of an org becomes Admin by default

### 4.2 Image Generation (P0)

Ported from existing n8n workflow logic. All AI services (vision, LLM, image generation) are **provider-agnostic** — see §4.7 Provider System.

1. **Upload Reference Images** (0–N): Style/brand direction images. Each is analyzed by the active vision provider to extract a text description.
2. **Upload Product Images** (1–N): The product(s) to feature in generated images.
3. **Write Generation Prompt**: Creative direction text from the user.
4. **Select Image Model**: Choose which image generation model to use (e.g., Seedream, Nano Banana Pro). Defaults to the org's default model.
5. **Set Image Count**: Number of prompt variants to generate (default: 3).
6. **Configure Options**: Aspect ratio, quality level (defaults from selected model).
7. **Generate**: System produces N creative prompts via LLM provider chain (with automatic fallback), then sends each prompt + product images to the selected image generation provider.
8. **Async Callback**: Provider calls back with results; images are stored and displayed.

### 4.3 Gallery (P0)

- View all generated images for the current user
- View generation details: prompt used, reference images, product images, timestamp
- Download individual images
- Admin can view gallery across all org members

### 4.4 Organization & Credit Management (P0)

- Each organization has a shared credit pool
- Image generation deducts credits (cost defined **per provider model** — different models have different costs)
- The credit estimate updates dynamically when the user selects a different model
- Admin dashboard shows:
  - Current credit balance
  - Credit usage over time (chart)
  - Usage breakdown by member and by provider/model
- Credit top-up is manual (admin action or future Stripe integration)

### 4.5 Member Management (P1)

- Admin can invite members (via email)
- Admin can remove members
- Admin can change member roles (admin/member)
- Member list with usage stats

### 4.6 Prompt History (P1)

- All prompts are saved to database with metadata
- Users can re-use past prompts
- Search/filter prompt history

### 4.7 Provider System (P0)

All external AI services are managed through a **pluggable provider registry**. This enables:

- **Multiple image generation models**: Admin can configure different models (e.g., Seedream, Nano Banana Pro, FLUX) with different costs per image. Users select which model to use when generating.
- **LLM fallback chain**: Multiple LLM providers are tried in priority order. If one fails, the next is attempted automatically. Admin can reorder, add, or disable providers without code changes.
- **Swappable vision models**: The reference image analysis provider can be changed (e.g., from OpenAI GPT-4o to Gemini Vision) without affecting the rest of the pipeline.
- **Per-provider pricing**: Each image generation model has its own credit cost, shown to users before generation.
- **Admin configuration**: Add new providers, update API keys, change priorities, set costs, enable/disable — all through the admin panel. No redeployment needed.
- **Connection testing**: Admin can test a provider's connectivity before making it active.

---

## 5. User Flows

### 5.1 New User Onboarding

```
Google SSO Login
  → If no org exists: Create org (user becomes admin)
  → If invited to org: Join org as member
  → Redirect to dashboard
```

### 5.2 Image Generation Flow

```
Upload reference images (optional)
  → Upload product images (required)
  → Write prompt
  → Select image model (dropdown, shows cost per image)
  → Configure aspect ratio + quality (optional, defaults from model)
  → Set image count
  → Review credit estimate (count × model cost)
  → Click "Generate"
  → System: analyze references (vision provider)
         → generate prompt variants (LLM chain with fallback)
         → call image generation provider
  → Async: receive callbacks → store images → update gallery
  → User sees generated images in real-time as they complete
```

### 5.3 Admin: Credit Management

```
View credit balance
  → View usage breakdown (by member, by date)
  → Add credits (manual top-up)
```

---

## 6. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Image generation latency | < 120s per image (provider dependent) |
| Concurrent users per org | Up to 50 |
| Image storage | Cloudflare R2 via S3-compatible API |
| Auth provider | Supabase Auth (Google OAuth) |
| Provider switch latency | < 1s (in-memory registry, DB-backed) |
| LLM fallback total time | < 30s across full chain |
| Uptime | 99.5% |

---

## 7. Out of Scope (v1)

- Stripe/payment integration (credits added manually)
- Multi-org membership (1 user = 1 org)
- Image editing/post-processing
- Public sharing/link generation
- Mobile-native apps

---

## 8. Success Metrics

| Metric | Target |
|--------|--------|
| Generation success rate | > 95% |
| Average generation time | < 90s |
| User retention (weekly active) | > 60% |
| Credits consumed per user/week | Tracked for growth |

---

## 9. Dependencies

Core infrastructure (always required):

| Service | Purpose |
|---------|---------|
| **Supabase** | Auth (Google SSO) + PostgreSQL database |
| **Cloudflare R2** | Image storage (S3-compatible) |

AI providers (pluggable via provider registry):

| Service | Category | Default Role |
|---------|----------|-------------|
| **kie.ai API** | Image Gen + LLM | Primary image generation (Seedream, Nano Banana Pro) + LLM proxy (Gemini 3 Pro, Gemini 2.5 Pro) |
| **OpenAI API** | Vision | Reference image analysis (GPT-4o) |
| **Google Gemini API** | LLM | Backup LLM for prompt generation (direct API) |

All AI providers can be added, removed, or replaced via the admin panel without code changes.
