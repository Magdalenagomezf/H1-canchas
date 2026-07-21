package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/database"

	"github.com/jmoiron/sqlx"
)

type paymentRepo struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *paymentRepo {
	return &paymentRepo{db: db}
}

// Create inserta un intento de pago nuevo dentro de una transacción.
func (r *paymentRepo) Create(ctx context.Context, tx database.Tx, p *domain.Payment) (int64, error) {
	query := `
		INSERT INTO payments (
			booking_id,
			kind,
			provider,
			preference_id,
			external_payment_id,
			status,
			amount,
			raw_status_detail
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING id`

	var id int64
	err := tx.QueryRowContext(
		ctx,
		query,
		p.BookingID,
		p.Kind,
		p.Provider,
		p.PreferenceID,
		p.ExternalPaymentID,
		p.Status,
		p.Amount,
		p.RawStatusDetail,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("paymentRepo.Create: %w", err)
	}

	return id, nil
}

// GetByID busca un pago por ID.
// Si no existe, devuelve nil, nil.
func (r *paymentRepo) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
	query := `
		SELECT
			id,
			booking_id,
			kind,
			provider,
			preference_id,
			external_payment_id,
			status,
			amount,
			raw_status_detail,
			created_at,
			updated_at
		FROM payments
		WHERE id = $1`

	var payment domain.Payment
	err := r.db.GetContext(ctx, &payment, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("paymentRepo.GetByID: %w", err)
	}

	return &payment, nil
}

// GetLatestByBookingID devuelve el intento de pago más reciente
// para una reserva dada, para el tramo (kind) indicado. Una reserva
// tiene dos tramos independientes (deposit/balance), cada uno con su
// propio historial de intentos, así que "más reciente" solo tiene
// sentido acotado a un tramo: GenerateForBooking necesita saber si ya
// existe un intento pendiente para ESE tramo puntual antes de crear
// uno nuevo (evitar preferencias duplicadas), no el intento más
// reciente sin importar a qué tramo pertenece.
// Si no hay ninguno, devuelve nil, nil.
func (r *paymentRepo) GetLatestByBookingID(ctx context.Context, bookingID int64, kind string) (*domain.Payment, error) {
	query := `
		SELECT
			id,
			booking_id,
			kind,
			provider,
			preference_id,
			external_payment_id,
			status,
			amount,
			raw_status_detail,
			created_at,
			updated_at
		FROM payments
		WHERE booking_id = $1
		AND kind = $2
		ORDER BY created_at DESC
		LIMIT 1`

	var payment domain.Payment
	err := r.db.GetContext(ctx, &payment, query, bookingID, kind)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("paymentRepo.GetLatestByBookingID: %w", err)
	}

	return &payment, nil
}

// UpdateStatus actualiza el estado de un pago dentro de una transacción.
// externalPaymentID, si viene cargado, reemplaza el valor existente
// (usado cuando el webhook confirma el ID de pago real de Mercado Pago).
// Devuelve true si realmente modificó una fila.
//
// database.Tx solo expone QueryRowContext (no ExecContext), así que
// se usa RETURNING id en vez de RowsAffected para detectar el "no-op".
func (r *paymentRepo) UpdateStatus(ctx context.Context, tx database.Tx, id int64, status string, rawDetail *string, externalPaymentID *string) (bool, error) {
	query := `
		UPDATE payments
		SET status = $1,
			raw_status_detail = $2,
			external_payment_id = COALESCE($3, external_payment_id)
		WHERE id = $4
		RETURNING id`

	var updatedID int64
	err := tx.QueryRowContext(ctx, query, status, rawDetail, externalPaymentID, id).Scan(&updatedID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("paymentRepo.UpdateStatus: %w", err)
	}

	return true, nil
}
