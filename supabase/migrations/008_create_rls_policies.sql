-- Enable RLS on all tables and create policies

-- Organizations: Users can only see their own org
CREATE POLICY "Users can view own organization"
    ON organizations FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = organizations.id
            AND profiles.user_id = auth.uid()
        )
    );

CREATE POLICY "Only admins can update organization"
    ON organizations FOR UPDATE
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = organizations.id
            AND profiles.user_id = auth.uid()
            AND profiles.role = 'admin'
        )
    );

-- Profiles: Users can view profiles in their org
CREATE POLICY "Users can view profiles in their org"
    ON profiles FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM profiles AS my_profile
            WHERE my_profile.organization_id = profiles.organization_id
            AND my_profile.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can update own profile"
    ON profiles FOR UPDATE
    USING (user_id = auth.uid());

-- Generations: Users can view generations in their org
CREATE POLICY "Users can view generations in their org"
    ON generations FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = generations.organization_id
            AND profiles.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can create generations in their org"
    ON generations FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = generations.organization_id
            AND profiles.user_id = auth.uid()
        )
    );

-- Generation Images: Cascade through generations
CREATE POLICY "Users can view generation images in their org"
    ON generation_images FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM generations
            JOIN profiles ON profiles.organization_id = generations.organization_id
            WHERE generation_images.generation_id = generations.id
            AND profiles.user_id = auth.uid()
        )
    );

-- Credit Ledger: Users can view ledger for their org
CREATE POLICY "Users can view credit ledger in their org"
    ON credit_ledger FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = credit_ledger.organization_id
            AND profiles.user_id = auth.uid()
        )
    );

CREATE POLICY "Only admins can insert into credit ledger"
    ON credit_ledger FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = credit_ledger.organization_id
            AND profiles.user_id = auth.uid()
            AND profiles.role = 'admin'
        )
    );

-- Invitations: Only admins can manage
CREATE POLICY "Admins can manage invitations"
    ON invitations FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE profiles.organization_id = invitations.organization_id
            AND profiles.user_id = auth.uid()
            AND profiles.role = 'admin'
        )
    );

CREATE POLICY "Users can view invitations by email"
    ON invitations FOR SELECT
    USING (email = auth.email());

-- Providers: Read-only for all authenticated users
CREATE POLICY "Authenticated users can view active providers"
    ON providers FOR SELECT
    USING (is_active = TRUE);

CREATE POLICY "Only service role can manage providers"
    ON providers FOR ALL
    USING (false); -- Managed via service role key only
