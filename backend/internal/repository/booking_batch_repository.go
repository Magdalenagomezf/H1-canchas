package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"H1-canchas/internal/domain"

	"github.com/jmoiron/sqlx"
)

type bookingBatchRepo struct {
	db *sqlx.DB
}

func NewBookingBatchRepository(db *sqlx.DB) *bookingBatchRepo {
	return &bookingBatchRepo{db: db}
}

// Create inserta la "receta" del batch (turno fijo o bloqueo de mantenimiento).
// Los bookings individuales se insertan aparte, uno por fecha, y se etiquetan
// con este ID vía bookingRepo.SetBatchID.
func (r *bookingBatchRepo) Create(ctx context.Context, b *domain.BookingBatch) (int64, error) {
	query := `
		INSERT INTO booking_batches (
			type, space_id, slot_id, weekday, start_date, end_date, reason, created_by, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	var id int64
	err := r.db.QueryRowContext(
		ctx, query,
		b.Type, b.SpaceID, b.SlotID, b.Weekday, b.StartDate, b.EndDate, b.Reason, b.CreatedBy, b.Status,
	).Scan(&id, &b.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("bookingBatchRepo.Create: %w", err)
	}
	return id, nil
}

// GetByID busca un batch por ID. Si no existe, devuelve nil, nil.
func (r *bookingBatchRepo) GetByID(ctx context.Context, id int64) (*domain.BookingBatch, error) {
	query := `
		SELECT id, type, space_id, slot_id, weekday, start_date, end_date, reason, created_by, status, created_at
		FROM booking_batches
		WHERE id = $1`

	var batch domain.BookingBatch
	err := r.db.GetContext(ctx, &batch, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("bookingBatchRepo.GetByID: %w", err)
	}
	return &batch, nil
}

// GetAll devuelve los batches, opcionalmente filtrados por tipo,
// más recientes primero.
func (r *bookingBatchRepo) GetAll(ctx context.Context, batchType *string) ([]domain.BookingBatch, error) {
	query := `
		SELECT id, type, space_id, slot_id, weekday, start_date, end_date, reason, created_by, status, created_at
		FROM booking_batches`

	args := []any{}
	if batchType != nil {
		query += " WHERE type = $1"
		args = append(args, *batchType)
	}
	query += " ORDER BY created_at DESC"

	var batches []domain.BookingBatch
	if err := r.db.SelectContext(ctx, &batches, query, args...); err != nil {
		return nil, fmt.Errorf("bookingBatchRepo.GetAll: %w", err)
	}
	return batches, nil
}

// GetAllDetailed devuelve los batches con nombre de cancha y label de turno
// ya cruzados, opcionalmente filtrados por tipo, más recientes primero.
func (r *bookingBatchRepo) GetAllDetailed(ctx context.Context, batchType *string) ([]domain.BookingBatchDetail, error) {
	query := `
		SELECT
			bb.id, bb.type, bb.weekday, bb.start_date, bb.end_date,
			bb.reason, bb.status, bb.created_at,
			s.id AS space_id, s.name AS space_name, s.type AS space_type,
			bb.slot_id, sl.label AS slot_label
		FROM booking_batches bb
		JOIN spaces s            ON s.id  = bb.space_id
		LEFT JOIN space_slots sl ON sl.id = bb.slot_id`

	args := []any{}
	if batchType != nil {
		query += " WHERE bb.type = $1"
		args = append(args, *batchType)
	}
	query += " ORDER BY bb.created_at DESC"

	var batches []domain.BookingBatchDetail
	if err := r.db.SelectContext(ctx, &batches, query, args...); err != nil {
		return nil, fmt.Errorf("bookingBatchRepo.GetAllDetailed: %w", err)
	}
	return batches, nil
}

// UpdateStatus cambia el estado del batch (ej. a 'cancelled').
func (r *bookingBatchRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE booking_batches SET status = $1 WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, query, status, id); err != nil {
		return fmt.Errorf("bookingBatchRepo.UpdateStatus: %w", err)
	}
	return nil
}
