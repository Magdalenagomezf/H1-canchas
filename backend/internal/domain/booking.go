package domain

import "time"

// Booking representa una reserva. El cliente puede ser un usuario
// registrado (CustomerUserID != nil) o anónimo (solo nombre y teléfono).
type Booking struct {
    ID             int64      `db:"id"`
    CustomerUserID *int64     `db:"customer_user_id"` // nil si es anónimo
    CreatedBy      int64      `db:"created_by"`
    CustomerName   *string    `db:"customer_name"`
    CustomerPhone  *string    `db:"customer_phone"`
    SpaceID        int64      `db:"space_id"`
    SlotID         int64      `db:"slot_id"`
    BookingDate    time.Time  `db:"booking_date"`
    Status         string     `db:"status"`
    TotalPrice     float64    `db:"total_price"`
    CreatedAt      time.Time  `db:"created_at"`
    UpdatedAt      time.Time  `db:"updated_at"`
}

// Estados posibles de una reserva.
const (
    BookingStatusPending  = "pending"
    BookingStatusConfirmed = "confirmed"
    BookingStatusCancelled = "cancelled"
    BookingStatusCompleted = "completed"
)