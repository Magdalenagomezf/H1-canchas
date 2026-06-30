package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"H1-canchas/internal/domain"

	"github.com/jmoiron/sqlx"
)

type spaceRepo struct {
	db *sqlx.DB
}

func NewSpaceRepository(db *sqlx.DB) *spaceRepo {
	return &spaceRepo{db: db}
}

// crear cancha o salon
func (r *spaceRepo) Create(ctx context.Context, space *domain.Space) (int64, error) {
	query := `
		INSERT INTO spaces (name, type, description, price_per_slot)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		space.Name,
		space.Type,
		space.Description,
		space.PricePerSlot,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("spaceRepo.Create: %w", err)
	}

	return id, nil
}

// devuelve todos los espacios activos
func (r *spaceRepo) GetAll(ctx context.Context) ([]domain.Space, error) {
	query := `
        SELECT id, name, type, description, price_per_slot, is_active, created_at
        FROM spaces
        WHERE is_active = true
        ORDER BY type, name`

	var spaces []domain.Space
	err := r.db.SelectContext(ctx, &spaces, query)
	if err != nil {
		return nil, fmt.Errorf("spaceRepo.GetAll: %w", err)
	}
	return spaces, nil
}

// devuelve un espacio por su ID
func (r *spaceRepo) GetByID(ctx context.Context, id int64) (*domain.Space, error) {
	query := `
        SELECT id, name, type, description, price_per_slot, is_active, created_at
        FROM spaces
        WHERE id = $1`

	var space domain.Space
	err := r.db.GetContext(ctx, &space, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("spaceRepo.GetByID: %w", err)
	}
	return &space, nil
}

// devuelve los turnos disponibles para un espacio
func (r *spaceRepo) GetSlotsBySpaceID(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error) {
	query := `
        SELECT id, space_id, label, description, start_time, end_time, is_active
        FROM space_slots
        WHERE space_id = $1
        AND is_active = true
        ORDER BY start_time`

	var slots []domain.SpaceSlot
	err := r.db.SelectContext(ctx, &slots, query, spaceID)
	if err != nil {
		return nil, fmt.Errorf("spaceRepo.GetSlotsBySpaceID: %w", err)
	}
	return slots, nil
}

func (r *spaceRepo) GetAvailableSlotsForDate(ctx context.Context, spaceID int64, date string) ([]domain.SpaceSlot, error) {
	query := `
        SELECT ss.id, ss.space_id, ss.label, ss.description, ss.start_time, ss.end_time, ss.is_active
        FROM space_slots ss
        WHERE ss.space_id = $1
          AND ss.is_active = true
          AND NOT EXISTS (
              SELECT 1 FROM bookings b
              WHERE b.space_id = ss.space_id
                AND b.slot_id = ss.id
                AND b.booking_date = $2::date
                AND b.status IN ('pending', 'confirmed')
          )
        ORDER BY ss.start_time`

	var slots []domain.SpaceSlot
	err := r.db.SelectContext(ctx, &slots, query, spaceID, date)
	if err != nil {
		return nil, fmt.Errorf("spaceRepo.GetAvailableSlotsForDate: %w", err)
	}
	return slots, nil
}

func (r *spaceRepo) CreateSlot(ctx context.Context, slot *domain.SpaceSlot) (int64, error) {
	query := `
		INSERT INTO space_slots (space_id, label, description, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		slot.SpaceID,
		slot.Label,
		slot.Description,
		slot.StartTime,
		slot.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("spaceRepo.CreateSlot: %w", err)
	}
	return id, nil
}

func (r *spaceRepo) Update(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error {
	query := `
		UPDATE spaces
		SET name = $1, description = $2, price_per_slot = $3
		WHERE id = $4 AND is_active = true`

	_, err := r.db.ExecContext(ctx, query, name, description, pricePerSlot, id)
	if err != nil {
		return fmt.Errorf("spaceRepo.Update: %w", err)
	}
	return nil
}

func (r *spaceRepo) HasBookings(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE space_id = $1 LIMIT 1)`, id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("spaceRepo.HasBookings: %w", err)
	}
	return exists, nil
}

func (r *spaceRepo) HardDelete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("spaceRepo.HardDelete: begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM space_slots WHERE space_id = $1`, id); err != nil {
		return fmt.Errorf("spaceRepo.HardDelete: delete slots: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM spaces WHERE id = $1`, id); err != nil {
		return fmt.Errorf("spaceRepo.HardDelete: delete space: %w", err)
	}

	return tx.Commit()
}

// desactiva un espacio
func (r *spaceRepo) Deactivate(ctx context.Context, id int64) error {
	query := `
		UPDATE spaces
		SET is_active = false
		WHERE id = $1
		AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("spaceRepo.Deactivate: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("spaceRepo.Deactivate rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
