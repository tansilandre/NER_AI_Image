# API Specification Document

## NER Studio — Go Fiber REST API

**Version:** 1.0
**Date:** 2026-02-10
**Base URL:** `/api/v1`

---

## 1. Authentication

All endpoints (except `/auth/*` and `/callbacks/*`) require a valid Supabase JWT in the `Authorization` header:

```
Authorization: Bearer <supabase_access_token>
```

The middleware extracts `user_id` and `org_id` from the JWT + profiles table.

---

## 2. Endpoints

### 2.1 Auth

#### `POST /api/v1/auth/callback`

Handle post-OAuth callback from Supabase. Creates org/profile if needed.

**Request:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

**Response (200):**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "John Doe",
    "avatar_url": "https://...",
    "role": "admin",
    "org": {
      "id": "uuid",
      "name": "My Studio",
      "slug": "my-studio",
      "credits": 10000
    }
  },
  "needs_onboarding": false
}
```

#### `POST /api/v1/auth/onboarding`

Create or join an organization (first-time users without invitation).

**Request:**
```json
{
  "action": "create",
  "org_name": "My Studio"
}
```

**Response (201):**
```json
{
  "org": {
    "id": "uuid",
    "name": "My Studio",
    "slug": "my-studio",
    "credits": 0
  }
}
```

---

### 2.2 Generations

#### `POST /api/v1/generations`

Start a new image generation batch.

**Request (multipart/form-data or JSON):**
```json
{
  "reference_image_urls": ["https://bucket.tansil.pro/ref1.png"],
  "product_image_urls": ["https://bucket.tansil.pro/prod1.png"],
  "initial_prompt": "Think as a lifestyle photographer...",
  "number_of_prompts": 3,
  "image_provider_slug": "kieai-nano-banana",
  "aspect_ratio": "9:16",
  "quality": "high"
}
```

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `reference_image_urls` | No | `[]` | Style/brand reference images |
| `product_image_urls` | **Yes** | — | Product images to feature |
| `initial_prompt` | **Yes** | — | Creative direction text |
| `number_of_prompts` | No | `3` | How many prompt variants |
| `image_provider_slug` | No | Default provider (priority 0) | Which image gen provider to use |
| `aspect_ratio` | No | Provider's default | Image aspect ratio |
| `quality` | No | Provider's default | Image quality level |

**Response (202 Accepted):**
```json
{
  "generation": {
    "id": "uuid",
    "status": "pending",
    "number_of_prompts": 3,
    "image_provider": {
      "slug": "kieai-nano-banana",
      "name": "kie.ai Nano Banana Pro",
      "cost_per_image": 10
    },
    "estimated_credits": 30,
    "created_at": "2026-02-10T10:00:00Z"
  }
}
```

**Errors:**
- `402` — Insufficient credits
- `400` — No product images provided
- `400` — Invalid provider slug

#### `GET /api/v1/generations`

List generations for the current user (or all org members for admin).

**Query params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| page | int | 1 | Page number |
| limit | int | 20 | Items per page |
| status | string | all | Filter by status |
| user_id | uuid | (self) | Admin only: filter by user |

**Response (200):**
```json
{
  "generations": [
    {
      "id": "uuid",
      "status": "completed",
      "initial_prompt": "Think as...",
      "number_of_prompts": 3,
      "credits_used": 21,
      "images": [
        {
          "id": "uuid",
          "status": "success",
          "image_url": "https://bucket.tansil.pro/gen1.jpg",
          "prompt_text": "A sleek..."
        }
      ],
      "user": {
        "id": "uuid",
        "display_name": "John"
      },
      "created_at": "2026-02-10T10:00:00Z",
      "completed_at": "2026-02-10T10:01:30Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

#### `GET /api/v1/generations/:id`

Get a single generation with all images and their statuses. Used for polling.

**Response (200):**
```json
{
  "generation": {
    "id": "uuid",
    "status": "generating",
    "initial_prompt": "...",
    "generated_prompts": ["prompt1", "prompt2", "prompt3"],
    "reference_images": [{ "url": "...", "analysis": "..." }],
    "product_images": [{ "url": "..." }],
    "image_provider": {
      "slug": "kieai-nano-banana",
      "name": "kie.ai Nano Banana Pro"
    },
    "generation_config": {
      "aspect_ratio": "9:16",
      "quality": "high"
    },
    "images": [
      { "id": "uuid", "status": "success", "image_url": "https://...", "prompt_index": 0 },
      { "id": "uuid", "status": "processing", "image_url": null, "prompt_index": 1 },
      { "id": "uuid", "status": "pending", "image_url": null, "prompt_index": 2 }
    ],
    "credits_used": 10,
    "created_at": "..."
  }
}
```

---

### 2.3 Callbacks (No Auth — Verified by task_id lookup)

#### `POST /api/v1/callbacks/:provider_slug`

Receive image generation results from external providers. The `:provider_slug` routes to the correct callback parser.

**Example: `POST /api/v1/callbacks/kie`** (kie.ai provider)

**Request (from kie.ai):**
```json
{
  "code": 200,
  "data": {
    "taskId": "abc123",
    "state": "success",
    "resultJson": "{\"resultUrls\":[\"https://tempfile...\"]}",
    "model": "seedream/4.5-edit",
    "costTime": 51,
    "createTime": 1769704716000,
    "completeTime": 1769704767000
  },
  "msg": "Playground task completed successfully."
}
```

**Response (200):** `{ "ok": true }`

**Internal actions:**
1. Route to provider's `ParseCallback()` based on `:provider_slug`
2. Find `generation_images` by `task_id` + `provider_slug`
3. Download image from provider's temp URL
4. Re-upload to R2 with permanent URL
5. Update `generation_images` record
6. Check if all images for this generation are done → update generation status
7. Deduct credits from org (using `provider.cost_per_use`)

**Adding a new provider's callback:** Each provider implements `ParseCallback(body []byte) → ImageGenResult`. The route handler is generic — it only needs the slug to dispatch.

---

### 2.4 Gallery

#### `GET /api/v1/gallery`

Get all completed images for the user (or org for admin).

**Query params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| page | int | 1 | Page number |
| limit | int | 40 | Items per page |
| user_id | uuid | (self) | Admin: filter by member |
| search | string | | Search in prompts |

**Response (200):**
```json
{
  "images": [
    {
      "id": "uuid",
      "image_url": "https://bucket.tansil.pro/...",
      "prompt_text": "A sleek minimalist...",
      "generation_id": "uuid",
      "user": { "id": "uuid", "display_name": "John" },
      "created_at": "2026-02-10T10:01:30Z"
    }
  ],
  "total": 200,
  "page": 1,
  "limit": 40
}
```

---

### 2.5 Upload

#### `POST /api/v1/uploads`

Upload an image file to R2. Returns the public URL.

**Request:** `multipart/form-data` with `file` field

**Response (201):**
```json
{
  "url": "https://bucket.tansil.pro/20260210120000_0.png",
  "filename": "20260210120000_0.png"
}
```

---

### 2.6 Admin — Organization

#### `GET /api/v1/admin/organization`

Get org details with stats. **Admin only.**

**Response (200):**
```json
{
  "organization": {
    "id": "uuid",
    "name": "My Studio",
    "credits": 104344.68,
    "member_count": 5,
    "total_generations": 1420,
    "total_images_generated": 4260
  }
}
```

#### `POST /api/v1/admin/credits`

Add credits to the organization. **Admin only.**

**Request:**
```json
{
  "amount": 5000,
  "description": "Monthly top-up"
}
```

**Response (200):**
```json
{
  "new_balance": 109344.68,
  "transaction_id": "uuid"
}
```

#### `GET /api/v1/admin/credits/history`

Credit transaction history. **Admin only.**

**Query params:** `page`, `limit`, `type` (topup|generation|refund)

**Response (200):**
```json
{
  "transactions": [
    {
      "id": "uuid",
      "type": "generation",
      "amount": -21,
      "balance_after": 104344.68,
      "user": { "display_name": "John" },
      "description": "Image generation",
      "created_at": "2026-02-10T10:00:00Z"
    }
  ],
  "total": 500,
  "page": 1
}
```

---

### 2.7 Admin — Members

#### `GET /api/v1/admin/members`

List all organization members with usage stats. **Admin only.**

**Response (200):**
```json
{
  "members": [
    {
      "id": "uuid",
      "email": "john@example.com",
      "display_name": "John Doe",
      "avatar_url": "https://...",
      "role": "member",
      "total_generations": 45,
      "total_credits_used": 315,
      "last_active_at": "2026-02-10T09:00:00Z"
    }
  ]
}
```

#### `POST /api/v1/admin/members/invite`

Invite a new member. **Admin only.**

**Request:**
```json
{
  "email": "newuser@example.com",
  "role": "member"
}
```

**Response (201):**
```json
{
  "invitation": {
    "id": "uuid",
    "email": "newuser@example.com",
    "role": "member",
    "expires_at": "2026-02-17T10:00:00Z"
  }
}
```

#### `DELETE /api/v1/admin/members/:id`

Remove a member from the organization. **Admin only.**

**Response (200):** `{ "ok": true }`

#### `PATCH /api/v1/admin/members/:id/role`

Change a member's role. **Admin only.**

**Request:**
```json
{
  "role": "admin"
}
```

**Response (200):** `{ "ok": true }`

---

### 2.8 Providers

#### `GET /api/v1/providers`

List all active providers. Available to all authenticated users. Used by the frontend to populate model picker dropdowns. **API keys are never exposed.**

**Query params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| category | string | all | Filter: `image_generation`, `llm`, `vision` |

**Response (200):**
```json
{
  "providers": [
    {
      "slug": "kieai-seedream",
      "name": "kie.ai Seedream 4.5 Edit",
      "category": "image_generation",
      "status": "active",
      "cost_per_use": 7,
      "config": {
        "default_aspect_ratio": "9:16",
        "default_quality": "high",
        "available_aspect_ratios": ["9:16", "1:1", "16:9", "4:5"],
        "available_qualities": ["high", "standard"],
        "model": "seedream/4.5-edit"
      }
    },
    {
      "slug": "kieai-nano-banana",
      "name": "kie.ai Nano Banana Pro",
      "category": "image_generation",
      "status": "active",
      "cost_per_use": 10,
      "config": {
        "default_aspect_ratio": "9:16",
        "default_quality": "high",
        "available_aspect_ratios": ["9:16", "1:1", "16:9"],
        "available_qualities": ["high", "standard", "ultra"],
        "model": "nano-banana-pro"
      }
    }
  ]
}
```

**Note:** The `config` object returned to the frontend is a **sanitized subset** — it excludes `base_url`, `endpoint`, `auth_type`, and other internal fields. Only UI-relevant config is returned: defaults, model name, and `available_aspect_ratios`/`available_qualities` arrays for populating frontend dropdowns.

---

### 2.9 Admin — Providers

#### `GET /api/v1/admin/providers`

List all providers with full config (including internal fields). **Admin only.**

**Response (200):**
```json
{
  "providers": [
    {
      "id": "uuid",
      "slug": "kieai-seedream",
      "name": "kie.ai Seedream 4.5 Edit",
      "category": "image_generation",
      "status": "active",
      "priority": 0,
      "cost_per_use": 7,
      "auth_type": "bearer",
      "has_api_key": true,
      "config": {
        "base_url": "https://api.kie.ai",
        "endpoint": "/api/v1/jobs/createTask",
        "model": "seedream/4.5-edit",
        "default_aspect_ratio": "9:16",
        "default_quality": "high",
        "available_aspect_ratios": ["9:16", "1:1", "16:9", "4:5"],
        "available_qualities": ["high", "standard"],
        "timeout_seconds": 600,
        "callback_path": "kie"
      },
      "created_at": "2026-02-10T00:00:00Z",
      "updated_at": "2026-02-10T00:00:00Z"
    }
  ]
}
```

**Note:** `has_api_key` is a boolean — the actual key is never returned in API responses.

#### `POST /api/v1/admin/providers`

Create a new provider. **Admin only.**

**Request:**
```json
{
  "category": "image_generation",
  "slug": "kieai-flux",
  "name": "kie.ai FLUX",
  "priority": 2,
  "config": {
    "base_url": "https://api.kie.ai",
    "endpoint": "/api/v1/jobs/createTask",
    "model": "flux-pro",
    "default_aspect_ratio": "1:1",
    "default_quality": "high",
    "timeout_seconds": 600,
    "callback_path": "kie"
  },
  "api_key": "sk-...",
  "auth_type": "bearer",
  "cost_per_use": 12
}
```

**Response (201):**
```json
{
  "provider": {
    "id": "uuid",
    "slug": "kieai-flux",
    "name": "kie.ai FLUX",
    "status": "active"
  }
}
```

#### `PATCH /api/v1/admin/providers/:slug`

Update a provider's config, status, priority, cost, or API key. **Admin only.**

**Request (partial update):**
```json
{
  "status": "inactive",
  "cost_per_use": 15,
  "config": {
    "model": "flux-pro-v2",
    "default_quality": "ultra"
  }
}
```

**Response (200):** `{ "ok": true }`

**Notes:**
- `config` is **merged** (not replaced) — only provided keys are updated
- To update the API key, include `"api_key": "new-key"` in the request
- Setting `"api_key": null` removes the key
- Changing `status` to `"inactive"` removes the provider from the active registry

#### `DELETE /api/v1/admin/providers/:slug`

Delete a provider. **Admin only.** Cannot delete if it has been used in any generation.

**Response (200):** `{ "ok": true }`
**Error (409):** `{ "error": { "code": "PROVIDER_IN_USE", "message": "Provider has been used in 42 generations" } }`

#### `POST /api/v1/admin/providers/:slug/test`

Test a provider's connectivity. Sends a minimal test request. **Admin only.**

**Response (200):**
```json
{
  "success": true,
  "latency_ms": 340,
  "message": "Connection successful"
}
```

**Response (200, failure):**
```json
{
  "success": false,
  "latency_ms": 5000,
  "message": "Timeout after 5000ms"
}
```

---

## 3. Error Response Format

All errors follow a consistent format:

```json
{
  "error": {
    "code": "INSUFFICIENT_CREDITS",
    "message": "Not enough credits. Required: 21, Available: 10",
    "status": 402
  }
}
```

Common error codes:
| Code | Status | Description |
|------|--------|-------------|
| UNAUTHORIZED | 401 | Invalid or missing token |
| FORBIDDEN | 403 | Not an admin / wrong org |
| NOT_FOUND | 404 | Resource not found |
| INSUFFICIENT_CREDITS | 402 | Not enough credits |
| VALIDATION_ERROR | 400 | Invalid request body |
| INVALID_PROVIDER | 400 | Unknown provider slug |
| PROVIDER_IN_USE | 409 | Cannot delete provider with existing generations |
| PROVIDER_UNAVAILABLE | 503 | All providers in chain failed |
| GENERATION_FAILED | 500 | External API failure |
