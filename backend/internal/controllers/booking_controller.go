package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

type bookingServiceI interface {
	Create(ctx context.Context, customerUserID int64, spaceID, slotID int64, date time.Time) (*domain.Booking, error)
	CreateManual(ctx context.Context, createdBy int64, spaceID, slotID int64, date time.Time, customerName, customerPhone string) (*domain.Booking, error)
	GetMyBookings(ctx context.Context, userID int64) ([]domain.BookingDetail, error)
	GetAll(ctx context.Context, date *time.Time) ([]domain.BookingDetail, error)
	Cancel(ctx context.Context, bookingID, requesterID int64, requesterRole string) error
	GetByID(ctx context.Context, bookingID, requesterID int64, requesterRole string) (*domain.BookingDetail, error)
}

type BookingController struct {
	bookingService bookingServiceI
}

func NewBookingController(bookingService bookingServiceI) *BookingController {
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

	role := c.GetString("role")
	detail, err := h.bookingService.GetByID(c.Request.Context(), booking.ID, userID, role)
	if err != nil {
		h.handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.FromBookingDetail(*detail))
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

	role := c.GetString("role")
	detail, err := h.bookingService.GetByID(c.Request.Context(), booking.ID, createdBy, role)
	if err != nil {
		h.handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.FromBookingDetail(*detail))
}

// GetAll godoc — GET /admin/bookings
// Staff ve todas las reservas, con filtro opcional por fecha.
func (h *BookingController) GetAll(c *gin.Context) {
	var dateFilter *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingDate.Error()})
			return
		}
		dateFilter = &d
	}

	rows, err := h.bookingService.GetAll(c.Request.Context(), dateFilter)
	if err != nil {
		log.Printf("GetAll error: %v", err)
		h.handleBookingError(c, err)
		return
	}

	response := make([]dto.BookingDetailResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.FromBookingDetail(row))
	}
	c.JSON(http.StatusOK, response)
}

// GetMyBookings godoc — GET /bookings
// Devuelve las reservas del usuario autenticado, sin importar el rol.
func (h *BookingController) GetMyBookings(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	rows, err := h.bookingService.GetMyBookings(c.Request.Context(), userID)
	if err != nil {
		log.Printf("GetMyBookings error: %v", err)
		h.handleBookingError(c, err)
		return
	}

	response := make([]dto.BookingDetailResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.FromBookingDetail(row))
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

func (h *BookingController) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidBookingID.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	role := c.GetString("role")

	row, err := h.bookingService.GetByID(c.Request.Context(), id, userID, role)
	if err != nil {
		log.Printf("GetByID error: %v", err)
		h.handleBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.FromBookingDetail(*row))
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

	case errors.Is(err, service.ErrCancelRequiresStaffAfterPayment):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
	}
}
