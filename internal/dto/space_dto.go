package dto

import "time"

type CreateSpaceRequest struct {
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	Description  *string `json:"description,omitempty"`
	PricePerSlot float64 `json:"price_per_slot" binding:"required"`
}

// SpaceResponse es lo que devuelve la API al cliente.
// No expone is_active ni campos internos.
type SpaceResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Description  *string   `json:"description,omitempty"`
	PricePerSlot float64   `json:"price_per_slot"`
	CreatedAt    time.Time `json:"created_at"`
}

// SlotResponse es lo que devuelve la API para cada turno.
type SlotResponse struct {
	ID          int64   `json:"id"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
}

// SpaceWithSlotsResponse agrupa espacio y sus turnos en una sola respuesta.
