package domain

import "time"

// User representa tanto clientes como recepcionistas/admin.
// El rol determina qué puede hacer en el sistema.
type User struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Email        *string   `db:"email"`
	Phone        string    `db:"phone"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
}

// Roles disponibles. Usamos constantes para no tener strings mágicos
// dispersos por todo el código.
const (
	RoleCustomer     = "customer"
	RoleReceptionist = "receptionist"
	RoleAdmin        = "admin"
)
