package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

// passwordResetRequestedMessage is the only answer to a reset request: it
// must not reveal whether the email belongs to an account.
const passwordResetRequestedMessage = "Si el email corresponde a una cuenta, te mandamos un link para restablecer la contraseña."

type passwordResetServiceI interface {
	RequestReset(email string)
	Confirm(ctx context.Context, token, password string) error
}

type PasswordResetController struct {
	svc passwordResetServiceI
}

func NewPasswordResetController(svc passwordResetServiceI) *PasswordResetController {
	return &PasswordResetController{svc: svc}
}

// Request — POST /auth/password-reset
// Always 202 with the same body; the email goes out in the background.
func (h *PasswordResetController) Request(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el email es obligatorio"})
		return
	}

	h.svc.RequestReset(req.Email)
	c.JSON(http.StatusAccepted, gin.H{"message": passwordResetRequestedMessage})
}

// Confirm — POST /auth/password-reset/confirm
func (h *PasswordResetController) Confirm(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido"})
		return
	}

	err := h.svc.Confirm(c.Request.Context(), req.Token, req.Password)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, service.ErrInvalidResetToken):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_reset_token"})
	case errors.Is(err, service.ErrPasswordRequired), errors.Is(err, service.ErrPasswordTooShort), errors.Is(err, service.ErrPasswordTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("PasswordReset Confirm error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
	}
}
