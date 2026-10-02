package service

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeEmail_MaxLength(t *testing.T) {
	const suffix = "@example.com"
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{"exactly max length is accepted", strings.Repeat("a", maxEmailLength-len(suffix)) + suffix, nil},
		{"one over max length is rejected", strings.Repeat("a", maxEmailLength-len(suffix)+1) + suffix, ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEmail(tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (got == nil || len(*got) != maxEmailLength) {
				t.Errorf("expected a %d-char email, got %v", maxEmailLength, got)
			}
		})
	}
}
