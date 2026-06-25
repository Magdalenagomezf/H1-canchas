package repository_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"H1-canchas/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var slotCounter int64

var testDB *sqlx.DB

const (
	adminDSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	testDSN  = "postgres://postgres:postgres@localhost:5432/h1_canchas_test?sslmode=disable"
	testDBName = "h1_canchas_test"
)

func TestMain(m *testing.M) {
	// Connect to admin DB to create the test DB.
	admin, err := sqlx.Connect("postgres", adminDSN)
	if err != nil {
		fmt.Printf("skipping integration tests: postgres not available (%v)\n", err)
		os.Exit(0)
	}

	admin.MustExec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	admin.MustExec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	admin.Close()

	// Connect to the test DB and apply migrations.
	testDB, err = sqlx.Connect("postgres", testDSN)
	if err != nil {
		fmt.Printf("skipping integration tests: could not connect to test DB (%v)\n", err)
		os.Exit(0)
	}
	defer testDB.Close()

	if err := applyMigrations(testDB); err != nil {
		fmt.Printf("migration failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func applyMigrations(db *sqlx.DB) error {
	_, file, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")

	files := []string{
		"001_create_users.sql",
		"002_create_spaces.sql",
		"003_create_space_slots.sql",
		"004_create_bookings.sql",
	}

	for _, f := range files {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("reading %s: %w", f, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return fmt.Errorf("applying %s: %w", f, err)
		}
	}
	return nil
}

// truncateAll wipes all tables before each test.
func truncateAll(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE users, spaces, space_slots, bookings CASCADE")
	if err != nil {
		t.Fatalf("truncateAll: %v", err)
	}
}

// --- seed helpers ---

func seedUser(t *testing.T) int64 {
	t.Helper()
	var id int64
	err := testDB.QueryRowContext(context.Background(),
		`INSERT INTO users (name, phone, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`,
		"Test User", "1122334455", "hash", domain.RoleCustomer,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return id
}

func seedSpace(t *testing.T) int64 {
	t.Helper()
	var id int64
	err := testDB.QueryRowContext(context.Background(),
		`INSERT INTO spaces (name, type, price_per_slot) VALUES ($1, $2, $3) RETURNING id`,
		"Cancha Test", domain.SpaceTypePadel, 1000.0,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedSpace: %v", err)
	}
	return id
}

func seedSlot(t *testing.T, spaceID int64) int64 {
	t.Helper()
	n := atomic.AddInt64(&slotCounter, 1)
	label := fmt.Sprintf("slot-%d", n)
	var id int64
	err := testDB.QueryRowContext(context.Background(),
		`INSERT INTO space_slots (space_id, label) VALUES ($1, $2) RETURNING id`,
		spaceID, label,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedSlot: %v", err)
	}
	return id
}

func tomorrow() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, 1)
}
