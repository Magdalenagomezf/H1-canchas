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

type BookingDetailResponse struct {
	ID          int64     `json:"id"`
	BookingDate string    `json:"booking_date"`
	Status      string    `json:"status"`
	TotalPrice  float64   `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Space    SpaceInfo    `json:"space"`
	Slot     SlotInfo     `json:"slot"`
	Customer CustomerInfo `json:"customer"`
}

type SpaceInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type SlotInfo struct {
	ID        int64   `json:"id"`
	Label     string  `json:"label"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
}

type CustomerInfo struct {
	UserID   *int64 `json:"user_id,omitempty"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	IsManual bool   `json:"is_manual"`
}

func FromBookingDetail(b domain.BookingDetail) BookingDetailResponse {
	customer := CustomerInfo{
		UserID:   b.CustomerUserID,
		IsManual: b.CustomerUserID == nil,
	}

	if b.CustomerUserID != nil {
		// reserva de usuario registrado — usamos datos de la tabla users
		if b.CustomerUserName != nil {
			customer.Name = *b.CustomerUserName
		}
		if b.CustomerUserPhone != nil {
			customer.Phone = *b.CustomerUserPhone
		}
	} else {
		// reserva manual — usamos los campos del booking
		if b.CustomerName != nil {
			customer.Name = *b.CustomerName
		}
		if b.CustomerPhone != nil {
			customer.Phone = *b.CustomerPhone
		}
	}

	return BookingDetailResponse{
		ID:          b.ID,
		BookingDate: b.BookingDate.Format("2006-01-02"),
		Status:      b.Status,
		TotalPrice:  b.TotalPrice,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		Space: SpaceInfo{
			ID:   b.SpaceID,
			Name: b.SpaceName,
			Type: b.SpaceType,
		},
		Slot: SlotInfo{
			ID:        b.SlotID,
			Label:     b.SlotLabel,
			StartTime: b.SlotStartTime,
			EndTime:   b.SlotEndTime,
		},
		Customer: customer,
	}
}
