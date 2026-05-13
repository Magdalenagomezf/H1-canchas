package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

type BookingController struct {
	bookingService *service.BookingService
}

func NewBookingController(bookingService *service.BookingService) *BookingController {
	return &BookingController{bookingService: bookingService}
}

// Create godoc — POST /bookings
// Customer crea su propia reserva desde la web.
func (h *BookingController) Create(c *gin.Context) {
	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.BookingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingDate.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	booking, err := h.bookingService.Create(
		c.Request.Context(),
		userID,
		req.SpaceID,
		req.SlotID,
		date,
	)
	if err != nil {
		h.handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.FromBooking(*booking))
}

// CreateManual godoc — POST /bookings/manual
// Recepcionista/admin carga una reserva para un cliente sin cuenta.
func (h *BookingController) CreateManual(c *gin.Context) {
	var req dto.CreateManualBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.BookingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingDate.Error()})
		return
	}

	createdBy := c.GetInt64("user_id")
	if createdBy <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	booking, err := h.bookingService.CreateManual(
		c.Request.Context(),
		createdBy,
		req.SpaceID,
		req.SlotID,
		date,
		req.CustomerName,
		req.CustomerPhone,
	)
	if err != nil {
		h.handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.FromBooking(*booking))
}

// GetMyBookings godoc — GET /bookings
// Customer ve sus reservas. Recepcionista/admin ve todas.
func (h *BookingController) GetMyBookings(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	role := c.GetString("role")

	var (
		bookings []domain.Booking
		err      error
	)

	if role == domain.RoleReceptionist || role == domain.RoleAdmin {
		var dateFilter *time.Time

		if dateStr := c.Query("date"); dateStr != "" {
			d, parseErr := time.Parse("2006-01-02", dateStr)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingDate.Error()})
				return
			}

			dateFilter = &d
		}

		bookings, err = h.bookingService.GetAll(c.Request.Context(), dateFilter)
	} else {
		bookings, err = h.bookingService.GetMyBookings(c.Request.Context(), userID)
	}

	if err != nil {
		h.handleBookingError(c, err)
		return
	}

	response := make([]dto.BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		response = append(response, dto.FromBooking(booking))
	}

	c.JSON(http.StatusOK, response)
}

// Cancel godoc — PATCH /bookings/:id/cancel
func (h *BookingController) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingID.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	role := c.GetString("role")

	err = h.bookingService.Cancel(c.Request.Context(), id, userID, role)
	if err != nil {
		h.handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reserva cancelada correctamente"})
}

func (h *BookingController) handleBookingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidBookingID),
		errors.Is(err, service.ErrInvalidSpaceID),
		errors.Is(err, service.ErrInvalidSlotID),
		errors.Is(err, service.ErrInvalidBookingDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrNameRequired),
		errors.Is(err, service.ErrPhoneRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrSlotDoesNotBelong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso no encontrado"})

	case errors.Is(err, service.ErrSlotNotAvailable):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "no tenés permiso para realizar esta acción"})

	case errors.Is(err, service.ErrBookingAlreadyCancelled):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrBookingAlreadyCompleted):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
	}
}
