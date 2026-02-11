# Frontend Specification Document

## NER Studio — React + Vite

**Version:** 1.0
**Date:** 2026-02-10

---

## 1. Design System — Bauhaus

The UI follows Bauhaus design principles:

### 1.1 Core Principles

- **Geometric forms**: Circles, squares, triangles as structural elements
- **Primary colors**: Red (#E63946), Yellow (#F4A100), Blue (#1D3557) as accents
- **Neutral base**: White (#FFFFFF), Off-white (#F8F8F8), Black (#1A1A1A), Gray (#6B7280)
- **Grid-based layouts**: Strong alignment, mathematical spacing
- **Typography-first**: Bold sans-serif headings, clean body text
- **Functional minimalism**: Every element serves a purpose

### 1.2 Typography

```
Font family: Inter (body), Space Grotesk (headings)
Heading 1: 32px / bold / Space Grotesk
Heading 2: 24px / bold / Space Grotesk
Heading 3: 18px / semibold / Space Grotesk
Body: 14px / regular / Inter
Caption: 12px / regular / Inter
```

### 1.3 Spacing Scale

```
4px / 8px / 12px / 16px / 24px / 32px / 48px / 64px
```

### 1.4 Component Style

- **Cards**: Sharp corners (no border-radius or 2px max), subtle shadow
- **Buttons**: Geometric, bold fills, uppercase labels
- **Inputs**: Thick bottom-border style (not full border)
- **Icons**: Line-style, geometric
- **Color blocks**: Accent colors used as bold geometric shapes/bars, not gradients

---

## 2. Pages & Routes

```
/                       → Redirect to /generate
/login                  → Google SSO login page
/onboarding             → Create/join org (first-time users)
/generate               → Image generation workspace (main page)
/gallery                → User's generated images gallery
/gallery/:id            → Single generation detail view
/admin                  → Admin dashboard (redirect to /admin/credits)
/admin/credits          → Credit management + usage stats
/admin/members          → Member management
/admin/providers        → AI provider configuration (models, API keys, fallback chain)
/admin/gallery          → Org-wide gallery (all members)
```

---

## 3. Page Specifications

### 3.1 Login (`/login`)

**Layout:** Centered card on geometric background

**Elements:**
- NER Studio logo + tagline
- "Sign in with Google" button (Supabase Auth UI or custom)
- Bauhaus geometric decorations (colored shapes)

**Behavior:**
- If already authenticated → redirect to `/generate`
- On successful login → check if profile exists → redirect to `/generate` or `/onboarding`

---

### 3.2 Generate (`/generate`) — Main Page

**Layout:** Matches current MVP (see image.png) with Bauhaus styling

**Sections:**

#### Header Bar
- Logo: "NER Studio" + "AI Image Generation"
- Credit display: `{credits} credits` (top right)
- User avatar + dropdown (profile, logout)

#### Reference Images Section
- Label: "REFERENCE IMAGES"
- Upload slots (click to upload, drag & drop)
- Show thumbnails with delete button
- Max 5 reference images

#### Generation Prompt Section
- Label: "GENERATION PROMPT"
- Large textarea with placeholder
- Clear (X) button

#### Product Images Section
- Label: "PRODUCT IMAGES"
- Upload slots + "+ Add" button
- Show thumbnails with delete button
- At least 1 required

#### Model & Options Row
- Model picker dropdown: Lists active `image_generation` providers from `GET /api/v1/providers?category=image_generation`
  - Each option shows: model name + cost per image (e.g., "Seedream 4.5 — 7 credits", "Nano Banana Pro — 10 credits")
  - Default: provider with priority 0
- Aspect ratio picker: Options from selected provider's `config.available_aspect_ratios` array. Default from `config.default_aspect_ratio`. When model changes, dropdown options update.
- Quality picker: Options from selected provider's `config.available_qualities` array. Default from `config.default_quality`. When model changes, dropdown options update.

#### Controls Row
- Image count: "Images — [count] +" (stepper, min 1, max 20)
- Credit estimate: "{count * selectedModel.cost_per_use} credits ({selectedModel.cost_per_use} per image)" — updates dynamically when model changes
- Generate button: Bold yellow/orange, "Generate" with sparkle icon

#### Generated Images Grid
- Label: "Generated Images"
- Grid of image cards (responsive: 2-6 columns)
- Each card shows:
  - Image thumbnail (or loading skeleton/spinner)
  - Index label (#1, #2, etc.)
  - Status indicator (pending/processing/done)
- Click to expand/download

**Behavior:**
- On "Generate" click:
  1. Validate: at least 1 product image, prompt not empty
  2. Check sufficient credits (count × selected model's cost)
  3. POST to `/api/v1/generations` with `image_provider_slug`, `aspect_ratio`, `quality`
  4. Show skeleton cards for expected images
  5. Poll `GET /api/v1/generations/:id` every 5 seconds
  6. As each image completes, replace skeleton with actual image
  7. Stop polling when all images are done or failed
- On model change: recalculate credit estimate, update aspect ratio / quality dropdowns (options + defaults from new provider's config)

---

### 3.3 Gallery (`/gallery`)

**Layout:** Masonry or uniform grid

**Elements:**
- Search bar (search in prompts)
- Filter by date range
- Image grid with infinite scroll
- Each card: image thumbnail, prompt preview, date, user name
- Click → expand to detail view

**Detail View (`/gallery/:id`):**
- All images from that generation
- Full prompt text
- Reference images used
- Product images used
- Generation metadata (date, time taken, credits used)
- Download individual or all images (zip)

---

### 3.4 Admin — Credits (`/admin/credits`)

**Layout:** Dashboard with stats + table

**Elements:**
- Current balance (large number)
- "Add Credits" button → modal with amount + description
- Usage chart (bar chart: credits used per day, last 30 days)
- Usage breakdown by member (table: name, generations, credits used)
- Credit transaction history (table with pagination)

---

### 3.5 Admin — Members (`/admin/members`)

**Layout:** Table with actions

**Elements:**
- Member count header
- "Invite Member" button → modal (email + role)
- Members table:
  - Avatar, name, email, role, generations count, credits used, last active
  - Actions: change role (dropdown), remove (with confirmation)
- Pending invitations section

---

### 3.6 Admin — Providers (`/admin/providers`)

**Layout:** Card-based list grouped by category

**Sections:**

#### Image Generation Providers
- Card per provider showing:
  - Name, model, status (active/inactive toggle)
  - Cost per image (editable)
  - Priority (drag to reorder, or number input)
  - Default aspect ratio and quality
- "Add Provider" button → modal
- Each card has: Edit, Test Connection, Delete actions

#### LLM Providers (Prompt Generation)
- Card per provider showing:
  - Name, model, status, priority (fallback order)
  - Request format badge: "OpenAI" or "Gemini"
  - Image format badge: "URL" or "Base64"
- Drag to reorder priority (determines fallback chain order)
- "Add Provider" button → modal

#### Vision Providers (Reference Analysis)
- Card per provider showing:
  - Name, model, status
  - Only one can be active at a time (radio selection)

#### Add/Edit Provider Modal
- Fields:
  - Name (display name)
  - Slug (auto-generated from name, editable)
  - Category (dropdown: image_generation / llm / vision)
  - API Key (password field, shows "Key configured" if set)
  - Auth type (bearer / header / query_param)
  - Cost per use (for image_generation only)
  - Config (JSON editor for advanced settings — base_url, endpoint, model, etc.)
- "Test Connection" button in the modal — shows latency and success/failure
- Save button

**Behavior:**
- Changes take effect immediately (backend reloads provider registry)
- Deleting a provider that has been used shows a confirmation with usage count
- Test Connection calls `POST /api/v1/admin/providers/:slug/test`

---

### 3.7 Admin — Gallery (`/admin/gallery`)

**Layout:** Same as user gallery but with member filter

**Additional elements:**
- Member filter dropdown (view specific member's images or all)
- All other gallery features

---

## 4. Navigation

**Sidebar (persistent, left side):**

```
┌─────────────────┐
│  NER Studio      │
│  AI Image Gen    │
├─────────────────┤
│  ⬡ Generate      │  ← active state: bold + accent bar
│  ◻ Gallery       │
│                  │
│  ADMIN           │  ← only visible to admin role
│  ◻ Credits       │
│  ◻ Members       │
│  ◻ Providers     │  ← AI model configuration
│  ◻ All Gallery   │
├─────────────────┤
│  ○ John Doe      │  ← avatar + name
│  Logout          │
└─────────────────┘
```

- Bauhaus style: geometric icons, bold section dividers
- Collapsible on mobile → hamburger menu

---

## 5. State Management (Zustand)

```typescript
// stores/auth.ts
interface AuthStore {
  user: User | null
  org: Organization | null
  isLoading: boolean
  login: () => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

// stores/generation.ts
interface GenerationStore {
  referenceImages: UploadedImage[]
  productImages: UploadedImage[]
  prompt: string
  imageCount: number
  selectedProvider: Provider | null     // selected image gen provider
  availableProviders: Provider[]        // from GET /api/v1/providers
  aspectRatio: string                   // from provider default or user override
  quality: string                       // from provider default or user override
  isGenerating: boolean
  currentGeneration: Generation | null
  addReferenceImage: (file: File) => Promise<void>
  addProductImage: (file: File) => Promise<void>
  setPrompt: (prompt: string) => void
  setImageCount: (count: number) => void
  setProvider: (slug: string) => void   // updates aspectRatio/quality defaults
  setAspectRatio: (ratio: string) => void
  setQuality: (quality: string) => void
  fetchProviders: () => Promise<void>   // load available providers
  generate: () => Promise<void>
  pollStatus: (generationId: string) => void
  estimatedCredits: () => number        // computed: imageCount * provider.cost_per_use
  reset: () => void
}

// stores/gallery.ts
interface GalleryStore {
  images: GalleryImage[]
  total: number
  isLoading: boolean
  filters: GalleryFilters
  fetchImages: (page: number) => Promise<void>
  setFilters: (filters: Partial<GalleryFilters>) => void
}
```

---

## 6. Key Libraries

| Library | Purpose |
|---------|---------|
| `react` 19 | UI framework |
| `react-router` 7 | Routing |
| `@supabase/supabase-js` | Auth client |
| `zustand` | State management |
| `tailwindcss` 4 | Styling |
| `axios` | HTTP client |
| `react-dropzone` | File upload drag & drop |
| `recharts` | Charts (admin dashboard) |
| `sonner` | Toast notifications |
| `lucide-react` | Icons |

---

## 7. Responsive Breakpoints

```
sm: 640px    → Mobile
md: 768px    → Tablet
lg: 1024px   → Desktop
xl: 1280px   → Large desktop
```

- Mobile: Sidebar collapses, single-column grid
- Tablet: 2-column image grid
- Desktop: Full sidebar, 3-4 column grid
- Large: Up to 6-column grid
