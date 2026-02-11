-- Seed default providers

-- Vision Provider
INSERT INTO providers (slug, name, category, base_url, model, priority, is_active, cost_per_use, config)
VALUES (
    'openai-gpt4o',
    'OpenAI GPT-4o Vision',
    'vision',
    'https://api.openai.com',
    'gpt-4o',
    0,
    true,
    5,
    '{"timeout_ms": 30000}'::jsonb
);

-- LLM Providers (fallback chain)
INSERT INTO providers (slug, name, category, base_url, model, priority, is_active, cost_per_use, config)
VALUES (
    'kieai-gemini3',
    'Kie.ai Gemini 3.0',
    'llm',
    'https://api.kie.ai',
    'gemini-3.0',
    0,
    true,
    3,
    '{"timeout_ms": 60000, "error_code_for_fallback": ["timeout", "rate_limit"]}'::jsonb
);

INSERT INTO providers (slug, name, category, base_url, model, priority, is_active, cost_per_use, config)
VALUES (
    'kieai-gemini25',
    'Kie.ai Gemini 2.5',
    'llm',
    'https://api.kie.ai',
    'gemini-2.5',
    1,
    true,
    2,
    '{"timeout_ms": 60000, "error_code_for_fallback": ["timeout", "rate_limit"]}'::jsonb
);

INSERT INTO providers (slug, name, category, base_url, model, priority, is_active, cost_per_use, config)
VALUES (
    'google-gemini',
    'Google Gemini (Direct)',
    'llm',
    'https://generativelanguage.googleapis.com',
    'gemini-2.0-flash',
    2,
    true,
    2,
    '{"timeout_ms": 60000}'::jsonb
);

-- Image Generation Providers
INSERT INTO providers (slug, name, category, base_url, model, priority, is_active, cost_per_use, config)
VALUES (
    'kieai-seedream',
    'Kie.ai Seedream',
    'image_generation',
    'https://api.kie.ai',
    'seedream-v1',
    0,
    true,
    10,
    '{"timeout_ms": 120000}'::jsonb
);

INSERT INTO providers (slug, name, category, base_url, model, priority, is_active, cost_per_use, config)
VALUES (
    'kieai-nano',
    'Kie.ai Nano Banana Pro',
    'image_generation',
    'https://api.kie.ai',
    'nano-banana-pro',
    1,
    true,
    8,
    '{"timeout_ms": 120000}'::jsonb
);
