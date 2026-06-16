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

// UpdateSpaceRequest es el JSON para editar un espacio existente.
// El tipo no es editable — cambiar cancha_padel a quincho rompe datos históricos.
type UpdateSpaceRequest struct {
	Name         string  `json:"name"          binding:"required"`
	Description  *string `json:"description"`
	PricePerSlot float64 `json:"price_per_slot" binding:"required"`
}

// CreateSlotRequest es el JSON para agregar un slot a un espacio.
type CreateSlotRequest struct {
	Label       string  `json:"label"       binding:"required"`
	Description *string `json:"description"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
}

// SpaceWithSlotsResponse agrupa espacio y sus turnos en una sola respuesta.
