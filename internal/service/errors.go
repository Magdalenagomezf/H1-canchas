// internal/service/errors.go
package service

import "errors"

var (
	// auth
	ErrPhoneAlreadyExists = errors.New("el teléfono ya está registrado")
	ErrInvalidCredentials = errors.New("credenciales incorrectas")
	ErrUserInactive       = errors.New("usuario inactivo")

	// validaciones de usuario
	ErrNameRequired     = errors.New("el nombre es obligatorio")
	ErrPhoneRequired    = errors.New("el teléfono es obligatorio")
	ErrPasswordRequired = errors.New("la contraseña es obligatoria")
	ErrPasswordTooShort = errors.New("la contraseña debe tener al menos 6 caracteres")

	// genéricos — cualquier service los puede usar
	ErrNotFound     = errors.New("recurso no encontrado")
	ErrUnauthorized = errors.New("no autorizado")

	// spaces
	ErrSpaceNameRequired = errors.New("el nombre del espacio es obligatorio")
	ErrInvalidSpaceType  = errors.New("tipo de espacio inválido")
	ErrInvalidSpacePrice = errors.New("el precio del espacio debe ser mayor a cero")
	ErrInvalidSpaceID    = errors.New("id de espacio inválido")

	ErrSlotNotAvailable        = errors.New("el turno no está disponible")
	ErrSlotDoesNotBelong       = errors.New("el slot no pertenece al espacio indicado")
	ErrBookingAlreadyCancelled = errors.New("la reserva ya está cancelada")
	ErrBookingAlreadyCompleted = errors.New("la reserva ya está finalizada")
	ErrInvalidBookingID        = errors.New("id de reserva inválido")
	ErrInvalidSlotID           = errors.New("id de slot inválido")
	ErrInvalidBookingDate      = errors.New("fecha de reserva inválida")
	// bookings — los agregás cuando llegues a ese service
	// ErrSlotNotAvailable = errors.New("el turno no está disponible")
	// ErrBookingNotFound  = errors.New("reserva no encontrada")
)
