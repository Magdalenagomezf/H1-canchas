package phone

import (
	"errors"
	"testing"
)

func TestNormalize_Valid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"local format without 9", "3834123456", "+5493834123456"},
		{"trunk 0 and 15 prefix", "0383 15-412-3456", "+5493834123456"},
		{"international with 9", "+54 9 383 412 3456", "+5493834123456"},
		{"international without 9", "+54 383 412 3456", "+5493834123456"},
		{"already E.164", "+5493834123456", "+5493834123456"},
		{"surrounding spaces", "  3834123456  ", "+5493834123456"},
		{"buenos aires area code", "11 2233-4455", "+5491122334455"},
		{"buenos aires with 15", "011 15 2233-4455", "+5491122334455"},
		{"foreign number is kept", "+598 99 123 456", "+59899123456"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.in)
			if err != nil {
				t.Fatalf("Normalize(%q): unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalize_Invalid(t *testing.T) {
	for _, in := range []string{"", "   ", "abc", "-", "123", "0000000000", "+54 9 383 12"} {
		t.Run(in, func(t *testing.T) {
			got, err := Normalize(in)
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("Normalize(%q) = %q, %v; want ErrInvalid", in, got, err)
			}
		})
	}
}
