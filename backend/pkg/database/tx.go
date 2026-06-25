package database

import (
	"context"
	"database/sql"
)

// Tx is the minimal transaction interface used across the service and repository layers.
// *sqlx.Tx satisfies this automatically, so no casting or wrapping is needed.
type Tx interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
}
