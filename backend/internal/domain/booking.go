package domain

import "time"

// Booking representa una reserva. El cliente puede ser un usuario
// registrado (CustomerUserID != nil) o anónimo (solo nombre y teléfono).
type Booking struct {
    ID                int64      `db:"id"`
    CustomerUserID    *int64     `db:"customer_user_id"` // nil si es anónimo
    CreatedBy         int64      `db:"created_by"`
    CustomerName      *string    `db:"customer_name"`
    CustomerPhone     *string    `db:"customer_phone"`
    SpaceID           int64      `db:"space_id"`
    SlotID            int64      `db:"slot_id"`
    BookingDate       time.Time  `db:"booking_date"`
    Status            string     `db:"status"`
    TotalPrice        float64    `db:"total_price"`
    DepositAmount     float64    `db:"deposit_amount"`
    DepositStatus     string     `db:"deposit_status"`
    DepositMethod     *string    `db:"deposit_method"`
    DepositRecordedBy *int64     `db:"deposit_recorded_by"`
    ExpiresAt         *time.Time `db:"expires_at"`
    BalanceAmount     float64    `db:"balance_amount"`
    BalanceStatus     string     `db:"balance_status"`
    BalanceMethod     *string    `db:"balance_method"`
    BalanceRecordedBy *int64     `db:"balance_recorded_by"`
    CreatedAt         time.Time  `db:"created_at"`
    UpdatedAt         time.Time  `db:"updated_at"`
    BatchID           *int64     `db:"batch_id"` // nil si no pertenece a un turno fijo ni a un bloqueo de mantenimiento
}

// Estados posibles de una reserva.
const (
    BookingStatusPending   = "pending"
    BookingStatusConfirmed = "confirmed"
    BookingStatusCancelled = "cancelled"
    BookingStatusCompleted = "completed"
    BookingStatusExpired   = "expired"
)

// Estados posibles de un tramo de pago (seña o saldo) de una reserva.
// Pool compartido: DepositStatus y BalanceStatus usan los mismos valores,
// no hay un enum separado por tramo.
const (
    PaymentStatusUnpaid   = "unpaid"
    PaymentStatusPending  = "pending"
    PaymentStatusPaid     = "paid"
    PaymentStatusRefunded = "refunded"
)

// Métodos de pago soportados para marcar un tramo (seña o saldo) como pagado.
// Pool compartido: DepositMethod y BalanceMethod usan los mismos valores.
const (
    PaymentMethodCash        = "efectivo"
    PaymentMethodTransfer    = "transferencia"
    PaymentMethodPosnet      = "posnet"
    PaymentMethodMercadoPago = "mercado_pago"
)