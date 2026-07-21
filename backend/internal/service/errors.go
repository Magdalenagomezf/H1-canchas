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
	ErrInvalidRole       = errors.New("rol inválido")

	ErrSlotLabelRequired       = errors.New("el label del slot es obligatorio")
	ErrSlotNotAvailable        = errors.New("el turno no está disponible")
	ErrSlotDoesNotBelong       = errors.New("el slot no pertenece al espacio indicado")
	ErrBookingAlreadyCancelled = errors.New("la reserva ya está cancelada")
	ErrBookingAlreadyCompleted = errors.New("la reserva ya está finalizada")
	ErrInvalidBookingID        = errors.New("id de reserva inválido")
	ErrInvalidSlotID           = errors.New("id de slot inválido")
	ErrInvalidBookingDate      = errors.New("fecha de reserva inválida")

	// delete guards
	ErrUserHasBookings  = errors.New("el usuario tiene reservas asociadas y no puede eliminarse")
	ErrSpaceHasBookings = errors.New("el espacio tiene reservas asociadas y no puede eliminarse")
	ErrCannotDeleteSelf = errors.New("no podés eliminar tu propia cuenta")

	// pagos
	ErrBookingExpired                  = errors.New("la reserva expiró y el slot ya no está disponible")
	ErrBookingNotPayable               = errors.New("la reserva no admite pagos en su estado actual")
	ErrPaymentAlreadyCompleted         = errors.New("ese tramo de la reserva ya está pagado")
	ErrInvalidPaymentKind              = errors.New("tramo de pago inválido")
	ErrInvalidPaymentMethod            = errors.New("método de pago inválido")
	ErrInvalidPaymentStatus            = errors.New("estado de pago inválido")
	ErrCancelRequiresStaffAfterPayment = errors.New("una vez pagada la seña, solo el staff puede cancelar la reserva")
	ErrMercadoPagoUnavailable          = errors.New("mercado pago no está disponible en este momento")

	// turnos fijos y bloqueos de mantenimiento
	ErrInvalidWeekday        = errors.New("día de la semana inválido")
	ErrInvalidDateRange      = errors.New("el rango de fechas es inválido")
	ErrDateRangeTooLong      = errors.New("el rango de fechas no puede superar 1 año")
	ErrReasonRequired        = errors.New("el motivo es obligatorio")
	ErrBatchNotFound         = errors.New("turno fijo o bloqueo no encontrado")
	ErrBatchAlreadyCancelled = errors.New("el turno fijo o bloqueo ya está cancelado")
)
