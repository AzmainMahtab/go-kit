-- +goose Up
CREATE TABLE identity.sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    refresh_jti TEXT NOT NULL UNIQUE,
    ip TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id ON identity.sessions(user_id);
CREATE INDEX idx_sessions_refresh_jti ON identity.sessions(refresh_jti);
CREATE INDEX idx_sessions_expires_at ON identity.sessions(expires_at);

-- +goose Down
DROP TABLE identity.sessions;
