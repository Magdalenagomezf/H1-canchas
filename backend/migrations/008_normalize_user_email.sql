-- 008: emails are stored as lower(trim(email)), never as '', and are unique
-- case-insensitively.
--
-- Run with: psql -v ON_ERROR_STOP=1 -f 008_normalize_user_email.sql
--
-- Aborts (and rolls back) if two users would collide once lowercased. Nothing
-- is deleted or merged automatically: fix the listed emails by hand and re-run.

BEGIN;

-- Inline UNIQUE in 001 creates a constraint named users_email_key. It is
-- dropped first: otherwise lowercasing two emails that differ only by case
-- fails on it before the duplicate check below can report them. If the
-- check aborts, the transaction rolls back and the constraint is restored.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

UPDATE users SET email = NULL WHERE trim(email) = '';

UPDATE users SET email = lower(trim(email)) WHERE email IS NOT NULL;

DO $$
DECLARE
    duplicates text;
BEGIN
    SELECT string_agg(email, ', ' ORDER BY email)
    INTO duplicates
    FROM (
        SELECT email
        FROM users
        WHERE email IS NOT NULL
        GROUP BY email
        HAVING count(*) > 1
    ) d;

    IF duplicates IS NOT NULL THEN
        RAISE EXCEPTION 'duplicate emails after normalization, fix them manually and re-run: %', duplicates;
    END IF;
END $$;

CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));

COMMIT;
