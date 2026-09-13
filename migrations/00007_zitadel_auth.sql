-- +goose Up
ALTER TABLE users
    ADD COLUMN oidc_issuer TEXT,
    ADD COLUMN oidc_subject TEXT,
    ALTER COLUMN password_hash DROP NOT NULL;

CREATE UNIQUE INDEX users_oidc_identity
    ON users (oidc_issuer, oidc_subject)
    WHERE oidc_issuer IS NOT NULL AND oidc_subject IS NOT NULL;

CREATE TABLE auth_sessions (
    token_hash BYTEA PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX auth_sessions_expires_at ON auth_sessions(expires_at);

-- Legacy refresh-token rows are retained for rollback safety, but are no
-- longer read or written after the Zitadel migration.

-- +goose Down
DROP TABLE auth_sessions;
DROP INDEX users_oidc_identity;

UPDATE users SET password_hash = 'disabled-after-oidc-rollback' WHERE password_hash IS NULL;

ALTER TABLE users
    DROP COLUMN oidc_issuer,
    DROP COLUMN oidc_subject,
    ALTER COLUMN password_hash SET NOT NULL;
