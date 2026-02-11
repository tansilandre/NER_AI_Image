-- Create credit_ledger table (immutable transaction log)
CREATE TABLE credit_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL, -- positive = add, negative = deduct
    type TEXT NOT NULL CHECK (type IN ('generation', 'refund', 'purchase', 'adjustment')),
    description TEXT NOT NULL,
    generation_id UUID REFERENCES generations(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_credit_ledger_organization_id ON credit_ledger(organization_id);
CREATE INDEX idx_credit_ledger_user_id ON credit_ledger(user_id);
CREATE INDEX idx_credit_ledger_created_at ON credit_ledger(created_at DESC);

-- Enable RLS
ALTER TABLE credit_ledger ENABLE ROW LEVEL SECURITY;

-- Function to atomically deduct credits
CREATE OR REPLACE FUNCTION deduct_credits(
    p_organization_id UUID,
    p_user_id UUID,
    p_amount BIGINT,
    p_description TEXT,
    p_generation_id UUID DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    current_credits BIGINT;
BEGIN
    -- Lock the organization row
    SELECT credits INTO current_credits
    FROM organizations
    WHERE id = p_organization_id
    FOR UPDATE;

    -- Check sufficient credits
    IF current_credits < p_amount THEN
        RETURN FALSE;
    END IF;

    -- Deduct credits
    UPDATE organizations
    SET credits = credits - p_amount,
        updated_at = NOW()
    WHERE id = p_organization_id;

    -- Record in ledger
    INSERT INTO credit_ledger (
        id,
        organization_id,
        user_id,
        amount,
        type,
        description,
        generation_id
    ) VALUES (
        gen_random_uuid(),
        p_organization_id,
        p_user_id,
        -p_amount,
        'generation',
        p_description,
        p_generation_id
    );

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
