-- 011: one-time tokens for "forgot my password".
--
-- Run with: psql -v ON_ERROR_STOP=1 -f 011_password_reset_tokens.sql
--
-- Only the SHA-256 of the token is stored: a database leak does not expose
-- usable links. Users are hard-deleted, so their tokens go with them.

BEGIN;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Cooldown lookup: latest token of a user.
CREATE INDEX IF NOT EXISTS password_reset_tokens_user_created_idx
    ON password_reset_tokens (user_id, created_at DESC);

COMMIT;
