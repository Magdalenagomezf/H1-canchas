package domain

import "time"

// BookingBatchDetail es un BookingBatch con nombre de cancha y label de
// turno ya cruzados, para listados del panel de recepcionista.
// No es una tabla nueva: sale de un JOIN entre booking_batches, spaces
// y space_slots (este último LEFT JOIN porque slot_id puede ser NULL
// en un bloqueo de mantenimiento que abarca todos los turnos del día).
type BookingBatchDetail struct {
	ID        int64     `db:"id"`
	Type      string    `db:"type"`
	Weekday   *int      `db:"weekday"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	Reason    string    `db:"reason"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`

	SpaceID   int64  `db:"space_id"`
	SpaceName string `db:"space_name"`
	SpaceType string `db:"space_type"`

	SlotID    *int64  `db:"slot_id"`
	SlotLabel *string `db:"slot_label"`
}
