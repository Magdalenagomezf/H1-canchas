// Package google verifies Google Sign-In ID tokens.
package google

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/idtoken"
)

// Identity is the subset of a verified Google ID token that callers use.
type Identity struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
}

// Verifier validates Google ID tokens with Google's public keys.
type Verifier struct {
	clientID string
}

// NewVerifier returns a Verifier that only accepts tokens issued for clientID.
func NewVerifier(clientID string) *Verifier {
	return &Verifier{clientID: clientID}
}

// Verify checks signature, audience and expiry (idtoken.Validate) plus the
// issuer, and returns the identity inside the token.
func (v *Verifier) Verify(ctx context.Context, credential string) (*Identity, error) {
	payload, err := idtoken.Validate(ctx, credential, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("google: validating id token: %w", err)
	}
	return identityFromClaims(payload.Issuer, payload.Subject, payload.Claims)
}

// identityFromClaims is split from Verify so the issuer and claim handling can
// be tested without calling Google.
func identityFromClaims(issuer, subject string, claims map[string]any) (*Identity, error) {
	if issuer != "accounts.google.com" && issuer != "https://accounts.google.com" {
		return nil, fmt.Errorf("google: unexpected issuer %q", issuer)
	}
	if subject == "" {
		return nil, errors.New("google: token without subject")
	}

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	verified, _ := claims["email_verified"].(bool)

	return &Identity{
		Sub:           subject,
		Email:         email,
		EmailVerified: verified,
		Name:          name,
	}, nil
}
