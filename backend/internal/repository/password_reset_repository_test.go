package repository_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"H1-canchas/internal/repository"
)

const (
	resetTTL      = time.Hour
	resetCooldown = 2 * time.Minute
)

func hashOf(c byte) string { return strings.Repeat(string(c), 64) }

func passwordHashOf(t *testing.T, userID int64) string {
	t.Helper()
	var h string
	if err := testDB.Get(&h, `SELECT password_hash FROM users WHERE id = $1`, userID); err != nil {
		t.Fatalf("read password_hash: %v", err)
	}
	return h
}

func unusedTokens(t *testing.T, userID int64) int {
	t.Helper()
	var n int
	if err := testDB.Get(&n, `SELECT count(*) FROM password_reset_tokens WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
		t.Fatalf("count tokens: %v", err)
	}
	return n
}

// mustCreateToken inserts a token as test setup and stops the test if it was
// not created, so no test can pass because its setup silently failed.
func mustCreateToken(t *testing.T, repo interface {
	CreateIfNoRecent(context.Context, int64, string, time.Duration, time.Duration) (bool, error)
}, userID int64, tokenHash string) {
	t.Helper()
	created, err := repo.CreateIfNoRecent(context.Background(), userID, tokenHash, resetTTL, resetCooldown)
	if err != nil || !created {
		t.Fatalf("setup: created=%v err=%v", created, err)
	}
}

func TestPasswordResetRepo_CreateIfNoRecent_StoresTokenWithExpiry(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)

	created, err := repo.CreateIfNoRecent(context.Background(), userID, hashOf('a'), resetTTL, resetCooldown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected the token to be created")
	}

	var minutes float64
	if err := testDB.Get(&minutes,
		`SELECT EXTRACT(EPOCH FROM expires_at - created_at) / 60 FROM password_reset_tokens WHERE token_hash = $1`,
		hashOf('a')); err != nil {
		t.Fatalf("read token: %v", err)
	}
	if minutes < 59.9 || minutes > 60.1 {
		t.Errorf("token lasts %.2f minutes, want 60", minutes)
	}
}

func TestPasswordResetRepo_CreateIfNoRecent_CooldownBlocksSecondToken(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)
	ctx := context.Background()

	mustCreateToken(t, repo, userID, hashOf('a'))
	created, err := repo.CreateIfNoRecent(ctx, userID, hashOf('b'), resetTTL, resetCooldown)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if created {
		t.Error("second token inside the cooldown must not be created")
	}
	if n := unusedTokens(t, userID); n != 1 {
		t.Errorf("got %d tokens, want 1", n)
	}

	// Another user is not affected by this user's cooldown.
	other := seedUser(t)
	created, err = repo.CreateIfNoRecent(ctx, other, hashOf('c'), resetTTL, resetCooldown)
	if err != nil || !created {
		t.Errorf("other user: created=%v err=%v, want created", created, err)
	}
}

func TestPasswordResetRepo_CreateIfNoRecent_AllowedAfterCooldown(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)
	ctx := context.Background()

	mustCreateToken(t, repo, userID, hashOf('a'))
	testDB.MustExec(`UPDATE password_reset_tokens SET created_at = now() - interval '3 minutes'`)

	created, err := repo.CreateIfNoRecent(ctx, userID, hashOf('b'), resetTTL, resetCooldown)
	if err != nil || !created {
		t.Errorf("created=%v err=%v, want created", created, err)
	}
}

func TestPasswordResetRepo_ResetPassword_SetsHashAndInvalidatesAllTokens(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)
	ctx := context.Background()

	// Two pending tokens: an older one and the one being redeemed.
	mustCreateToken(t, repo, userID, hashOf('a'))
	testDB.MustExec(`UPDATE password_reset_tokens SET created_at = now() - interval '3 minutes'`)
	mustCreateToken(t, repo, userID, hashOf('b'))

	ok, err := repo.ResetPassword(ctx, hashOf('b'), "new-hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected the token to be redeemed")
	}
	if got := passwordHashOf(t, userID); got != "new-hash" {
		t.Errorf("password_hash = %q, want new-hash", got)
	}
	if n := unusedTokens(t, userID); n != 0 {
		t.Errorf("got %d unused tokens, want 0", n)
	}

	// The older token no longer works either.
	ok, err = repo.ResetPassword(ctx, hashOf('a'), "other-hash")
	if err != nil || ok {
		t.Errorf("older token: ok=%v err=%v, want not redeemable", ok, err)
	}
	if got := passwordHashOf(t, userID); got != "new-hash" {
		t.Errorf("password_hash = %q, want new-hash", got)
	}
}

func TestPasswordResetRepo_ResetPassword_GivesPasswordToGoogleOnlyAccount(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	ctx := context.Background()

	var userID int64
	if err := testDB.Get(&userID,
		`INSERT INTO users (name, email, google_sub, role) VALUES ('G', 'g@example.com', 'sub-1', 'customer') RETURNING id`); err != nil {
		t.Fatalf("seed google user: %v", err)
	}
	mustCreateToken(t, repo, userID, hashOf('a'))

	ok, err := repo.ResetPassword(ctx, hashOf('a'), "new-hash")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v, want redeemed", ok, err)
	}
	if got := passwordHashOf(t, userID); got != "new-hash" {
		t.Errorf("password_hash = %q, want new-hash", got)
	}
}

func TestPasswordResetRepo_ResetPassword_RejectsUnredeemableTokens(t *testing.T) {
	cases := []struct {
		name  string
		setup func(userID int64)
		hash  string
	}{
		{"unknown", func(int64) {}, hashOf('z')},
		{"used", func(int64) {
			testDB.MustExec(`UPDATE password_reset_tokens SET used_at = now()`)
		}, hashOf('a')},
		{"expired", func(int64) {
			testDB.MustExec(`UPDATE password_reset_tokens SET expires_at = now() - interval '1 second'`)
		}, hashOf('a')},
		{"inactive user", func(userID int64) {
			testDB.MustExec(`UPDATE users SET is_active = false WHERE id = $1`, userID)
		}, hashOf('a')},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			truncateAll(t)
			repo := repository.NewPasswordResetRepository(testDB)
			userID := seedUser(t)
			ctx := context.Background()
			mustCreateToken(t, repo, userID, hashOf('a'))
			tc.setup(userID)

			ok, err := repo.ResetPassword(ctx, tc.hash, "new-hash")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Error("expected not redeemable")
			}
			if got := passwordHashOf(t, userID); got != "hash" {
				t.Errorf("password_hash = %q, want unchanged", got)
			}
		})
	}
}

func TestPasswordResetRepo_ResetPassword_ConcurrentRedeemSucceedsOnce(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)
	ctx := context.Background()
	mustCreateToken(t, repo, userID, hashOf('a'))

	const n = 5
	var wg sync.WaitGroup
	results := make(chan bool, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := repo.ResetPassword(ctx, hashOf('a'), "new-hash")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			results <- ok
		}()
	}
	wg.Wait()
	close(results)

	wins := 0
	for ok := range results {
		if ok {
			wins++
		}
	}
	if wins != 1 {
		t.Errorf("token redeemed %d times, want 1", wins)
	}
}

func TestPasswordResetRepo_TokensDeletedWithUser(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)
	mustCreateToken(t, repo, userID, hashOf('a'))

	testDB.MustExec(`DELETE FROM users WHERE id = $1`, userID)

	if n := unusedTokens(t, userID); n != 0 {
		t.Errorf("got %d tokens after deleting the user, want 0", n)
	}
}

func TestPasswordResetRepo_IsRedeemable_MatchesResetPasswordRules(t *testing.T) {
	cases := []struct {
		name  string
		setup func(userID int64)
		hash  string
		want  bool
	}{
		{"fresh token", func(int64) {}, hashOf('a'), true},
		{"unknown", func(int64) {}, hashOf('z'), false},
		{"used", func(int64) {
			testDB.MustExec(`UPDATE password_reset_tokens SET used_at = now()`)
		}, hashOf('a'), false},
		{"expired", func(int64) {
			testDB.MustExec(`UPDATE password_reset_tokens SET expires_at = now() - interval '1 second'`)
		}, hashOf('a'), false},
		{"inactive user", func(userID int64) {
			testDB.MustExec(`UPDATE users SET is_active = false WHERE id = $1`, userID)
		}, hashOf('a'), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			truncateAll(t)
			repo := repository.NewPasswordResetRepository(testDB)
			userID := seedUser(t)
			mustCreateToken(t, repo, userID, hashOf('a'))
			tc.setup(userID)

			got, err := repo.IsRedeemable(context.Background(), tc.hash)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("IsRedeemable = %v, want %v", got, tc.want)
			}
			// Checking never uses the token up.
			var used int
			testDB.Get(&used, `SELECT count(*) FROM password_reset_tokens WHERE used_at IS NOT NULL AND token_hash = $1`, hashOf('a'))
			if tc.name == "fresh token" && used != 0 {
				t.Error("IsRedeemable marked the token as used")
			}
		})
	}
}

func TestPasswordResetRepo_IsRedeemable_FalseAfterRedeem(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPasswordResetRepository(testDB)
	userID := seedUser(t)
	mustCreateToken(t, repo, userID, hashOf('a'))
	ctx := context.Background()

	if ok, _ := repo.IsRedeemable(ctx, hashOf('a')); !ok {
		t.Fatal("fresh token should be redeemable")
	}
	if ok, err := repo.ResetPassword(ctx, hashOf('a'), "new-hash"); err != nil || !ok {
		t.Fatalf("redeem: ok=%v err=%v", ok, err)
	}
	if ok, _ := repo.IsRedeemable(ctx, hashOf('a')); ok {
		t.Error("a redeemed token must not be redeemable")
	}
}
