package domain

// SpaceSlot define un turno disponible para un espacio.
// Ej: "Noche" → 19:00 a 00:00 para un quincho.
// Ej: "08:00-09:00" → franja horaria de una cancha.
type SpaceSlot struct {
	ID          int64   `db:"id"`
	SpaceID     int64   `db:"space_id"`
	Label       string  `db:"label"`
	Description *string `db:"description"`
	StartTime   *string `db:"start_time"` // "HH:MM:SS" → viene así de MySQL
	EndTime     *string `db:"end_time"`
	IsActive    bool    `db:"is_active"`
}
