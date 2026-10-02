package service

import (
	"net/mail"
	"strings"
)

const maxEmailLength = 150 // users.email VARCHAR(150)

// normalizeEmail trims and lowercases raw. An empty input returns (nil, nil)
// so it is stored as NULL, never as "". A non-empty input must be a bare
// address: "Name <a@b.com>" is rejected even though net/mail can parse it.
func normalizeEmail(raw string) (*string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return nil, nil
	}
	if len(email) > maxEmailLength {
		return nil, ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return nil, ErrInvalidEmail
	}
	return &email, nil
}
