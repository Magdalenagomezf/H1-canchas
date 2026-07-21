package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"
	"H1-canchas/pkg/mercadopago"

	"github.com/gin-gonic/gin"
)

type paymentServiceI interface {
	GenerateForBooking(ctx context.Context, bookingID int64, kind string, requesterID int64, requesterRole string) (*mercadopago.PreferenceResponse, error)
	HandleWebhook(ctx context.Context, xSignature, xRequestID, dataID string) error
	MarkPaidManually(ctx context.Context, bookingID int64, kind, method string, recordedByID int64) error
	MarkUnpaidManually(ctx context.Context, bookingID int64, kind string) error
}

type PaymentController struct {
	paymentService paymentServiceI
}

func NewPaymentController(paymentService paymentServiceI) *PaymentController {
	return &PaymentController{paymentService: paymentService}
}

// GeneratePreference godoc — POST /bookings/:id/payments
func (h *PaymentController) GeneratePreference(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingID.Error()})
		return
	}

	var req dto.GeneratePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	role := c.GetString("role")

	resp, err := h.paymentService.GenerateForBooking(c.Request.Context(), id, req.Kind, userID, role)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.FromPreferenceResponse(resp))
}

// MarkPayment godoc — PATCH /bookings/:id/payment
func (h *PaymentController) MarkPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingID.Error()})
		return
	}

	var req dto.MarkPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	if req.PaymentStatus == "paid" {
		// binding:"oneof" can't express "required only when a sibling field
		// has a given value", so this check is manual.
		if req.PaymentMethod == nil || *req.PaymentMethod == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidPaymentMethod.Error()})
			return
		}
		recordedBy := c.GetInt64("user_id")
		if err := h.paymentService.MarkPaidManually(c.Request.Context(), id, req.Kind, *req.PaymentMethod, recordedBy); err != nil {
			h.handlePaymentError(c, err)
			return
		}
	} else {
		if err := h.paymentService.MarkUnpaidManually(c.Request.Context(), id, req.Kind); err != nil {
			h.handlePaymentError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "estado de pago actualizado"})
}

// Webhook godoc — POST /payments/webhook
// Público, sin auth: lo llama Mercado Pago, no el front.
func (h *PaymentController) Webhook(c *gin.Context) {
	// "Pagos (legacy)" (el evento recomendado por MP para Checkout Pro) manda
	// ?id=X&topic=payment — no ?data.id=X&type=payment (formato Webhooks v2).
	// Soportamos ambos por si el evento suscripto cambia más adelante.
	topic := c.Query("topic")
	if topic == "" {
		topic = c.Query("type")
	}
	if topic != "" && topic != "payment" {
		// merchant_order u otro topic no relacionado a pagos: no hay nada que hacer.
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	dataID := c.Query("data.id")
	if dataID == "" {
		dataID = c.Query("id")
	}

	xSignature := c.GetHeader("x-signature")
	xRequestID := c.GetHeader("x-request-id")

	err := h.paymentService.HandleWebhook(c.Request.Context(), xSignature, xRequestID, dataID)
	if err != nil {
		// Any failure (GetPayment hiccup, DB blip, transient network) is
		// exactly what MP's notification retry is built to heal, so we
		// deliberately return 500 instead of swallowing it as a 200.
		log.Printf("payments webhook error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentController) handlePaymentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPaymentKind),
		errors.Is(err, service.ErrInvalidPaymentMethod),
		errors.Is(err, service.ErrInvalidPaymentStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso no encontrado"})

	case errors.Is(err, service.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "no tenés permiso para realizar esta acción"})

	case errors.Is(err, service.ErrBookingExpired),
		errors.Is(err, service.ErrBookingNotPayable),
		errors.Is(err, service.ErrPaymentAlreadyCompleted):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrMercadoPagoUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
	}
}
