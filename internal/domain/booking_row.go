package domain

import "time"

// BookingDetail representa una reserva con datos ya cruzados
// para listados del front.
// No es una tabla nueva: sale de un JOIN entre bookings, spaces,
// space_slots y users.
type BookingDetail struct {
	ID          int64     `db:"id"`
	BookingDate time.Time `db:"booking_date"`
	Status      string    `db:"status"`
	TotalPrice  float64   `db:"total_price"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	SpaceID   int64  `db:"space_id"`
	SpaceName string `db:"space_name"`
	SpaceType string `db:"space_type"`

	SlotID        int64   `db:"slot_id"`
	SlotLabel     string  `db:"slot_label"`
	SlotStartTime *string `db:"slot_start_time"`
	SlotEndTime   *string `db:"slot_end_time"`

	CustomerUserID    *int64  `db:"customer_user_id"`
	CustomerName      *string `db:"customer_name"`
	CustomerPhone     *string `db:"customer_phone"`
	CustomerUserName  *string `db:"customer_user_name"`
	CustomerUserPhone *string `db:"customer_user_phone"`
}
