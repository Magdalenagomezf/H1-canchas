package dto

import "H1-canchas/pkg/mercadopago"

// GeneratePreferenceRequest representa lo que manda el front al pedir un
// link de pago de Mercado Pago para un tramo de la reserva.
type GeneratePreferenceRequest struct {
	Kind string `json:"kind" binding:"required,oneof=deposit balance"`
}

type PreferenceResponseDTO struct {
	PreferenceID     string `json:"preference_id"`
	InitPoint        string `json:"init_point"`
	SandboxInitPoint string `json:"sandbox_init_point"`
}

func FromPreferenceResponse(r *mercadopago.PreferenceResponse) PreferenceResponseDTO {
	return PreferenceResponseDTO{
		PreferenceID:     r.ID,
		InitPoint:        r.InitPoint,
		SandboxInitPoint: r.SandboxInitPoint,
	}
}

// MarkPaymentRequest representa la carga manual de un pago por staff
// (efectivo, transferencia, posnet) o la reversión de un pago cargado
// por error.
type MarkPaymentRequest struct {
	Kind          string  `json:"kind" binding:"required,oneof=deposit balance"`
	PaymentStatus string  `json:"payment_status" binding:"required,oneof=paid unpaid"`
	PaymentMethod *string `json:"payment_method,omitempty"`
}
