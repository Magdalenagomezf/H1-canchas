package dto

import (
	"time"

	"H1-canchas/internal/domain"
)

// CreateBookingRequest representa lo que manda un customer desde la web.
// El usuario no manda customer_user_id ni created_by porque eso sale del JWT.
type CreateBookingRequest struct {
	SpaceID     int64  `json:"space_id" binding:"required"`
	SlotID      int64  `json:"slot_id" binding:"required"`
	BookingDate string `json:"booking_date" binding:"required"` // formato: "2026-06-15"
}

// CreateManualBookingRequest representa una reserva cargada por recepcionista/admin.
// En este caso el cliente puede no tener cuenta, por eso se piden nombre y teléfono.
type CreateManualBookingRequest struct {
	SpaceID       int64  `json:"space_id" binding:"required"`
	SlotID        int64  `json:"slot_id" binding:"required"`
	BookingDate   string `json:"booking_date" binding:"required"` // formato: "2026-06-15"`
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
}

// BookingResponse representa lo que devuelve la API.
// No devuelve password ni datos internos de users.
// booking_date se devuelve como string para evitar hora/zona horaria innecesaria.
type BookingResponse struct {
	ID             int64  `json:"id"`
	CustomerUserID *int64 `json:"customer_user_id,omitempty"`
	CreatedBy      int64  `json:"created_by"`

	CustomerName  *string `json:"customer_name,omitempty"`
	CustomerPhone *string `json:"customer_phone,omitempty"`

	SpaceID     int64   `json:"space_id"`
	SlotID      int64   `json:"slot_id"`
	BookingDate string  `json:"booking_date"`
	Status      string  `json:"status"`
	TotalPrice  float64 `json:"total_price"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromBooking convierte domain.Booking a BookingResponse.
func FromBooking(b domain.Booking) BookingResponse {
	return BookingResponse{
		ID:             b.ID,
		CustomerUserID: b.CustomerUserID,
		CreatedBy:      b.CreatedBy,

		CustomerName:  b.CustomerName,
		CustomerPhone: b.CustomerPhone,

		SpaceID:     b.SpaceID,
		SlotID:      b.SlotID,
		BookingDate: b.BookingDate.Format("2006-01-02"),
		Status:      b.Status,
		TotalPrice:  b.TotalPrice,

		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}
