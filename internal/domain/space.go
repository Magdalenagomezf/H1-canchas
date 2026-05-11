package domain

import "time"

// Space es un espacio reservable: cancha de pádel, tenis o quincho.
type Space struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Type         string    `db:"type"`
	Description  *string   `db:"description"`
	PricePerSlot float64   `db:"price_per_slot"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
}

// Tipos de espacio disponibles.
const (
	SpaceTypePadel   = "cancha_padel"
	SpaceTypeFutbol  = "cancha_futbol"
	SpaceTypeQuincho = "quincho"
)
