package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/dto"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

type userServiceI interface {
	Register(ctx context.Context, name, phone, password string, email *string) (*domain.User, string, error)
	CreateStaff(ctx context.Context, name, phone, password, role string, email *string) (int64, error)
	Login(ctx context.Context, phone, password string) (*domain.User, string, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	UpdateRole(ctx context.Context, id int64, newRole string) error
	DeleteUser(ctx context.Context, targetID, requesterID int64) error
}

// AuthController maneja los endpoints de autenticación.
// Solo conoce el service, nunca el repository ni la BD.
type AuthController struct {
	userService userServiceI
}

func NewAuthController(userService userServiceI) *AuthController {
	return &AuthController{userService: userService}
}

// userInputErrorStatus maps user-input domain errors (validation and
// uniqueness) to their HTTP status. ok is false for anything else.
func userInputErrorStatus(err error) (status int, ok bool) {
	switch {
	case errors.Is(err, service.ErrInvalidRole),
		errors.Is(err, service.ErrNameRequired),
		errors.Is(err, service.ErrPhoneRequired),
		errors.Is(err, service.ErrInvalidPhone),
		errors.Is(err, service.ErrPasswordRequired),
		errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrEmailRequired),
		errors.Is(err, service.ErrInvalidEmail):
		return http.StatusBadRequest, true
	case errors.Is(err, service.ErrPhoneAlreadyExists),
		errors.Is(err, service.ErrEmailAlreadyExists):
		return http.StatusConflict, true
	}
	return 0, false
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
		if status, ok := userInputErrorStatus(err); ok {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		log.Printf("Register error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: token,
		User:  dto.NewUserResponse(user),
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
		if status, ok := userInputErrorStatus(err); ok {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		log.Printf("CreateStaff error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ListUsers godoc — GET /admin/users
func (h *AuthController) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers(c.Request.Context())
	if err != nil {
		log.Printf("ListUsers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	resp := make([]dto.UserResponse, len(users))
	for i, u := range users {
		resp[i] = dto.NewUserResponse(&u)
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateRole godoc — PATCH /admin/users/:id/role
func (h *AuthController) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	if err := h.userService.UpdateRole(c.Request.Context(), id, req.Role); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		default:
			log.Printf("UpdateRole error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rol actualizado"})
}

// DeleteUser godoc — DELETE /admin/users/:id
func (h *AuthController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	requesterID := c.GetInt64("user_id")

	if err := h.userService.DeleteUser(c.Request.Context(), id, requesterID); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotDeleteSelf):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		case errors.Is(err, service.ErrUserHasBookings):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("DeleteUser error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		}
		return
	}

	c.Status(http.StatusNoContent)
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
		User:  dto.NewUserResponse(user),
	})
}
