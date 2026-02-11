# Database Schema Document

## NER Studio — Supabase PostgreSQL

**Version:** 1.0
**Date:** 2026-02-10

---

## 1. Overview

All tables live in Supabase PostgreSQL. We use `auth.users` (managed by Supabase Auth) for authentication and our own `public.*` tables for application data.

---

## 2. Entity Relationship Diagram

```
auth.users (Supabase managed)
    │
    │ 1:1
    ▼
┌──────────┐       ┌──────────────┐       ┌──────────────┐
│ profiles │──────>│ organizations│       │  providers   │
│          │  N:1  │              │       │  (global)    │
└──────────┘       └──────┬───────┘       └──────┬───────┘
    │                      │                      │
    │ 1:N                  │ 1:N                  │ referenced by
    ▼                      ▼                      ▼
┌──────────────┐   ┌──────────────────┐   ┌──────────────┐
│ generations  │   │ credit_ledger    │   │ generations  │
│              │   │                  │   │ (provider_id)│
└──────┬───────┘   └──────────────────┘   └──────────────┘
       │
       │ 1:N
       ▼
┌──────────────────┐
│ generation_images│
│ (provider_id)    │
└──────────────────┘
```

---

## 3. Tables

### 3.1 `organizations`

The billing and team entity.

```sql
CREATE TABLE organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    credits         DECIMAL(12,2) NOT NULL DEFAULT 0,
    -- cost_per_image is now per-provider (providers.cost_per_use)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
```

### 3.2 `profiles`

Extends Supabase `auth.users` with app-specific data.

```sql
CREATE TYPE user_role AS ENUM ('admin', 'member');

CREATE TABLE profiles (
    id              UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role            user_role NOT NULL DEFAULT 'member',
    display_name    TEXT,
    avatar_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_org_id ON profiles(org_id);
```

### 3.3 `providers`

Global registry of all external AI providers. Managed by system admin.

```sql
CREATE TYPE provider_category AS ENUM ('image_generation', 'llm', 'vision');
CREATE TYPE provider_status AS ENUM ('active', 'inactive', 'error');

CREATE TABLE providers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category        provider_category NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    -- Unique key: "kieai-seedream", "kieai-nano-banana", "google-gemini", etc.
    name            TEXT NOT NULL,
    -- Display name: "kie.ai Seedream 4.5 Edit", "Nano Banana Pro", etc.
    status          provider_status NOT NULL DEFAULT 'active',
    priority        INT NOT NULL DEFAULT 0,
    -- Lower = tried first. Used for LLM fallback ordering.
    -- For image_generation: 0 = default model shown in UI
    config          JSONB NOT NULL DEFAULT '{}',
    -- Provider-specific config. Examples:
    --
    -- image_generation (kie.ai):
    -- {
    --   "base_url": "https://api.kie.ai",
    --   "endpoint": "/api/v1/jobs/createTask",
    --   "model": "seedream/4.5-edit",
    --   "default_aspect_ratio": "9:16",
    --   "default_quality": "high",
    --   "available_aspect_ratios": ["9:16", "1:1", "16:9", "4:5"],
    --   "available_qualities": ["high", "standard"],
    --   "timeout_seconds": 600,
    --   "callback_path": "kie"
    -- }
    --
    -- image_generation (nano-banana-pro via kie.ai):
    -- {
    --   "base_url": "https://api.kie.ai",
    --   "endpoint": "/api/v1/jobs/createTask",
    --   "model": "nano-banana-pro",
    --   "default_aspect_ratio": "9:16",
    --   "default_quality": "high",
    --   "available_aspect_ratios": ["9:16", "1:1", "16:9"],
    --   "available_qualities": ["high", "standard", "ultra"],
    --   "timeout_seconds": 600,
    --   "callback_path": "kie"
    -- }
    --
    -- llm (kie.ai proxy):
    -- {
    --   "base_url": "https://api.kie.ai",
    --   "endpoint": "/gemini-3-pro/v1/chat/completions",
    --   "request_format": "openai",
    --   "image_format": "url",
    --   "model": "gemini-3-pro",
    --   "timeout_seconds": 600,
    --   "retry_attempts": 2,
    --   "error_code_for_fallback": 500
    -- }
    --
    -- llm (Google Gemini direct):
    -- {
    --   "base_url": "https://generativelanguage.googleapis.com",
    --   "endpoint": "/v1beta/models/gemini-3-pro-preview:generateContent",
    --   "request_format": "gemini",
    --   "image_format": "base64",
    --   "auth_type": "header",
    --   "timeout_seconds": 600,
    --   "retry_attempts": 2
    -- }
    --
    -- vision (OpenAI):
    -- {
    --   "base_url": "https://api.openai.com",
    --   "model": "gpt-4o",
    --   "max_description_length": 500,
    --   "timeout_seconds": 60
    -- }
    api_key_encrypted TEXT,
    -- Encrypted with PROVIDER_KEY_ENCRYPTION_SECRET env var
    -- NULL if auth uses a shared credential (e.g., same kie.ai key)
    auth_type       TEXT NOT NULL DEFAULT 'bearer',
    -- "bearer", "header", "query_param"
    cost_per_use    DECIMAL(8,2) NOT NULL DEFAULT 0,
    -- For image_generation: cost per image in credits
    -- For llm/vision: 0 (absorbed)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_providers_category ON providers(category);
CREATE INDEX idx_providers_category_priority ON providers(category, priority);
CREATE INDEX idx_providers_slug ON providers(slug);
```

### 3.4 `generations`

A single generation request (one click of "Generate").

```sql
CREATE TYPE generation_status AS ENUM ('pending', 'analyzing', 'prompting', 'generating', 'completed', 'failed');

CREATE TABLE generations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    status              generation_status NOT NULL DEFAULT 'pending',
    initial_prompt      TEXT NOT NULL,
    number_of_prompts   INT NOT NULL DEFAULT 3,
    reference_images    JSONB NOT NULL DEFAULT '[]',
    -- Format: [{ "url": "...", "analysis": "..." }]
    product_images      JSONB NOT NULL DEFAULT '[]',
    -- Format: [{ "url": "..." }]
    generated_prompts   JSONB DEFAULT '[]',
    -- Format: ["prompt1", "prompt2", ...]
    image_provider_id   UUID REFERENCES providers(id),
    -- Which image generation provider was used
    llm_provider_id     UUID REFERENCES providers(id),
    -- Which LLM provider actually succeeded (after fallback chain)
    vision_provider_id  UUID REFERENCES providers(id),
    -- Which vision provider analyzed references
    generation_config   JSONB DEFAULT '{}',
    -- User-selected overrides: { "aspect_ratio": "1:1", "quality": "high", "model": "nano-banana-pro" }
    credits_used        DECIMAL(8,2) NOT NULL DEFAULT 0,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generations_user_id ON generations(user_id);
CREATE INDEX idx_generations_org_id ON generations(org_id);
CREATE INDEX idx_generations_status ON generations(status);
CREATE INDEX idx_generations_created_at ON generations(created_at DESC);
CREATE INDEX idx_generations_image_provider ON generations(image_provider_id);
```

### 3.5 `generation_images`

Individual generated images (one per prompt/task).

```sql
CREATE TYPE image_status AS ENUM ('pending', 'processing', 'success', 'failed');

CREATE TABLE generation_images (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    generation_id   UUID NOT NULL REFERENCES generations(id) ON DELETE CASCADE,
    provider_id     UUID REFERENCES providers(id),
    -- Which image generation provider produced this image
    prompt_index    INT NOT NULL,
    prompt_text     TEXT NOT NULL,
    status          image_status NOT NULL DEFAULT 'pending',
    task_id         TEXT,
    -- External provider's task ID for tracking (e.g. kie.ai task ID)
    provider_slug   TEXT,
    -- Denormalized for fast callback routing (e.g. "kieai")
    image_url       TEXT,
    -- Final R2 URL after re-upload
    source_url      TEXT,
    -- Original provider's temporary URL
    cost_time       INT,
    -- Generation time in seconds (from provider callback)
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generation_images_generation_id ON generation_images(generation_id);
CREATE INDEX idx_generation_images_task_id ON generation_images(task_id);
CREATE INDEX idx_generation_images_status ON generation_images(status);
CREATE INDEX idx_generation_images_provider ON generation_images(provider_slug);
```

### 3.6 `credit_ledger`

Immutable log of all credit transactions.

```sql
CREATE TYPE credit_type AS ENUM ('topup', 'generation', 'refund', 'adjustment');

CREATE TABLE credit_ledger (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES profiles(id),
    -- NULL for admin top-ups
    type            credit_type NOT NULL,
    amount          DECIMAL(12,2) NOT NULL,
    -- Positive for topup/refund, negative for generation
    balance_after   DECIMAL(12,2) NOT NULL,
    description     TEXT,
    generation_id   UUID REFERENCES generations(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_ledger_org_id ON credit_ledger(org_id);
CREATE INDEX idx_credit_ledger_user_id ON credit_ledger(user_id);
CREATE INDEX idx_credit_ledger_created_at ON credit_ledger(created_at DESC);
```

### 3.7 `invitations`

Pending org invitations.

```sql
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'expired', 'revoked');

CREATE TABLE invitations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invited_by      UUID NOT NULL REFERENCES profiles(id),
    email           TEXT NOT NULL,
    role            user_role NOT NULL DEFAULT 'member',
    status          invitation_status NOT NULL DEFAULT 'pending',
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_org_id ON invitations(org_id);
```

---

## 4. Row Level Security (RLS) Policies

```sql
-- Enable RLS on all tables
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE generations ENABLE ROW LEVEL SECURITY;
ALTER TABLE generation_images ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_ledger ENABLE ROW LEVEL SECURITY;
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

-- Providers: readable by all authenticated users (needed for UI model picker)
-- API keys are encrypted and never exposed to frontend
CREATE POLICY "Authenticated users can read providers"
    ON providers FOR SELECT
    USING (auth.role() = 'authenticated');

-- Providers: only service_role can write (backend manages via admin endpoints)
-- No INSERT/UPDATE/DELETE policy for authenticated users

-- Profiles: users can read their own org members
CREATE POLICY "Users can view own org members"
    ON profiles FOR SELECT
    USING (org_id = (SELECT org_id FROM profiles WHERE id = auth.uid()));

-- Generations: users can view own org generations
CREATE POLICY "Users can view own generations"
    ON generations FOR SELECT
    USING (org_id = (SELECT org_id FROM profiles WHERE id = auth.uid()));

-- Generations: users can insert own generations
CREATE POLICY "Users can create generations"
    ON generations FOR INSERT
    WITH CHECK (user_id = auth.uid());

-- Generation images: follow generation access
CREATE POLICY "Users can view own org generation images"
    ON generation_images FOR SELECT
    USING (generation_id IN (
        SELECT id FROM generations
        WHERE org_id = (SELECT org_id FROM profiles WHERE id = auth.uid())
    ));

-- Credit ledger: users can view own org ledger
CREATE POLICY "Users can view own org credit ledger"
    ON credit_ledger FOR SELECT
    USING (org_id = (SELECT org_id FROM profiles WHERE id = auth.uid()));
```

---

## 5. Database Functions

```sql
-- Deduct credits atomically
CREATE OR REPLACE FUNCTION deduct_credits(
    p_org_id UUID,
    p_user_id UUID,
    p_amount DECIMAL,
    p_generation_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_current_balance DECIMAL;
BEGIN
    -- Lock the org row
    SELECT credits INTO v_current_balance
    FROM organizations
    WHERE id = p_org_id
    FOR UPDATE;

    IF v_current_balance < p_amount THEN
        RETURN FALSE;
    END IF;

    UPDATE organizations
    SET credits = credits - p_amount, updated_at = NOW()
    WHERE id = p_org_id;

    INSERT INTO credit_ledger (org_id, user_id, type, amount, balance_after, generation_id, description)
    VALUES (p_org_id, p_user_id, 'generation', -p_amount, v_current_balance - p_amount, p_generation_id, 'Image generation');

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;
```

---

## 6. Supabase Trigger: Auto-create Profile

```sql
-- When a new user signs up via Supabase Auth, create a profile
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if user was invited
    DECLARE
        v_invitation RECORD;
    BEGIN
        SELECT * INTO v_invitation
        FROM invitations
        WHERE email = NEW.email
          AND status = 'pending'
          AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT 1;

        IF FOUND THEN
            INSERT INTO profiles (id, org_id, role, display_name, avatar_url)
            VALUES (
                NEW.id,
                v_invitation.org_id,
                v_invitation.role,
                NEW.raw_user_meta_data->>'full_name',
                NEW.raw_user_meta_data->>'avatar_url'
            );

            UPDATE invitations SET status = 'accepted' WHERE id = v_invitation.id;
        END IF;
        -- If no invitation, user must create/join org via onboarding flow
    END;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION handle_new_user();
```

---

## 7. Provider Seed Data

Default providers to insert on first migration. API keys are added via admin panel after deployment.

```sql
-- Image Generation Providers
INSERT INTO providers (category, slug, name, priority, config, auth_type, cost_per_use) VALUES
('image_generation', 'kieai-seedream', 'kie.ai Seedream 4.5 Edit', 0, '{
  "base_url": "https://api.kie.ai",
  "endpoint": "/api/v1/jobs/createTask",
  "model": "seedream/4.5-edit",
  "default_aspect_ratio": "9:16",
  "default_quality": "high",
  "available_aspect_ratios": ["9:16", "1:1", "16:9", "4:5"],
  "available_qualities": ["high", "standard"],
  "timeout_seconds": 600,
  "callback_path": "kie"
}', 'bearer', 7.00),

('image_generation', 'kieai-nano-banana', 'kie.ai Nano Banana Pro', 1, '{
  "base_url": "https://api.kie.ai",
  "endpoint": "/api/v1/jobs/createTask",
  "model": "nano-banana-pro",
  "default_aspect_ratio": "9:16",
  "default_quality": "high",
  "available_aspect_ratios": ["9:16", "1:1", "16:9"],
  "available_qualities": ["high", "standard", "ultra"],
  "timeout_seconds": 600,
  "callback_path": "kie"
}', 'bearer', 10.00);

-- LLM Providers (ordered by fallback priority)
INSERT INTO providers (category, slug, name, priority, config, auth_type, cost_per_use) VALUES
('llm', 'kieai-gemini3', 'kie.ai Gemini 3 Pro', 0, '{
  "base_url": "https://api.kie.ai",
  "endpoint": "/gemini-3-pro/v1/chat/completions",
  "request_format": "openai",
  "image_format": "url",
  "timeout_seconds": 600,
  "retry_attempts": 2,
  "error_code_for_fallback": 500,
  "reasoning_effort": "low"
}', 'bearer', 0),

('llm', 'kieai-gemini25', 'kie.ai Gemini 2.5 Pro', 1, '{
  "base_url": "https://api.kie.ai",
  "endpoint": "/gemini-2.5-pro/v1/chat/completions",
  "request_format": "openai",
  "image_format": "url",
  "timeout_seconds": 600,
  "retry_attempts": 2,
  "error_code_for_fallback": 500,
  "reasoning_effort": "low"
}', 'bearer', 0),

('llm', 'google-gemini', 'Google Gemini 3 Pro (Direct)', 2, '{
  "base_url": "https://generativelanguage.googleapis.com",
  "endpoint": "/v1beta/models/gemini-3-pro-preview:generateContent",
  "request_format": "gemini",
  "image_format": "base64",
  "timeout_seconds": 600,
  "retry_attempts": 2,
  "thinking_level": "low"
}', 'header', 0);

-- Vision Providers
INSERT INTO providers (category, slug, name, priority, config, auth_type, cost_per_use) VALUES
('vision', 'openai-vision', 'OpenAI GPT-4o Vision', 0, '{
  "base_url": "https://api.openai.com",
  "model": "gpt-4o",
  "max_description_length": 500,
  "timeout_seconds": 60
}', 'bearer', 0);
```
