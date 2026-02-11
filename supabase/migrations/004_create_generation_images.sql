-- Create generation_images table
CREATE TABLE generation_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    generation_id UUID NOT NULL REFERENCES generations(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    image_url TEXT,
    r2_key TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    task_id TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX idx_generation_images_generation_id ON generation_images(generation_id);
CREATE INDEX idx_generation_images_task_id ON generation_images(task_id);
CREATE INDEX idx_generation_images_status ON generation_images(status);

-- Enable RLS
ALTER TABLE generation_images ENABLE ROW LEVEL SECURITY;

-- Create trigger for updated_at
CREATE TRIGGER update_generation_images_updated_at
    BEFORE UPDATE ON generation_images
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
