-- 010: Google login. Accounts created with Google have no password and no
-- phone until the user completes "Mi perfil".
--
-- Run with: psql -v ON_ERROR_STOP=1 -f 010_google_login.sql
--
-- Safe on a populated database: every existing row has a password_hash, so the
-- login-method CHECK passes.

BEGIN;

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL; -- UNIQUE is kept; Postgres allows many NULLs
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_sub VARCHAR(255) NULL UNIQUE;

-- Every account needs at least one way to sign in.
ALTER TABLE users ADD CONSTRAINT users_login_method_chk
    CHECK (password_hash IS NOT NULL OR google_sub IS NOT NULL);

COMMIT;
