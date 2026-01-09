-- Créer la base de données (à exécuter en tant que superuser)
-- psql -U postgres -c "CREATE DATABASE password_manager;"

-- migrations/001_initial_schema.sql
-- Connexion: psql -U postgres -d password_manager -f migrations/001_initial_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    master_password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(255) NOT NULL,
    public_key TEXT,
    private_key TEXT,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    backup_codes TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Vaults table
CREATE TABLE IF NOT EXISTS app.vaults (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES app.users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    website VARCHAR(500),
    username VARCHAR(255),
    encrypted_data TEXT NOT NULL,
    encryption_salt VARCHAR(255) NOT NULL,
    nonce VARCHAR(255) NOT NULL,
    folder VARCHAR(100),
    favorite BOOLEAN DEFAULT FALSE,
    last_used TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);


-- Shared passwords table
CREATE TABLE IF NOT EXISTS shared_passwords (
                                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vault_id UUID REFERENCES vaults(id) ON DELETE CASCADE,
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    recipient_id UUID REFERENCES users(id) ON DELETE SET NULL,
    recipient_email VARCHAR(255) NOT NULL,
    encrypted_data TEXT NOT NULL,
    share_token VARCHAR(100) UNIQUE NOT NULL,
    expires_at TIMESTAMP,
    max_views INTEGER,
    view_count INTEGER DEFAULT 0,
    require_password BOOLEAN DEFAULT FALSE,
    share_password_hash VARCHAR(255),
    can_view BOOLEAN DEFAULT TRUE,
    can_copy BOOLEAN DEFAULT TRUE,
    can_edit BOOLEAN DEFAULT FALSE,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed TIMESTAMP
    );

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
                                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    vault_id UUID REFERENCES vaults(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_vaults_user_id ON vaults(user_id);
CREATE INDEX IF NOT EXISTS idx_vaults_folder ON vaults(folder);
CREATE INDEX IF NOT EXISTS idx_vaults_created_at ON vaults(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shared_passwords_owner ON shared_passwords(owner_id);
CREATE INDEX IF NOT EXISTS idx_shared_passwords_recipient ON shared_passwords(recipient_id);
CREATE INDEX IF NOT EXISTS idx_shared_passwords_token ON shared_passwords(share_token);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS vaults_updated_at ON vaults;
CREATE TRIGGER vaults_updated_at
    BEFORE UPDATE ON vaults
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Insert test user (password: Test123!)
-- This is for development only, remove in production
INSERT INTO users (email, master_password_hash, salt, two_factor_enabled)
VALUES (
           'test@example.com',
           'hashed_password_here',
           'salt_here',
           false
       ) ON CONFLICT (email) DO NOTHING;