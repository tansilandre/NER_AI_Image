-- Create providers table (global config, managed by system admins)
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('image_generation', 'llm', 'vision')),
    api_key TEXT, -- encrypted
    base_url TEXT,
    model TEXT NOT NULL,
    config JSONB DEFAULT '{}',
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    cost_per_use BIGINT DEFAULT 0, -- credits per use
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_providers_category ON providers(category);
CREATE INDEX idx_providers_slug ON providers(slug);
CREATE INDEX idx_providers_is_active ON providers(is_active);
CREATE INDEX idx_providers_priority ON providers(priority);

-- Enable RLS
ALTER TABLE providers ENABLE ROW LEVEL SECURITY;

-- Create trigger for updated_at
CREATE TRIGGER update_providers_updated_at
    BEFORE UPDATE ON providers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
