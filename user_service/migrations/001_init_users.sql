CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_status_check
        CHECK (status IN ('active', 'blocked', 'pending')),
    CONSTRAINT users_username_not_blank_check
        CHECK (btrim(username) <> ''),
    CONSTRAINT users_username_length_check
        CHECK (char_length(username) BETWEEN 3 AND 32),
    CONSTRAINT users_username_no_spaces_check
        CHECK (username !~ '\s'),
    CONSTRAINT users_email_not_blank_check
        CHECK (btrim(email) <> ''),
    CONSTRAINT users_email_format_check
        CHECK (email ~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$')
);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    surname TEXT,
    name TEXT,
    middle_name TEXT,
    birth_date DATE,
    avatar_file_id TEXT,
    phone TEXT,
    CONSTRAINT user_profiles_surname_not_blank_check
        CHECK (surname IS NULL OR btrim(surname) <> ''),
    CONSTRAINT user_profiles_name_not_blank_check
        CHECK (name IS NULL OR btrim(name) <> ''),
    CONSTRAINT user_profiles_middle_name_not_blank_check
        CHECK (middle_name IS NULL OR btrim(middle_name) <> ''),
    CONSTRAINT user_profiles_phone_format_check
        CHECK (phone IS NULL OR phone ~ '^\+?[0-9]{10,15}$'),
    CONSTRAINT user_profiles_birth_date_range_check
        CHECK (
            birth_date IS NULL
            OR (birth_date >= DATE '1900-01-01' AND birth_date <= CURRENT_DATE)
        )
);

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme TEXT NOT NULL DEFAULT 'light',
    language TEXT NOT NULL DEFAULT 'ru',
    notes_view TEXT NOT NULL DEFAULT 'list',
    graph_enabled BOOLEAN NOT NULL DEFAULT true,
    event_email_enabled BOOLEAN NOT NULL DEFAULT true,
    default_remind_before_minutes INT NOT NULL DEFAULT 30,
    timezone TEXT NOT NULL DEFAULT 'Europe/Moscow',
    preferences JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_preferences_default_remind_before_minutes_check
        CHECK (default_remind_before_minutes BETWEEN 0 AND 10080),
    CONSTRAINT user_preferences_theme_check
        CHECK (theme IN ('light', 'dark')),
    CONSTRAINT user_preferences_language_check
        CHECK (language IN ('ru', 'en')),
    CONSTRAINT user_preferences_notes_view_check
        CHECK (notes_view IN ('list', 'cards')),
    CONSTRAINT user_preferences_timezone_not_blank_check
        CHECK (btrim(timezone) <> '')
);

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS user_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    status TEXT NOT NULL DEFAULT 'success',
    metadata JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_tokens_purpose_check
        CHECK (purpose IN ('reset_password', 'verify_email', 'magic_link'))
);

CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    visibility TEXT NOT NULL DEFAULT 'private',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT workspaces_visibility_check
        CHECK (visibility IN ('private', 'invite_only')),
    CONSTRAINT workspaces_name_not_blank_check
        CHECK (btrim(name) <> ''),
    CONSTRAINT workspaces_name_length_check
        CHECK (char_length(btrim(name)) BETWEEN 3 AND 120)
);

CREATE TABLE IF NOT EXISTS workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'viewer',
    status TEXT NOT NULL DEFAULT 'pending',
    joined_at TIMESTAMPTZ,
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    PRIMARY KEY (workspace_id, user_id),
    CONSTRAINT workspace_members_role_check
        CHECK (role IN ('owner', 'editor', 'viewer')),
    CONSTRAINT workspace_members_status_check
        CHECK (status IN ('active', 'pending', 'removed'))
);

CREATE TABLE IF NOT EXISTS workspace_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_created_at
    ON user_sessions (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_token_hash
    ON user_sessions (refresh_token_hash);

CREATE INDEX IF NOT EXISTS idx_user_actions_user_created_at
    ON user_actions (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_actions_entity
    ON user_actions (entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_user_tokens_user_purpose
    ON user_tokens (user_id, purpose);

CREATE INDEX IF NOT EXISTS idx_user_tokens_expires_at
    ON user_tokens (expires_at);

CREATE INDEX IF NOT EXISTS idx_workspaces_owner_user_id
    ON workspaces (owner_user_id);

CREATE INDEX IF NOT EXISTS idx_workspace_members_user_id
    ON workspace_members (user_id);

CREATE INDEX IF NOT EXISTS idx_workspace_invites_workspace_id
    ON workspace_invites (workspace_id);

CREATE INDEX IF NOT EXISTS idx_workspace_invites_email
    ON workspace_invites (email);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trg_users_set_updated_at'
    ) THEN
        CREATE TRIGGER trg_users_set_updated_at
            BEFORE UPDATE ON users
            FOR EACH ROW
            EXECUTE FUNCTION set_updated_at();
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trg_user_preferences_set_updated_at'
    ) THEN
        CREATE TRIGGER trg_user_preferences_set_updated_at
            BEFORE UPDATE ON user_preferences
            FOR EACH ROW
            EXECUTE FUNCTION set_updated_at();
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trg_workspaces_set_updated_at'
    ) THEN
        CREATE TRIGGER trg_workspaces_set_updated_at
            BEFORE UPDATE ON workspaces
            FOR EACH ROW
            EXECUTE FUNCTION set_updated_at();
    END IF;
END $$;
