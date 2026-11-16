package dto

import "H1-canchas/internal/domain"

// RegisterRequest es el JSON que manda el cliente para registrarse.
// binding:"required" hace que ShouldBindJSON devuelva error si falta el campo.
type RegisterRequest struct {
	Name     string `json:"name"     binding:"required"`
	Phone    string `json:"phone"    binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"    binding:"required"`
}

// LoginRequest es el JSON que manda el cliente para iniciar sesión.
type LoginRequest struct {
	Phone    string `json:"phone"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateStaffRequest es el JSON que manda el admin para crear un usuario staff.
type CreateStaffRequest struct {
	Name     string `json:"name"     binding:"required"`
	Phone    string `json:"phone"    binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"     binding:"required"` // "receptionist" o "admin"
	Email    string `json:"email"`                       // opcional
}

// UserResponse es el subconjunto de datos del usuario que se expone en la API.
// No incluye password_hash ni campos internos.
type UserResponse struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Phone *string `json:"phone"` // null (never omitted) when the account has no phone yet
	Email *string `json:"email,omitempty"`
	Role  string  `json:"role"`
}

// NewUserResponse builds the public view of a user.
func NewUserResponse(u *domain.User) UserResponse {
	return UserResponse{ID: u.ID, Name: u.Name, Phone: u.Phone, Email: u.Email, Role: u.Role}
}

// AuthResponse es lo que devuelven /auth/login y /auth/register.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// GoogleLoginRequest is the JSON for POST /auth/google: the ID token that
// Google Sign-In hands to the browser.
type GoogleLoginRequest struct {
	Credential string `json:"credential" binding:"required"`
}

// UpdateProfileRequest is the JSON for PATCH /me. A missing field (nil
// pointer) means "leave unchanged".
type UpdateProfileRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

// UpdateRoleRequest es el JSON para cambiar el rol de un usuario.
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// PasswordResetRequest is the JSON for POST /auth/password-reset.
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required"`
}

// PasswordResetConfirmRequest is the JSON for POST /auth/password-reset/confirm.
// Both fields are checked by the service, so a missing one gets the same
// answer as an invalid one.
type PasswordResetConfirmRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// PasswordResetCheckRequest is the JSON for POST /auth/password-reset/check.
type PasswordResetCheckRequest struct {
	Token string `json:"token"`
}
