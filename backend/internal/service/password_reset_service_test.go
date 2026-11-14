package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"H1-canchas/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type fakeResetUsers struct {
	byEmail map[string]*domain.User
}

func (f *fakeResetUsers) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	return f.byEmail[strings.ToLower(email)], nil
}

type fakeResetRepo struct {
	createCalls  int
	userID       int64
	tokenHash    string
	ttl          time.Duration
	cooldown     time.Duration
	inCooldown   bool
	redeemable   map[string]bool
	resetHash    string
	resetPwdHash string
}

func (f *fakeResetRepo) CreateIfNoRecent(_ context.Context, userID int64, tokenHash string, ttl, cooldown time.Duration) (bool, error) {
	f.createCalls++
	f.userID, f.tokenHash, f.ttl, f.cooldown = userID, tokenHash, ttl, cooldown
	return !f.inCooldown, nil
}

func (f *fakeResetRepo) ResetPassword(_ context.Context, tokenHash, passwordHash string) (bool, error) {
	f.resetHash, f.resetPwdHash = tokenHash, passwordHash
	return f.redeemable[tokenHash], nil
}

type sentMail struct{ to, subject, html, text string }

type fakeSender struct {
	sent []sentMail
	err  error
}

func (f *fakeSender) Send(_ context.Context, to, subject, html, text string) error {
	f.sent = append(f.sent, sentMail{to, subject, html, text})
	return f.err
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func newResetFixture(users ...*domain.User) (*PasswordResetService, *fakeResetRepo, *fakeSender) {
	byEmail := map[string]*domain.User{}
	for _, u := range users {
		byEmail[strings.ToLower(*u.Email)] = u
	}
	repo := &fakeResetRepo{redeemable: map[string]bool{}}
	sender := &fakeSender{}
	svc := NewPasswordResetService(&fakeResetUsers{byEmail: byEmail}, repo, sender, "https://h1.example/")
	return svc, repo, sender
}

func activeUser(email string) *domain.User {
	return &domain.User{ID: 7, Name: "Ana", Email: &email, Role: domain.RoleCustomer, IsActive: true}
}

func TestPasswordReset_SendResetLink_EmailsOneTimeLink(t *testing.T) {
	svc, repo, sender := newResetFixture(activeUser("ana@example.com"))

	if err := svc.SendResetLink(context.Background(), "  ANA@example.com "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.sent) != 1 {
		t.Fatalf("sent %d mails, want 1", len(sender.sent))
	}
	mail := sender.sent[0]
	if mail.to != "ana@example.com" {
		t.Errorf("to = %q, want ana@example.com", mail.to)
	}

	const prefix = "https://h1.example/restablecer-contrasena#token="
	i := strings.Index(mail.text, prefix)
	if i < 0 {
		t.Fatalf("text body has no reset link: %q", mail.text)
	}
	token := strings.Fields(mail.text[i+len(prefix):])[0]
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		t.Errorf("token %q is not 32 bytes of base64url (len %d, err %v)", token, len(raw), err)
	}
	if !strings.Contains(mail.html, prefix+token) {
		t.Error("html body does not contain the same link")
	}
	for _, body := range []string{mail.text, mail.html} {
		if !strings.Contains(body, "1 hora") {
			t.Error("body does not say the link lasts 1 hora")
		}
	}

	if repo.userID != 7 || repo.tokenHash != sha256Hex(token) {
		t.Errorf("stored user %d hash %q, want user 7 and the token's SHA-256", repo.userID, repo.tokenHash)
	}
	if repo.ttl != time.Hour || repo.cooldown != 2*time.Minute {
		t.Errorf("ttl %v cooldown %v, want 1h and 2m", repo.ttl, repo.cooldown)
	}
}

func TestPasswordReset_SendResetLink_EachTokenIsDifferent(t *testing.T) {
	svc, repo, _ := newResetFixture(activeUser("ana@example.com"))
	ctx := context.Background()

	svc.SendResetLink(ctx, "ana@example.com")
	first := repo.tokenHash
	svc.SendResetLink(ctx, "ana@example.com")

	if first == repo.tokenHash {
		t.Error("two requests produced the same token")
	}
}

func TestPasswordReset_SendResetLink_SilentWhenNothingToSend(t *testing.T) {
	inactive := activeUser("off@example.com")
	inactive.IsActive = false

	cases := []struct {
		name       string
		email      string
		inCooldown bool
	}{
		{"unknown email", "nobody@example.com", false},
		{"malformed email", "not-an-email", false},
		{"inactive user", "off@example.com", false},
		{"cooldown", "ana@example.com", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, sender := newResetFixture(activeUser("ana@example.com"), inactive)
			repo.inCooldown = tc.inCooldown

			if err := svc.SendResetLink(context.Background(), tc.email); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sender.sent) != 0 {
				t.Errorf("sent %d mails, want 0", len(sender.sent))
			}
		})
	}
}

func TestPasswordReset_Confirm_SetsBcryptHashForTokenHash(t *testing.T) {
	svc, repo, _ := newResetFixture()
	repo.redeemable[sha256Hex("the-token")] = true

	if err := svc.Confirm(context.Background(), "the-token", "nueva-clave"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.resetHash != sha256Hex("the-token") {
		t.Errorf("looked up %q, want the token's SHA-256", repo.resetHash)
	}
	if bcrypt.CompareHashAndPassword([]byte(repo.resetPwdHash), []byte("nueva-clave")) != nil {
		t.Error("stored hash does not match the new password")
	}
	if cost, _ := bcrypt.Cost([]byte(repo.resetPwdHash)); cost != 12 {
		t.Errorf("bcrypt cost %d, want 12", cost)
	}
}

func TestPasswordReset_Confirm_InvalidToken(t *testing.T) {
	svc, _, _ := newResetFixture()

	for _, token := range []string{"unknown", ""} {
		if err := svc.Confirm(context.Background(), token, "nueva-clave"); !errors.Is(err, ErrInvalidResetToken) {
			t.Errorf("token %q: err = %v, want ErrInvalidResetToken", token, err)
		}
	}
}

func TestPasswordReset_Confirm_Accepts72BytePassword(t *testing.T) {
	svc, repo, _ := newResetFixture()
	repo.redeemable[sha256Hex("t")] = true

	if err := svc.Confirm(context.Background(), "t", strings.Repeat("a", 72)); err != nil {
		t.Errorf("72 bytes: unexpected error %v", err)
	}
}

func TestPasswordReset_Confirm_UsesRegisterPasswordRule(t *testing.T) {
	svc, repo, _ := newResetFixture()
	repo.redeemable[sha256Hex("t")] = true

	if err := svc.Confirm(context.Background(), "t", ""); !errors.Is(err, ErrPasswordRequired) {
		t.Errorf("empty: err = %v, want ErrPasswordRequired", err)
	}
	if err := svc.Confirm(context.Background(), "t", "12345"); !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("short: err = %v, want ErrPasswordTooShort", err)
	}
	if err := svc.Confirm(context.Background(), "t", strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
		t.Errorf("over 72 bytes: err = %v, want ErrPasswordTooLong", err)
	}
	// The limit counts bytes, like bcrypt: 37 x "ñ" is 37 characters but 74 bytes.
	if err := svc.Confirm(context.Background(), "t", strings.Repeat("ñ", 37)); !errors.Is(err, ErrPasswordTooLong) {
		t.Errorf("37 x ñ (74 bytes): err = %v, want ErrPasswordTooLong", err)
	}
	if repo.resetHash != "" {
		t.Error("an invalid password must not reach the repository")
	}
}
