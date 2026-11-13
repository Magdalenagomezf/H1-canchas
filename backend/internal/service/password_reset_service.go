package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	htmltemplate "html/template"
	"log"
	"strings"
	texttemplate "text/template"
	"time"

	"H1-canchas/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordResetTTL      = time.Hour
	passwordResetCooldown = 2 * time.Minute
	passwordResetSubject  = "Restablecé tu contraseña — H1 Canchas"
)

//go:embed templates/password_reset.html
var passwordResetHTML string

//go:embed templates/password_reset.txt
var passwordResetText string

var (
	passwordResetHTMLTmpl = htmltemplate.Must(htmltemplate.New("password_reset.html").Parse(passwordResetHTML))
	passwordResetTextTmpl = texttemplate.Must(texttemplate.New("password_reset.txt").Parse(passwordResetText))
)

// PasswordResetUserRepo is what password reset needs from users.
type PasswordResetUserRepo interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

// PasswordResetRepo stores the hashed one-time tokens.
type PasswordResetRepo interface {
	CreateIfNoRecent(ctx context.Context, userID int64, tokenHash string, ttl, cooldown time.Duration) (bool, error)
	IsRedeemable(ctx context.Context, tokenHash string) (bool, error)
	ResetPassword(ctx context.Context, tokenHash, passwordHash string) (bool, error)
}

type PasswordResetService struct {
	users       PasswordResetUserRepo
	tokens      PasswordResetRepo
	sender      EmailSender
	frontendURL string
}

func NewPasswordResetService(users PasswordResetUserRepo, tokens PasswordResetRepo, sender EmailSender, frontendURL string) *PasswordResetService {
	return &PasswordResetService{
		users:       users,
		tokens:      tokens,
		sender:      sender,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

type passwordResetData struct {
	Name     string
	ResetURL string
}

// RequestReset starts the reset in the background and returns at once, so
// the caller answers the same way, and just as fast, whether the account
// exists or not. It uses its own context: the request one is cancelled as
// soon as the handler responds.
func (s *PasswordResetService) RequestReset(email string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.SendResetLink(ctx, email); err != nil {
			// Never log the token or the email: the error only carries the cause.
			log.Printf("WARN password reset: %v", err)
		}
	}()
}

// SendResetLink emails a one-time reset link to the active account with
// that email. It returns nil without sending anything when there is no such
// account or a link was sent less than the cooldown ago. Synchronous so it
// can be tested.
func (s *PasswordResetService) SendResetLink(ctx context.Context, rawEmail string) error {
	email, err := normalizeOptionalEmail(&rawEmail)
	if err != nil || email == nil {
		return nil // malformed or empty: no account can match
	}

	user, err := s.users.FindByEmail(ctx, *email)
	if err != nil {
		return fmt.Errorf("PasswordResetService.SendResetLink: %w", err)
	}
	if user == nil || !user.IsActive || user.Email == nil {
		return nil
	}

	token, err := newResetToken()
	if err != nil {
		return fmt.Errorf("PasswordResetService.SendResetLink: %w", err)
	}
	created, err := s.tokens.CreateIfNoRecent(ctx, user.ID, hashResetToken(token), passwordResetTTL, passwordResetCooldown)
	if err != nil {
		return fmt.Errorf("PasswordResetService.SendResetLink: %w", err)
	}
	if !created {
		return nil // cooldown: the previous link is still in the user's inbox
	}

	data := passwordResetData{
		Name:     strings.TrimSpace(user.Name),
		ResetURL: s.frontendURL + "/restablecer-contrasena#token=" + token,
	}
	var htmlBuf, textBuf bytes.Buffer
	if err := passwordResetHTMLTmpl.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("PasswordResetService.SendResetLink: render html: %w", err)
	}
	if err := passwordResetTextTmpl.Execute(&textBuf, data); err != nil {
		return fmt.Errorf("PasswordResetService.SendResetLink: render text: %w", err)
	}

	if err := s.sender.Send(ctx, *user.Email, passwordResetSubject, htmlBuf.String(), textBuf.String()); err != nil {
		return fmt.Errorf("PasswordResetService.SendResetLink: sending mail to user %d: %w", user.ID, err)
	}
	return nil
}

// CheckToken tells the reset page, as soon as it opens, whether its link still
// works, so the user does not fill in the form for nothing. It never uses the
// token up; Confirm checks again when redeeming.
func (s *PasswordResetService) CheckToken(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidResetToken
	}
	ok, err := s.tokens.IsRedeemable(ctx, hashResetToken(token))
	if err != nil {
		return fmt.Errorf("PasswordResetService.CheckToken: %w", err)
	}
	if !ok {
		return ErrInvalidResetToken
	}
	return nil
}

// Confirm sets a new password with a reset token. Any token that cannot be
// redeemed (unknown, used, expired, inactive user) gives ErrInvalidResetToken.
func (s *PasswordResetService) Confirm(ctx context.Context, token, password string) error {
	if err := validateNewPassword(password); err != nil {
		return err
	}
	if token == "" {
		return ErrInvalidResetToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("PasswordResetService.Confirm: hashing password: %w", err)
	}
	ok, err := s.tokens.ResetPassword(ctx, hashResetToken(token), string(hash))
	if err != nil {
		return fmt.Errorf("PasswordResetService.Confirm: %w", err)
	}
	if !ok {
		return ErrInvalidResetToken
	}
	return nil
}

// newResetToken returns 32 random bytes from crypto/rand as base64url.
func newResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating reset token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashResetToken is what the database stores and looks up. Looking up by a
// hash means an attacker cannot learn the token byte by byte from timing.
func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
