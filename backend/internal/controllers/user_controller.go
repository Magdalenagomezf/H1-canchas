package controllers

import (
	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthController maneja los endpoints de autenticación.
// Solo conoce el service, nunca el repository ni la BD.
type AuthController struct {
	userService *service.UserService
}

func NewAuthController(userService *service.UserService) *AuthController {
	return &AuthController{userService: userService}
}

// Register godoc
// POST /auth/register
func (h *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var email *string
	if req.Email != "" {
		email = &req.Email
	}

	user, token, err := h.userService.Register(c.Request.Context(), req.Name, req.Phone, req.Password, email)
	if err != nil {
		log.Printf("Register error: %v", err)
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: token,
		User:  dto.UserResponse{ID: user.ID, Name: user.Name, Phone: user.Phone, Role: user.Role},
	})
}

// CreateStaff godoc — POST /admin/users
// Admin crea un usuario con rol receptionist o admin.
func (h *AuthController) CreateStaff(c *gin.Context) {
	var req dto.CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	var email *string
	if req.Email != "" {
		email = &req.Email
	}

	id, err := h.userService.CreateStaff(c.Request.Context(), req.Name, req.Phone, req.Password, req.Role, email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRole),
			errors.Is(err, service.ErrNameRequired),
			errors.Is(err, service.ErrPhoneRequired),
			errors.Is(err, service.ErrPasswordRequired),
			errors.Is(err, service.ErrPasswordTooShort):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrPhoneAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("CreateStaff error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// Login godoc
// POST /auth/login
func (h *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.userService.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) ||
			errors.Is(err, service.ErrUserInactive) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  dto.UserResponse{ID: user.ID, Name: user.Name, Phone: user.Phone, Role: user.Role},
	})
}
