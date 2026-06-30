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

// UserResponse es el subconjunto de datos del usuario que se expone en la API.
// No incluye password_hash ni campos internos.
type UserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

// AuthResponse es lo que devuelven /auth/login y /auth/register.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UpdateRoleRequest es el JSON para cambiar el rol de un usuario.
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
