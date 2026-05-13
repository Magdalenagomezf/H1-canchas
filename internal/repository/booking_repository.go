package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"H1-canchas/internal/domain"

	"github.com/jmoiron/sqlx"
)

type bookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *bookingRepo {
	return &bookingRepo{db: db}
}

// BeginTx abre una transacción nueva.
// El service la maneja: hace commit si todo sale bien, rollback si algo falla.
func (r *bookingRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// ExistsActiveBooking verifica si ya existe una reserva activa
// para el mismo space + slot + fecha.
// Una reserva activa es pending o confirmed.
// Las reservas cancelled/completed no bloquean disponibilidad.
func (r *bookingRepo) ExistsActiveBooking(ctx context.Context, spaceID, slotID int64, date time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE space_id = $1
			AND slot_id = $2
			AND booking_date = $3
			AND status IN ('pending', 'confirmed')
		)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, spaceID, slotID, date)
	if err != nil {
		return false, fmt.Errorf("bookingRepo.ExistsActiveBooking: %w", err)
	}

	return exists, nil
}

// SlotBelongsToSpace verifica que el slot exista, esté activo
// y pertenezca al space indicado.
// Esto es importante porque bookings tiene FK separadas:
// una a spaces y otra a space_slots, pero eso no asegura por sí solo
// que el slot corresponda a ese space.
func (r *bookingRepo) SlotBelongsToSpace(ctx context.Context, spaceID, slotID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM space_slots
			WHERE id = $1
			AND space_id = $2
			AND is_active = true
		)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, slotID, spaceID)
	if err != nil {
		return false, fmt.Errorf("bookingRepo.SlotBelongsToSpace: %w", err)
	}

	return exists, nil
}

// Create inserta una reserva nueva dentro de una transacción.
// Se espera que el service haya validado antes:
// - que el space exista
// - que el slot pertenezca al space
// - que aparentemente esté disponible
//
// De todos modos, PostgreSQL sigue siendo la defensa final
// gracias al índice unique_active_booking.
func (r *bookingRepo) Create(ctx context.Context, tx *sqlx.Tx, b *domain.Booking) (int64, error) {
	query := `
		INSERT INTO bookings (
			customer_user_id,
			created_by,
			customer_name,
			customer_phone,
			space_id,
			slot_id,
			booking_date,
			status,
			total_price
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING id`

	var id int64
	err := tx.QueryRowContext(
		ctx,
		query,
		b.CustomerUserID,
		b.CreatedBy,
		b.CustomerName,
		b.CustomerPhone,
		b.SpaceID,
		b.SlotID,
		b.BookingDate,
		b.Status,
		b.TotalPrice,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("bookingRepo.Create: %w", err)
	}

	return id, nil
}

// GetByID busca una reserva por ID.
// Si no existe, devuelve nil, nil.
// El service decide si eso significa ErrNotFound.
func (r *bookingRepo) GetByID(ctx context.Context, id int64) (*domain.Booking, error) {
	query := `
		SELECT
			id,
			customer_user_id,
			created_by,
			customer_name,
			customer_phone,
			space_id,
			slot_id,
			booking_date,
			status,
			total_price,
			created_at,
			updated_at
		FROM bookings
		WHERE id = $1`

	var booking domain.Booking
	err := r.db.GetContext(ctx, &booking, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("bookingRepo.GetByID: %w", err)
	}

	return &booking, nil
}

// GetByUserID devuelve todas las reservas asociadas a un usuario cliente.
// Sirve para "mis reservas".
func (r *bookingRepo) GetByUserID(ctx context.Context, userID int64) ([]domain.Booking, error) {
	query := `
		SELECT
			id,
			customer_user_id,
			created_by,
			customer_name,
			customer_phone,
			space_id,
			slot_id,
			booking_date,
			status,
			total_price,
			created_at,
			updated_at
		FROM bookings
		WHERE customer_user_id = $1
		ORDER BY booking_date DESC, created_at DESC`

	var bookings []domain.Booking
	err := r.db.SelectContext(ctx, &bookings, query, userID)
	if err != nil {
		return nil, fmt.Errorf("bookingRepo.GetByUserID: %w", err)
	}

	return bookings, nil
}

// GetAll devuelve todas las reservas.
// Si date viene cargada, filtra por fecha.
// Este método debería usarse solo desde endpoints de admin/recepcionista.
func (r *bookingRepo) GetAll(ctx context.Context, date *time.Time) ([]domain.Booking, error) {
	query := `
		SELECT
			id,
			customer_user_id,
			created_by,
			customer_name,
			customer_phone,
			space_id,
			slot_id,
			booking_date,
			status,
			total_price,
			created_at,
			updated_at
		FROM bookings`

	args := []interface{}{}

	if date != nil {
		query += `
		WHERE booking_date = $1`
		args = append(args, *date)
	}

	query += `
		ORDER BY booking_date DESC, created_at DESC`

	var bookings []domain.Booking
	err := r.db.SelectContext(ctx, &bookings, query, args...)
	if err != nil {
		return nil, fmt.Errorf("bookingRepo.GetAll: %w", err)
	}

	return bookings, nil
}

// UpdateStatus cambia parcialmente una reserva.
// Por ejemplo: pending/confirmed -> cancelled.
// Devuelve true si realmente modificó una fila.
func (r *bookingRepo) UpdateStatus(ctx context.Context, id int64, status string) (bool, error) {
	query := `
		UPDATE bookings
		SET status = $1
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return false, fmt.Errorf("bookingRepo.UpdateStatus: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("bookingRepo.UpdateStatus rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}
