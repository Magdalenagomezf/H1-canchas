package google

import "testing"

func TestIdentityFromClaims_Issuer(t *testing.T) {
	claims := map[string]any{"email": "a@b.com", "email_verified": true, "name": "Ana"}

	tests := []struct {
		name    string
		issuer  string
		wantErr bool
	}{
		{"bare issuer", "accounts.google.com", false},
		{"https issuer", "https://accounts.google.com", false},
		{"other issuer", "https://evil.example.com", true},
		{"empty issuer", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := identityFromClaims(tt.issuer, "sub-1", claims)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (got.Sub != "sub-1" || got.Email != "a@b.com" || !got.EmailVerified || got.Name != "Ana") {
				t.Errorf("unexpected identity: %+v", got)
			}
		})
	}
}

func TestIdentityFromClaims_MissingClaims(t *testing.T) {
	got, err := identityFromClaims("accounts.google.com", "sub-1", map[string]any{"email_verified": "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EmailVerified {
		t.Error("a non-bool email_verified must not count as verified")
	}
	if got.Email != "" || got.Name != "" {
		t.Errorf("expected empty email and name, got %+v", got)
	}

	if _, err := identityFromClaims("accounts.google.com", "", nil); err == nil {
		t.Error("expected error for empty subject")
	}
}
