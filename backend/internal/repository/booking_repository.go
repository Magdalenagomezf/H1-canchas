package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

// GetAllWithDetails devuelve todas las reservas con JOIN a spaces, slots y users.
// Si date viene cargada, filtra por fecha.
// Si userID viene cargado, filtra por cliente (para GET /bookings del customer).
func (r *bookingRepo) GetAllWithDetails(ctx context.Context, userID *int64, date *time.Time) ([]domain.BookingDetail, error) {
	query := `
        SELECT
            b.id,
            b.booking_date,
            b.status,
            b.total_price,
            b.created_at,
            b.updated_at,
            b.customer_user_id,
            b.customer_name,
            b.customer_phone,

            s.id        AS space_id,
            s.name      AS space_name,
            s.type      AS space_type,

            sl.id         AS slot_id,
            sl.label      AS slot_label,
            sl.start_time AS slot_start_time,
            sl.end_time   AS slot_end_time,

            u.name  AS customer_user_name,
            u.phone AS customer_user_phone

        FROM bookings b
        JOIN spaces s        ON s.id  = b.space_id
        JOIN space_slots sl  ON sl.id = b.slot_id
        LEFT JOIN users u    ON u.id  = b.customer_user_id`

	args := []interface{}{}
	conditions := []string{}

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("b.customer_user_id = $%d", len(args)+1))
		args = append(args, *userID)
	}
	if date != nil {
		conditions = append(conditions, fmt.Sprintf("b.booking_date = $%d", len(args)+1))
		args = append(args, *date)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += ` ORDER BY b.booking_date DESC, b.created_at DESC`

	var rows []domain.BookingDetail
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("bookingRepo.GetAllWithDetails: %w", err)
	}
	return rows, nil
}

// GetByIDWithDetails devuelve una reserva por ID con JOIN completo.
func (r *bookingRepo) GetByIDWithDetails(ctx context.Context, id int64) (*domain.BookingDetail, error) {
	query := `
        SELECT
            b.id,
            b.booking_date,
            b.status,
            b.total_price,
            b.created_at,
            b.updated_at,
            b.customer_user_id,
            b.customer_name,
            b.customer_phone,

            s.id        AS space_id,
            s.name      AS space_name,
            s.type      AS space_type,

            sl.id         AS slot_id,
            sl.label      AS slot_label,
            sl.start_time AS slot_start_time,
            sl.end_time   AS slot_end_time,

            u.name  AS customer_user_name,
            u.phone AS customer_user_phone

        FROM bookings b
        JOIN spaces s        ON s.id  = b.space_id
        JOIN space_slots sl  ON sl.id = b.slot_id
        LEFT JOIN users u    ON u.id  = b.customer_user_id
        WHERE b.id = $1`

	var row domain.BookingDetail
	err := r.db.GetContext(ctx, &row, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("bookingRepo.GetByIDWithDetails: %w", err)
	}
	return &row, nil
}
