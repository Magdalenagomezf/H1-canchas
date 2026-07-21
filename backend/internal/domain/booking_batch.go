package domain

import "time"

// BookingBatch agrupa un conjunto de bookings creados juntos siguiendo un
// patrón de fechas: un turno fijo semanal para un profesor, o un bloqueo
// de mantenimiento en un rango de días. Cada booking generado queda
// referenciado por bookings.batch_id, lo que permite cancelar toda la
// serie de una sola vez en lugar de una por una.
type BookingBatch struct {
	ID        int64     `db:"id"`
	Type      string    `db:"type"`
	SpaceID   int64     `db:"space_id"`
	SlotID    *int64    `db:"slot_id"` // nil = todos los turnos activos del espacio (solo maintenance)
	Weekday   *int      `db:"weekday"` // 0=domingo .. 6=sábado, solo recurring_teacher
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	Reason    string    `db:"reason"`
	CreatedBy int64     `db:"created_by"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

const (
	BatchTypeRecurringTeacher = "recurring_teacher"
	BatchTypeMaintenance      = "maintenance"

	BatchStatusActive    = "active"
	BatchStatusCancelled = "cancelled"
)
