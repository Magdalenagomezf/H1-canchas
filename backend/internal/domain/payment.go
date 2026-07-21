package domain

import "time"

// Payment representa un intento de pago (online vía Mercado Pago o manual)
// asociado a una reserva. Es un historial append-only: una reserva puede
// tener varios intentos (rechazo + reintento, staff regenerando link).
type Payment struct {
	ID                 int64     `db:"id"`
	BookingID          int64     `db:"booking_id"`
	Kind               string    `db:"kind"`
	Provider           string    `db:"provider"`
	PreferenceID       *string   `db:"preference_id"`
	ExternalPaymentID  *string   `db:"external_payment_id"`
	Status             string    `db:"status"`
	Amount             float64   `db:"amount"`
	RawStatusDetail    *string   `db:"raw_status_detail"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// Tramo de reserva al que corresponde un intento de pago.
const (
	PaymentKindDeposit = "deposit"
	PaymentKindBalance = "balance"
)

// Estados posibles de un intento de pago (mapean a los estados
// de transacción que reporta Mercado Pago).
const (
	PaymentTxStatusPending   = "pending"
	PaymentTxStatusApproved  = "approved"
	PaymentTxStatusRejected  = "rejected"
	PaymentTxStatusCancelled = "cancelled"
	PaymentTxStatusInProcess = "in_process"
	PaymentTxStatusRefunded  = "refunded"
)

// Proveedores de pago soportados.
const (
	PaymentProviderMercadoPago = "mercadopago"
	PaymentProviderManual      = "manual"
)
