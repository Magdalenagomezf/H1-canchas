package dto

// RegisterRequest es el JSON que manda el cliente para registrarse.
// binding:"required" hace que ShouldBindJSON devuelva error si falta el campo.
type RegisterRequest struct {
	Name     string `json:"name"     binding:"required"`
	Phone    string `json:"phone"    binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"` // opcional, sin binding:"required"
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
