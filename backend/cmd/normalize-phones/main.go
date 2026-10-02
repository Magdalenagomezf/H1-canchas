// Command normalize-phones rewrites users.phone and bookings.customer_phone
// to the canonical E.164 form used by pkg/phone.
//
// By default it only prints what would change (dry run). Pass --apply to
// write the changes in a single transaction.
//
//	go run ./cmd/normalize-phones           # dry run
//	go run ./cmd/normalize-phones --apply   # apply
//
// Rows whose phone cannot be normalized are reported and left untouched.
// Users whose normalized phone would collide with another user are reported
// and skipped, never applied: resolve them by hand and run the command again.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"H1-canchas/config"
	"H1-canchas/pkg/database"
	"H1-canchas/pkg/phone"

	"github.com/jmoiron/sqlx"
)

// maintenancePlaceholder is what maintenance blocks store in customer_phone
// (see BookingBatchService.CreateMaintenanceBatch). It is not a real phone.
const maintenancePlaceholder = "-"

type row struct {
	ID    int64  `db:"id"`
	Phone string `db:"phone"`
}

type change struct {
	ID       int64
	Old, New string
}

// plan is the result of comparing stored phones with their normalized form.
type plan struct {
	changes []change // rows to update
	invalid []row    // rows that fail to normalize, left untouched
}

func main() {
	apply := flag.Bool("apply", false, "write the changes (default is a dry run)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.New(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := run(context.Background(), db, *apply); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, db *sqlx.DB, apply bool) error {
	var users, bookings []row
	if err := db.SelectContext(ctx, &users, `SELECT id, phone FROM users ORDER BY id`); err != nil {
		return fmt.Errorf("loading users: %w", err)
	}
	if err := db.SelectContext(ctx, &bookings,
		`SELECT id, customer_phone AS phone FROM bookings WHERE customer_phone IS NOT NULL ORDER BY id`); err != nil {
		return fmt.Errorf("loading bookings: %w", err)
	}

	userPlan := buildPlan(users)
	userChanges, collisions := splitCollisions(users, userPlan.changes)
	bookingPlan := buildPlan(bookings)

	report("users.phone", userChanges, userPlan.invalid)
	for _, c := range collisions {
		fmt.Printf("  SKIPPED (collides with another user): user %d  %q -> %q\n", c.ID, c.Old, c.New)
	}
	report("bookings.customer_phone", bookingPlan.changes, bookingPlan.invalid)

	if !apply {
		fmt.Println("\nDry run: nothing was written. Re-run with --apply to write these changes.")
		return nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, c := range userChanges {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET phone = $1 WHERE id = $2`, c.New, c.ID); err != nil {
			return fmt.Errorf("updating user %d: %w", c.ID, err)
		}
	}
	for _, c := range bookingPlan.changes {
		if _, err := tx.ExecContext(ctx, `UPDATE bookings SET customer_phone = $1 WHERE id = $2`, c.New, c.ID); err != nil {
			return fmt.Errorf("updating booking %d: %w", c.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("\nApplied: %d user(s), %d booking(s) updated.\n", len(userChanges), len(bookingPlan.changes))
	return nil
}

// buildPlan normalizes every row and splits the result into changes and
// rows that cannot be normalized. Already-canonical rows are ignored.
func buildPlan(rows []row) plan {
	var p plan
	for _, r := range rows {
		if r.Phone == maintenancePlaceholder {
			continue
		}
		normalized, err := phone.Normalize(r.Phone)
		if err != nil {
			p.invalid = append(p.invalid, r)
			continue
		}
		if normalized != r.Phone {
			p.changes = append(p.changes, change{ID: r.ID, Old: r.Phone, New: normalized})
		}
	}
	return p
}

// splitCollisions separates the user changes that are safe to apply from the
// ones whose new phone would be held by more than one user afterwards
// (users.phone is UNIQUE).
func splitCollisions(users []row, changes []change) (safe, collisions []change) {
	target := make(map[int64]string, len(users)) // phone each user ends up with
	for _, u := range users {
		target[u.ID] = u.Phone
	}
	for _, c := range changes {
		target[c.ID] = c.New
	}

	holders := make(map[string]int)
	for _, p := range target {
		holders[p]++
	}

	for _, c := range changes {
		if holders[c.New] > 1 {
			collisions = append(collisions, c)
		} else {
			safe = append(safe, c)
		}
	}
	return safe, collisions
}

func report(label string, changes []change, invalid []row) {
	fmt.Printf("\n== %s: %d to change, %d invalid ==\n", label, len(changes), len(invalid))
	for _, c := range changes {
		fmt.Printf("  id %d  %q -> %q\n", c.ID, c.Old, c.New)
	}
	for _, r := range invalid {
		fmt.Printf("  INVALID (left untouched): id %d  %q\n", r.ID, r.Phone)
	}
}
