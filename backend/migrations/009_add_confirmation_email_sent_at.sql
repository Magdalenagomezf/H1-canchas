-- 009: tracks whether the booking confirmation email was already sent.
-- NULL = not sent yet. Used as a send-once claim (see ClaimConfirmationEmail).
--
-- Run with: psql -v ON_ERROR_STOP=1 -f 009_add_confirmation_email_sent_at.sql

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS confirmation_email_sent_at TIMESTAMP NULL;
