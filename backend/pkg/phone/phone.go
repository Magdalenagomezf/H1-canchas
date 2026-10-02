// Package phone normalizes phone numbers to a single canonical form.
package phone

import (
	"errors"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// ErrInvalid is returned when the input is not a valid phone number.
var ErrInvalid = errors.New("phone: invalid number")

const (
	defaultRegion = "AR"
	arCountryCode = 54
)

// Normalize parses raw as a phone number (default region AR) and returns it
// in E.164 format, e.g. "+5493834123456".
//
// Argentine numbers need extra care: phonenumbers strips the local "0" and
// "15" mobile prefix, but it does NOT add the mobile "9" for numbers typed in
// local format ("3834123456" gives +543834123456, "0383 15-412-3456" gives
// +5493834123456). Since users type the same number in any of those ways, we
// always store Argentine numbers with the "9" (+549...), the form WhatsApp and
// mobile carriers use, so that every spelling maps to one string.
func Normalize(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalid
	}

	num, err := phonenumbers.Parse(raw, defaultRegion)
	if err != nil {
		return "", ErrInvalid
	}

	if num.GetCountryCode() == arCountryCode {
		national := phonenumbers.GetNationalSignificantNumber(num)
		if !strings.HasPrefix(national, "9") {
			num, err = phonenumbers.Parse("+549"+national, defaultRegion)
			if err != nil {
				return "", ErrInvalid
			}
		}
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", ErrInvalid
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
