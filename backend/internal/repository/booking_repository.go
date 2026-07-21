package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/database"

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
func (r *bookingRepo) BeginTx(ctx context.Context) (database.Tx, error) {
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
func (r *bookingRepo) Create(ctx context.Context, tx database.Tx, b *domain.Booking) (int64, error) {
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
			total_price,
			deposit_amount,
			expires_at,
			balance_amount,
			batch_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
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
		b.DepositAmount,
		b.ExpiresAt,
		b.BalanceAmount,
		b.BatchID,
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
			deposit_amount,
			deposit_status,
			deposit_method,
			deposit_recorded_by,
			expires_at,
			balance_amount,
			balance_status,
			balance_method,
			balance_recorded_by,
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
            b.deposit_amount,
            b.deposit_status,
            b.deposit_method,
            b.expires_at,
            b.balance_amount,
            b.balance_status,
            b.balance_method,
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
            u.phone AS customer_user_phone,

            b.batch_id,
            bb.type   AS batch_type,
            bb.reason AS batch_reason

        FROM bookings b
        JOIN spaces s        ON s.id  = b.space_id
        JOIN space_slots sl  ON sl.id = b.slot_id
        LEFT JOIN users u    ON u.id  = b.customer_user_id
        LEFT JOIN booking_batches bb ON bb.id = b.batch_id`

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
            b.deposit_amount,
            b.deposit_status,
            b.deposit_method,
            b.expires_at,
            b.balance_amount,
            b.balance_status,
            b.balance_method,
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
            u.phone AS customer_user_phone,

            b.batch_id,
            bb.type   AS batch_type,
            bb.reason AS batch_reason

        FROM bookings b
        JOIN spaces s        ON s.id  = b.space_id
        JOIN space_slots sl  ON sl.id = b.slot_id
        LEFT JOIN users u    ON u.id  = b.customer_user_id
        LEFT JOIN booking_batches bb ON bb.id = b.batch_id
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

// ConfirmAfterPayment marca un tramo de una reserva (deposit o balance)
// como pagado tras la confirmación de un pago (webhook o marcado manual).
// Solo el tramo "deposit" afecta el ciclo de vida de status/expires_at:
// si la reserva todavía estaba pending, la pasa a confirmed y limpia el
// hold; si ya estaba en otro estado no terminal lo deja como está. El
// tramo "balance" no tiene hold ni efecto sobre status — se paga después,
// sin condicionar si la reserva está confirmada.
//
// A prueba del caso límite en el que el pago se acredita después de que
// el hold ya expiró y el slot fue tomado por otro: si la reserva está
// cancelled o expired, no hace nada (el pago queda registrado en
// payments, pero la reserva vencida no revive).
//
// database.Tx solo expone QueryRowContext (no ExecContext), así que se
// usa RETURNING id en vez de RowsAffected para detectar el "no-op".
func (r *bookingRepo) ConfirmAfterPayment(ctx context.Context, tx database.Tx, bookingID int64, kind string) (bool, error) {
	var query string
	switch kind {
	case domain.PaymentKindDeposit:
		query = `
			UPDATE bookings
			SET deposit_status = 'paid',
				expires_at = NULL,
				status = CASE WHEN status = 'pending' THEN 'confirmed' ELSE status END
			WHERE id = $1
			AND status NOT IN ('cancelled', 'expired')
			RETURNING id`
	case domain.PaymentKindBalance:
		query = `
			UPDATE bookings
			SET balance_status = 'paid'
			WHERE id = $1
			AND status NOT IN ('cancelled', 'expired')
			RETURNING id`
	default:
		return false, fmt.Errorf("bookingRepo.ConfirmAfterPayment: invalid kind %q", kind)
	}

	var id int64
	err := tx.QueryRowContext(ctx, query, bookingID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("bookingRepo.ConfirmAfterPayment: %w", err)
	}

	return true, nil
}

// MarkPaidManually marca un tramo (deposit o balance) de una reserva
// como pagado a mano (efectivo, transferencia, posnet) por un miembro
// del staff, sin pasar por Mercado Pago.
func (r *bookingRepo) MarkPaidManually(ctx context.Context, id int64, kind, method string, recordedBy int64) (bool, error) {
	var query string
	switch kind {
	case domain.PaymentKindDeposit:
		query = `
			UPDATE bookings
			SET deposit_status = 'paid',
				deposit_method = $1,
				deposit_recorded_by = $2
			WHERE id = $3`
	case domain.PaymentKindBalance:
		query = `
			UPDATE bookings
			SET balance_status = 'paid',
				balance_method = $1,
				balance_recorded_by = $2
			WHERE id = $3`
	default:
		return false, fmt.Errorf("bookingRepo.MarkPaidManually: invalid kind %q", kind)
	}

	result, err := r.db.ExecContext(ctx, query, method, recordedBy, id)
	if err != nil {
		return false, fmt.Errorf("bookingRepo.MarkPaidManually: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("bookingRepo.MarkPaidManually rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// MarkUnpaidManually revierte el flag de pago manual de un tramo
// (deposit o balance) de una reserva (por ejemplo, el staff se equivocó
// al marcarla pagada).
func (r *bookingRepo) MarkUnpaidManually(ctx context.Context, id int64, kind string) (bool, error) {
	var query string
	switch kind {
	case domain.PaymentKindDeposit:
		query = `
			UPDATE bookings
			SET deposit_status = 'unpaid',
				deposit_method = NULL,
				deposit_recorded_by = NULL
			WHERE id = $1`
	case domain.PaymentKindBalance:
		query = `
			UPDATE bookings
			SET balance_status = 'unpaid',
				balance_method = NULL,
				balance_recorded_by = NULL
			WHERE id = $1`
	default:
		return false, fmt.Errorf("bookingRepo.MarkUnpaidManually: invalid kind %q", kind)
	}

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("bookingRepo.MarkUnpaidManually: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("bookingRepo.MarkUnpaidManually rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// ExpirePendingBookings transiciona a 'expired' las reservas pending
// de un space+slot+fecha puntual cuyo hold (expires_at) ya venció.
// Se corre antes de ExistsActiveBooking en Create: como el índice único
// parcial no puede expresar una condición por tiempo (NOW() no es
// IMMUTABLE), la liberación del hold se hace transicionando activamente
// las filas vencidas.
func (r *bookingRepo) ExpirePendingBookings(ctx context.Context, spaceID, slotID int64, date time.Time) (int64, error) {
	query := `
		UPDATE bookings
		SET status = 'expired'
		WHERE space_id = $1
		AND slot_id = $2
		AND booking_date = $3
		AND status = 'pending'
		AND expires_at IS NOT NULL
		AND expires_at < now()`

	result, err := r.db.ExecContext(ctx, query, spaceID, slotID, date)
	if err != nil {
		return 0, fmt.Errorf("bookingRepo.ExpirePendingBookings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("bookingRepo.ExpirePendingBookings rows affected: %w", err)
	}

	return rowsAffected, nil
}

// ExpireAllStalePending es un sweep global (housekeeping) para un ticker
// de background: transiciona a 'expired' cualquier reserva pending cuyo
// hold ya venció, sin filtrar por space/slot/fecha. No es la garantía de
// corrección — eso lo da el sweep acotado de ExpirePendingBookings.
func (r *bookingRepo) ExpireAllStalePending(ctx context.Context) (int64, error) {
	query := `
		UPDATE bookings
		SET status = 'expired'
		WHERE status = 'pending'
		AND expires_at IS NOT NULL
		AND expires_at < now()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("bookingRepo.ExpireAllStalePending: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("bookingRepo.ExpireAllStalePending rows affected: %w", err)
	}

	return rowsAffected, nil
}

// SetBatchID etiqueta una reserva ya creada como parte de un booking_batch
// (turno fijo o bloqueo de mantenimiento). Se llama justo después de crearla
// vía BookingService.CreateManual, que no conoce el concepto de batch.
func (r *bookingRepo) SetBatchID(ctx context.Context, bookingID, batchID int64) error {
	query := `UPDATE bookings SET batch_id = $1 WHERE id = $2`

	if _, err := r.db.ExecContext(ctx, query, batchID, bookingID); err != nil {
		return fmt.Errorf("bookingRepo.SetBatchID: %w", err)
	}
	return nil
}

// CancelFutureByBatch cancela todas las reservas activas (pending/confirmed)
// de un batch cuya fecha sea >= fromDate. Las ocurrencias pasadas quedan
// como quedaron (historial), no se tocan.
func (r *bookingRepo) CancelFutureByBatch(ctx context.Context, batchID int64, fromDate time.Time) (int64, error) {
	query := `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE batch_id = $1
		AND booking_date >= $2
		AND status IN ('pending', 'confirmed')`

	result, err := r.db.ExecContext(ctx, query, batchID, fromDate)
	if err != nil {
		return 0, fmt.Errorf("bookingRepo.CancelFutureByBatch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("bookingRepo.CancelFutureByBatch rows affected: %w", err)
	}
	return rowsAffected, nil
}
