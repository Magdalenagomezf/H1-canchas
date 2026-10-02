package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

type bookingBatchServiceI interface {
	CreateRecurringTeacherBatch(
		ctx context.Context, createdBy int64, spaceID, slotID int64, weekday int,
		customerName, customerPhone string, startDate, endDate time.Time, note string,
	) (*domain.BookingBatch, []time.Time, []time.Time, error)
	CreateMaintenanceBatch(
		ctx context.Context, createdBy int64, spaceID int64, slotID *int64,
		startDate, endDate time.Time, reason string,
	) (*domain.BookingBatch, []time.Time, []time.Time, error)
	CancelBatch(ctx context.Context, batchID int64) error
	GetAll(ctx context.Context, batchType *string) ([]domain.BookingBatchDetail, error)
}

type BookingBatchController struct {
	batchService bookingBatchServiceI
}

func NewBookingBatchController(batchService bookingBatchServiceI) *BookingBatchController {
	return &BookingBatchController{batchService: batchService}
}

// CreateRecurring godoc — POST /bookings/recurring
// Recepcionista/admin reserva un turno fijo semanal para un profesor.
func (h *BookingBatchController) CreateRecurring(c *gin.Context) {
	var req dto.CreateRecurringBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	startDate, endDate, ok := parseDateRange(c, req.StartDate, req.EndDate)
	if !ok {
		return
	}

	createdBy := c.GetInt64("user_id")
	if createdBy <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	batch, created, skipped, err := h.batchService.CreateRecurringTeacherBatch(
		c.Request.Context(), createdBy, req.SpaceID, req.SlotID, *req.Weekday,
		req.CustomerName, req.CustomerPhone, startDate, endDate, req.Note,
	)
	if err != nil {
		h.handleBatchError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toCreateBatchResponse(*batch, req.SpaceID, &req.SlotID, created, skipped))
}

// CreateBlock godoc — POST /bookings/block
// Recepcionista/admin bloquea una cancha (o un turno puntual) por mantenimiento.
func (h *BookingBatchController) CreateBlock(c *gin.Context) {
	var req dto.CreateMaintenanceBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	startDate, endDate, ok := parseDateRange(c, req.StartDate, req.EndDate)
	if !ok {
		return
	}

	createdBy := c.GetInt64("user_id")
	if createdBy <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	batch, created, skipped, err := h.batchService.CreateMaintenanceBatch(
		c.Request.Context(), createdBy, req.SpaceID, req.SlotID, startDate, endDate, req.Reason,
	)
	if err != nil {
		h.handleBatchError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toCreateBatchResponse(*batch, req.SpaceID, req.SlotID, created, skipped))
}

// GetAll godoc — GET /bookings/batches?type=recurring_teacher|maintenance
func (h *BookingBatchController) GetAll(c *gin.Context) {
	var batchType *string
	if t := c.Query("type"); t != "" {
		batchType = &t
	}

	rows, err := h.batchService.GetAll(c.Request.Context(), batchType)
	if err != nil {
		h.handleBatchError(c, err)
		return
	}

	response := make([]dto.BatchResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.FromBookingBatchDetail(row))
	}
	c.JSON(http.StatusOK, response)
}

// Cancel godoc — PATCH /bookings/batches/:id/cancel
// Cancela todas las ocurrencias futuras del turno fijo o bloqueo.
func (h *BookingBatchController) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.batchService.CancelBatch(c.Request.Context(), id); err != nil {
		h.handleBatchError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "turno fijo / bloqueo cancelado correctamente"})
}

func parseDateRange(c *gin.Context, startStr, endStr string) (time.Time, time.Time, bool) {
	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidDateRange.Error()})
		return time.Time{}, time.Time{}, false
	}
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrInvalidDateRange.Error()})
		return time.Time{}, time.Time{}, false
	}
	return startDate, endDate, true
}

func toCreateBatchResponse(batch domain.BookingBatch, spaceID int64, slotID *int64, created, skipped []time.Time) dto.CreateBatchResponse {
	createdStrs := make([]string, len(created))
	for i, d := range created {
		createdStrs[i] = d.Format("2006-01-02")
	}
	skippedStrs := make([]string, len(skipped))
	for i, d := range skipped {
		skippedStrs[i] = d.Format("2006-01-02")
	}

	resp := dto.BatchResponse{
		ID:        batch.ID,
		Type:      batch.Type,
		Weekday:   batch.Weekday,
		StartDate: batch.StartDate.Format("2006-01-02"),
		EndDate:   batch.EndDate.Format("2006-01-02"),
		Reason:    batch.Reason,
		Status:    batch.Status,
		CreatedAt: batch.CreatedAt,
		Space:     dto.SpaceInfo{ID: spaceID},
	}
	if slotID != nil {
		resp.Slot = &dto.SlotInfo{ID: *slotID}
	}

	return dto.CreateBatchResponse{
		Batch:        resp,
		CreatedDates: createdStrs,
		SkippedDates: skippedStrs,
	}
}

func (h *BookingBatchController) handleBatchError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNameRequired),
		errors.Is(err, service.ErrPhoneRequired),
		errors.Is(err, service.ErrInvalidPhone),
		errors.Is(err, service.ErrReasonRequired),
		errors.Is(err, service.ErrInvalidWeekday),
		errors.Is(err, service.ErrInvalidDateRange),
		errors.Is(err, service.ErrDateRangeTooLong),
		errors.Is(err, service.ErrInvalidBookingDate),
		errors.Is(err, service.ErrSlotDoesNotBelong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrNotFound), errors.Is(err, service.ErrBatchNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrBatchAlreadyCancelled):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
	}
}
