-- Note: RLS (Row Level Security) is a PostgreSQL feature
-- These policies enforce data access rules at the database level
-- Applications should also enforce these rules, but RLS provides defense in depth

-- Enable RLS on tables
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE generations ENABLE ROW LEVEL SECURITY;
ALTER TABLE generation_images ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_ledger ENABLE ROW LEVEL SECURITY;

-- Note: Since we're using standard PostgreSQL (not Supabase),
-- we need to implement row-level security in application code
-- These policies would require the current_setting('app.current_user_id') to be set

-- For now, we skip creating RLS policies since we don't have
-- the Supabase auth integration. Application-level authorization
-- is implemented in the Go middleware.

-- If you want to enable RLS with standard PostgreSQL,
-- you would need to:
-- 1. Set a session variable with the current user ID
-- 2. Create policies that check this variable
-- 3. Use database roles for each user (complex)

-- Example (not enabled by default):
-- CREATE POLICY "Users can view own organization"
--     ON organizations FOR SELECT
--     USING (id = current_setting('app.current_org_id', true)::uuid);
