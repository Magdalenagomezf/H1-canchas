package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"H1-canchas/internal/domain"
	jwtutil "H1-canchas/pkg"
	"H1-canchas/pkg/phone"
)

const maxNameLength = 100 // users.name VARCHAR(100)

// GoogleIdentity is the subset of a verified Google ID token the service uses.
type GoogleIdentity struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
}

// GoogleTokenVerifier validates a Google ID token ("credential"). It is
// defined here so tests never call Google.
type GoogleTokenVerifier interface {
	Verify(ctx context.Context, credential string) (*GoogleIdentity, error)
}

// truncateRunes cuts s to at most max characters (not bytes).
func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}

// googleDisplayName returns the name for a new Google account: Google's name
// trimmed to 100 characters, or the part of the email before the "@".
func googleDisplayName(name, email string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}
	return truncateRunes(name, maxNameLength)
}

// LoginWithGoogle signs a user in (creating or linking the account when
// needed) from a Google ID token. It returns the same values as Login.
func (s *UserService) LoginWithGoogle(ctx context.Context, credential string) (*domain.User, string, error) {
	if s.google == nil {
		return nil, "", ErrGoogleNotConfigured
	}

	identity, err := s.google.Verify(ctx, credential)
	if err != nil {
		// Keep the technical cause (audience, expiry, signature...) for the logs;
		// the controller only shows ErrInvalidGoogleToken's message.
		return nil, "", fmt.Errorf("%w: %v", ErrInvalidGoogleToken, err)
	}
	if identity == nil || identity.Sub == "" {
		return nil, "", fmt.Errorf("%w: verifier returned no identity", ErrInvalidGoogleToken)
	}
	if !identity.EmailVerified {
		return nil, "", ErrGoogleEmailNotVerified
	}

	email, err := normalizeEmail(identity.Email)
	if err != nil || email == nil {
		return nil, "", ErrInvalidGoogleToken
	}

	user, err := s.repo.FindByGoogleSub(ctx, identity.Sub)
	if err != nil {
		return nil, "", fmt.Errorf("LoginWithGoogle: %w", err)
	}

	if user == nil {
		user, err = s.linkOrCreateGoogleUser(ctx, identity, *email)
		if err != nil {
			return nil, "", err
		}
	}

	if !user.IsActive {
		return nil, "", ErrUserInactive
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("LoginWithGoogle: generando token: %w", err)
	}
	return user, token, nil
}

// linkOrCreateGoogleUser handles a Google identity whose sub is not known yet:
// it links an existing customer with the same email, or creates a new one.
func (s *UserService) linkOrCreateGoogleUser(ctx context.Context, identity *GoogleIdentity, email string) (*domain.User, error) {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("LoginWithGoogle: %w", err)
	}

	if existing != nil {
		if existing.Role != domain.RoleCustomer {
			return nil, ErrGoogleStaffAccount
		}
		if existing.GoogleSub != nil {
			return nil, ErrGoogleAccountConflict
		}
		if err := s.repo.SetGoogleSub(ctx, existing.ID, identity.Sub); err != nil {
			if errors.Is(err, domain.ErrDuplicateGoogleSub) {
				return nil, ErrGoogleAccountConflict
			}
			return nil, fmt.Errorf("LoginWithGoogle: vinculando cuenta: %w", err)
		}
		sub := identity.Sub
		existing.GoogleSub = &sub
		return existing, nil
	}

	sub := identity.Sub
	user := &domain.User{
		Name:      googleDisplayName(identity.Name, email),
		Email:     &email,
		GoogleSub: &sub,
		Role:      domain.RoleCustomer,
		IsActive:  true,
	}
	id, err := s.repo.Create(ctx, user)
	if err != nil {
		// A concurrent request created the same account; no retry.
		if errors.Is(err, domain.ErrDuplicateEmail) || errors.Is(err, domain.ErrDuplicateGoogleSub) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("LoginWithGoogle: creando usuario: %w", err)
	}
	user.ID = id
	return user, nil
}

// GetProfile returns the user behind the token.
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetProfile: %w", err)
	}
	if user == nil {
		return nil, ErrNotFound
	}
	return user, nil
}

// UpdateProfile changes name, email and/or phone of userID. A nil argument
// means "leave as is". Email and phone cannot be cleared.
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, name, rawEmail, rawPhone *string) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}
	if user == nil {
		return nil, ErrNotFound
	}

	newName := user.Name
	if name != nil {
		newName = strings.TrimSpace(*name)
		if newName == "" {
			return nil, ErrNameRequired
		}
		if utf8.RuneCountInString(newName) > maxNameLength {
			return nil, ErrNameTooLong
		}
	}

	newEmail := user.Email
	if rawEmail != nil {
		cleaned, err := normalizeEmail(*rawEmail)
		if err != nil {
			return nil, err
		}
		if cleaned == nil {
			return nil, ErrEmailRequired
		}
		other, err := s.repo.FindByEmail(ctx, *cleaned)
		if err != nil {
			return nil, fmt.Errorf("UpdateProfile: %w", err)
		}
		if other != nil && other.ID != userID {
			return nil, ErrEmailAlreadyExists
		}
		newEmail = cleaned
	}

	newPhone := user.Phone
	if rawPhone != nil {
		normalized, err := phone.Normalize(*rawPhone)
		if err != nil {
			return nil, ErrInvalidPhone
		}
		other, err := s.repo.FindByPhone(ctx, normalized)
		if err != nil {
			return nil, fmt.Errorf("UpdateProfile: %w", err)
		}
		if other != nil && other.ID != userID {
			return nil, ErrPhoneAlreadyExists
		}
		newPhone = &normalized
	}

	if err := s.repo.UpdateProfile(ctx, userID, newName, newEmail, newPhone); err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			return nil, ErrEmailAlreadyExists
		}
		if errors.Is(err, domain.ErrDuplicatePhone) {
			return nil, ErrPhoneAlreadyExists
		}
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}

	updated := *user
	updated.Name = newName
	updated.Email = newEmail
	updated.Phone = newPhone
	return &updated, nil
}
