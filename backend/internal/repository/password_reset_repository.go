package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// passwordResetRepo stores one-time password reset tokens (only their hash).
// Every time comparison uses the database clock (now()), so the app server
// clock and time zone never matter.
type passwordResetRepo struct {
	db *sqlx.DB
}

func NewPasswordResetRepository(db *sqlx.DB) *passwordResetRepo {
	return &passwordResetRepo{db: db}
}

// CreateIfNoRecent inserts a token for the user unless another one was
// created less than cooldown ago. created reports whether it was inserted.
// The check and the insert are one statement.
func (r *passwordResetRepo) CreateIfNoRecent(ctx context.Context, userID int64, tokenHash string, ttl, cooldown time.Duration) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
        INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
        SELECT $1, $2, now() + make_interval(secs => $3)
        WHERE NOT EXISTS (
            SELECT 1 FROM password_reset_tokens
            WHERE user_id = $1
            AND created_at > now() - make_interval(secs => $4)
        )`,
		userID, tokenHash, ttl.Seconds(), cooldown.Seconds())
	if err != nil {
		return false, fmt.Errorf("passwordResetRepo.CreateIfNoRecent: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("passwordResetRepo.CreateIfNoRecent: %w", err)
	}
	return n == 1, nil
}

// ResetPassword redeems a token: if it exists, is unused, has not expired and
// its user is active, it sets the user's password hash and marks every unused
// token of that user as used (the redeemed one included), in one transaction.
// ok is false when the token is not redeemable; nothing changes then.
func (r *passwordResetRepo) ResetPassword(ctx context.Context, tokenHash, passwordHash string) (ok bool, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("passwordResetRepo.ResetPassword: %w", err)
	}
	defer func() {
		if err != nil || !ok {
			_ = tx.Rollback()
		}
	}()

	// FOR UPDATE locks the token row: two concurrent redeems of the same
	// token serialize, and the second sees used_at set.
	var userID int64
	err = tx.QueryRowContext(ctx, `
        SELECT t.user_id
        FROM password_reset_tokens t
        JOIN users u ON u.id = t.user_id
        WHERE t.token_hash = $1
        AND t.used_at IS NULL
        AND t.expires_at > now()
        AND u.is_active
        FOR UPDATE OF t`, tokenHash).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("passwordResetRepo.ResetPassword: %w", err)
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`, passwordHash, userID); err != nil {
		return false, fmt.Errorf("passwordResetRepo.ResetPassword: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
		return false, fmt.Errorf("passwordResetRepo.ResetPassword: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("passwordResetRepo.ResetPassword: %w", err)
	}
	return true, nil
}
