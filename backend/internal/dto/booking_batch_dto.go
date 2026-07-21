package dto

import (
	"time"

	"H1-canchas/internal/domain"
)

// CreateRecurringBookingRequest es lo que manda el recepcionista para
// reservar un turno fijo semanal (ej. un profesor todos los lunes 18-19hs).
type CreateRecurringBookingRequest struct {
	SpaceID       int64  `json:"space_id" binding:"required"`
	SlotID        int64  `json:"slot_id" binding:"required"`
	Weekday       *int   `json:"weekday" binding:"required"` // 0=domingo .. 6=sábado. Puntero porque 0 es un valor válido.
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	StartDate     string `json:"start_date" binding:"required"` // "2026-06-15"
	EndDate       string `json:"end_date" binding:"required"`
	Note          string `json:"note"`
}

// CreateMaintenanceBlockRequest es lo que manda el recepcionista para
// bloquear una cancha por mantenimiento en un rango de fechas.
// SlotID es opcional: si no se manda, bloquea todos los turnos activos
// del espacio en ese rango (ej. "la cancha entera está en obra").
type CreateMaintenanceBlockRequest struct {
	SpaceID   int64  `json:"space_id" binding:"required"`
	SlotID    *int64 `json:"slot_id"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}

// BatchResponse es lo que devuelve la API para un turno fijo o bloqueo.
type BatchResponse struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Weekday   *int      `json:"weekday,omitempty"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`

	Space SpaceInfo `json:"space"`
	Slot  *SlotInfo `json:"slot,omitempty"` // nil = aplica a todos los turnos del espacio
}

func FromBookingBatchDetail(b domain.BookingBatchDetail) BatchResponse {
	resp := BatchResponse{
		ID:        b.ID,
		Type:      b.Type,
		Weekday:   b.Weekday,
		StartDate: b.StartDate.Format("2006-01-02"),
		EndDate:   b.EndDate.Format("2006-01-02"),
		Reason:    b.Reason,
		Status:    b.Status,
		CreatedAt: b.CreatedAt,
		Space:     SpaceInfo{ID: b.SpaceID, Name: b.SpaceName, Type: b.SpaceType},
	}
	if b.SlotID != nil {
		resp.Slot = &SlotInfo{ID: *b.SlotID}
		if b.SlotLabel != nil {
			resp.Slot.Label = *b.SlotLabel
		}
	}
	return resp
}

// CreateBatchResponse envuelve el batch creado junto con el detalle de qué
// fechas se pudieron reservar/bloquear y cuáles ya estaban ocupadas.
type CreateBatchResponse struct {
	Batch        BatchResponse `json:"batch"`
	CreatedDates []string      `json:"created_dates"`
	SkippedDates []string      `json:"skipped_dates"`
}
