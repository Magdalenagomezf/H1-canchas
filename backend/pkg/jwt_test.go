package jwtutil

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

// --- GenerateToken ---

func TestGenerateToken_EmptySecret(t *testing.T) {
	_, err := GenerateToken(1, "customer", "")
	if err == nil {
		t.Fatal("expected error for empty secret, got nil")
	}
}

func TestGenerateToken_HappyPath(t *testing.T) {
	token, err := GenerateToken(1, "customer", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateToken_ClaimsRoundtrip(t *testing.T) {
	const userID int64 = 42
	const role = "admin"

	tokenStr, err := GenerateToken(userID, role, testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("got UserID %d, want %d", claims.UserID, userID)
	}
	if claims.Role != role {
		t.Errorf("got Role %q, want %q", claims.Role, role)
	}
}

// --- ParseToken ---

func TestParseToken_EmptySecret(t *testing.T) {
	tokenStr, _ := GenerateToken(1, "customer", testSecret)
	_, err := ParseToken(tokenStr, "")
	if err == nil {
		t.Fatal("expected error for empty secret, got nil")
	}
}

func TestParseToken_InvalidTokenString(t *testing.T) {
	_, err := ParseToken("not.a.token", testSecret)
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	tokenStr, _ := GenerateToken(1, "customer", testSecret)
	_, err := ParseToken(tokenStr, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	claims := &Claims{
		UserID: 1,
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	claims := &Claims{
		UserID: 1,
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for wrong signing method, got nil")
	}
}

func TestParseToken_InvalidClaims_ZeroUserID(t *testing.T) {
	claims := &Claims{
		UserID: 0,
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for zero UserID, got nil")
	}
}

func TestParseToken_InvalidClaims_EmptyRole(t *testing.T) {
	claims := &Claims{
		UserID: 1,
		Role:   "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for empty Role, got nil")
	}
}

func TestParseToken_HappyPath(t *testing.T) {
	const userID int64 = 7
	const role = "receptionist"

	tokenStr, _ := GenerateToken(userID, role, testSecret)
	claims, err := ParseToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("got UserID %d, want %d", claims.UserID, userID)
	}
	if claims.Role != role {
		t.Errorf("got Role %q, want %q", claims.Role, role)
	}
	if claims.ExpiresAt == nil {
		t.Error("expected non-nil ExpiresAt")
	}
}
