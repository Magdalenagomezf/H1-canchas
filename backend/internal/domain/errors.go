package domain

import "errors"

// Repositories return these sentinels when an INSERT hits a unique index,
// so services can map concurrent duplicates to their own domain errors.
var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrDuplicatePhone = errors.New("duplicate phone")
)
