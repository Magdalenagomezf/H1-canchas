package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

type SpaceController struct {
	spaceService *service.SpaceService
}

func NewSpaceController(spaceService *service.SpaceService) *SpaceController {
	return &SpaceController{spaceService: spaceService}
}
func (h *SpaceController) Create(c *gin.Context) {
	var req dto.CreateSpaceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: " + err.Error()})
		return
	}

	id, err := h.spaceService.Create(
		c.Request.Context(),
		req.Name,
		req.Type,
		req.Description,
		req.PricePerSlot,
	)

	if err != nil {
		if errors.Is(err, service.ErrSpaceNameRequired) ||
			errors.Is(err, service.ErrInvalidSpaceType) ||
			errors.Is(err, service.ErrInvalidSpacePrice) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"message": "espacio creado correctamente",
	})
}

// GetAll godoc
// GET /spaces
func (h *SpaceController) GetAll(c *gin.Context) {
	spaces, err := h.spaceService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	// Convertimos domain → DTO antes de responder
	response := make([]dto.SpaceResponse, len(spaces))
	for i, s := range spaces {
		response[i] = toSpaceResponse(s)
	}

	c.JSON(http.StatusOK, response)
}

// GetByID godoc
// GET /spaces/:id
func (h *SpaceController) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	space, err := h.spaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSpaceID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "espacio no encontrado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusOK, toSpaceResponse(*space))
}

// GetSlots godoc
// GET /spaces/:id/slots
func (h *SpaceController) GetSlots(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	slots, err := h.spaceService.GetSlots(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSpaceID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "espacio no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	response := make([]dto.SlotResponse, len(slots))
	for i, s := range slots {
		response[i] = toSlotResponse(s)
	}

	c.JSON(http.StatusOK, response)
}
func (h *SpaceController) Deactivate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	err = h.spaceService.Deactivate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSpaceID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "espacio no encontrado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.Status(http.StatusNoContent)
}

// toSpaceResponse convierte domain.Space → dto.SpaceResponse.
// Es privada porque solo la usa este controller.
func toSpaceResponse(s domain.Space) dto.SpaceResponse {
	return dto.SpaceResponse{
		ID:           s.ID,
		Name:         s.Name,
		Type:         s.Type,
		Description:  s.Description,
		PricePerSlot: s.PricePerSlot,
		CreatedAt:    s.CreatedAt,
	}
}

// toSlotResponse convierte domain.SpaceSlot → dto.SlotResponse.
func toSlotResponse(s domain.SpaceSlot) dto.SlotResponse {
	return dto.SlotResponse{
		ID:          s.ID,
		Label:       s.Label,
		Description: s.Description,
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
	}
}
