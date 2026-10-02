package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/database"
	"H1-canchas/pkg/phone"
)

// BookingRepo es la interfaz que el service necesita.
// La define el service, la implementa el repository.
type BookingRepo interface {
	BeginTx(ctx context.Context) (database.Tx, error)
	ExistsActiveBooking(ctx context.Context, spaceID, slotID int64, date time.Time) (bool, error)
	SlotBelongsToSpace(ctx context.Context, spaceID, slotID int64) (bool, error)
	Create(ctx context.Context, tx database.Tx, b *domain.Booking) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id int64, status string) (bool, error)
	GetAllWithDetails(ctx context.Context, userID *int64, date *time.Time) ([]domain.BookingDetail, error)
	GetByIDWithDetails(ctx context.Context, id int64) (*domain.BookingDetail, error)
	ExpirePendingBookings(ctx context.Context, spaceID, slotID int64, date time.Time) (int64, error)
}

// SpaceRepoForBooking es lo mínimo que el booking service
// necesita saber de spaces: solo el precio.
type SpaceRepoForBooking interface {
	GetByID(ctx context.Context, id int64) (*domain.Space, error)
}

type BookingService struct {
	repo       BookingRepo
	spaceRepo  SpaceRepoForBooking
	holdTTL    time.Duration
	depositPct float64
}

func NewBookingService(repo BookingRepo, spaceRepo SpaceRepoForBooking, holdTTL time.Duration, depositPct float64) *BookingService {
	return &BookingService{repo: repo, spaceRepo: spaceRepo, holdTTL: holdTTL, depositPct: depositPct}
}

// depositAndBalance reparte el precio total en seña + saldo, redondeando
// la seña a centavos primero para que la suma de ambos tramos nunca
// difiera del total por errores de punto flotante.
func (s *BookingService) depositAndBalance(totalPrice float64) (deposit, balance float64) {
	deposit = math.Round(totalPrice*s.depositPct*100) / 100
	balance = totalPrice - deposit
	return deposit, balance
}

// Create crea una reserva para un usuario registrado.
// HACER MAS ADELANTE= Usa transacción + FOR UPDATE para evitar doble reserva.
func (s *BookingService) Create(ctx context.Context, customerUserID int64, spaceID, slotID int64, date time.Time) (*domain.Booking, error) {
	if date.Before(argToday()) {
		return nil, ErrInvalidBookingDate
	}

	space, err := s.spaceRepo.GetByID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("BookingService.Create: %w", err)
	}
	if space == nil || !space.IsActive {
		return nil, ErrNotFound
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("BookingService.Create: abriendo tx: %w", err)
	}
	// defer rollback: si algo falla después, deshace todo.
	// Si el commit ya se hizo, el rollback no tiene efecto.
	defer tx.Rollback()

	belongs, err := s.repo.SlotBelongsToSpace(ctx, spaceID, slotID)
	if err != nil {
		return nil, fmt.Errorf("BookingService.Create: %w", err)
	}
	if !belongs {
		return nil, ErrSlotDoesNotBelong
	}

	// Libera holds vencidos de este space+slot+fecha antes de chequear
	// disponibilidad, para que una reserva expirada no bloquee una nueva.
	if _, err := s.repo.ExpirePendingBookings(ctx, spaceID, slotID, date); err != nil {
		return nil, fmt.Errorf("BookingService.Create: liberando holds vencidos: %w", err)
	}

	exists, err := s.repo.ExistsActiveBooking(ctx, spaceID, slotID, date)
	if err != nil {
		return nil, fmt.Errorf("BookingService.Create: validando disponibilidad: %w", err)
	}
	if exists {
		return nil, ErrSlotNotAvailable
	}

	depositAmount, balanceAmount := s.depositAndBalance(space.PricePerSlot)
	expiresAt := time.Now().UTC().Add(s.holdTTL)

	booking := &domain.Booking{
		CustomerUserID: &customerUserID,
		CreatedBy:      customerUserID,
		SpaceID:        spaceID,
		SlotID:         slotID,
		BookingDate:    date,
		Status:         domain.BookingStatusPending,
		TotalPrice:     space.PricePerSlot,
		DepositAmount:  depositAmount,
		DepositStatus:  domain.PaymentStatusUnpaid,
		ExpiresAt:      &expiresAt,
		BalanceAmount:  balanceAmount,
		BalanceStatus:  domain.PaymentStatusUnpaid,
	}

	id, err := s.repo.Create(ctx, tx, booking)
	if err != nil {
		return nil, fmt.Errorf("BookingService.Create: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("BookingService.Create: commit: %w", err)
	}

	booking.ID = id
	return booking, nil
}

// CreateManual crea una reserva manual hecha por el recepcionista.
// El cliente puede no tener usuario, solo nombre y teléfono.
// El teléfono se guarda normalizado (E.164).
func (s *BookingService) CreateManual(ctx context.Context, createdBy int64, spaceID, slotID int64, date time.Time, customerName, customerPhone string) (*domain.Booking, error) {
	customerName = strings.TrimSpace(customerName)
	customerPhone = strings.TrimSpace(customerPhone)

	if customerName == "" {
		return nil, ErrNameRequired
	}
	if customerPhone == "" {
		return nil, ErrPhoneRequired
	}

	normalizedPhone, err := phone.Normalize(customerPhone)
	if err != nil {
		return nil, ErrInvalidPhone
	}

	return s.createManual(ctx, createdBy, spaceID, slotID, date, customerName, normalizedPhone)
}

// createManual inserta la reserva manual. Espera nombre y teléfono ya
// validados: el teléfono llega normalizado, o es el placeholder "-" de los
// bloqueos de mantenimiento, que no tienen un cliente real.
func (s *BookingService) createManual(ctx context.Context, createdBy int64, spaceID, slotID int64, date time.Time, customerName, customerPhone string) (*domain.Booking, error) {
	if date.Before(argToday()) {
		return nil, ErrInvalidBookingDate
	}

	space, err := s.spaceRepo.GetByID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("BookingService.CreateManual: %w", err)
	}
	if space == nil || !space.IsActive {
		return nil, ErrNotFound
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("BookingService.CreateManual: abriendo tx: %w", err)
	}
	defer tx.Rollback()

	belongs, err := s.repo.SlotBelongsToSpace(ctx, spaceID, slotID)
	if err != nil {
		return nil, fmt.Errorf("BookingService.CreateManual: validando slot: %w", err)
	}
	if !belongs {
		return nil, ErrSlotDoesNotBelong
	}

	exists, err := s.repo.ExistsActiveBooking(ctx, spaceID, slotID, date)
	if err != nil {
		return nil, fmt.Errorf("BookingService.CreateManual: validando disponibilidad: %w", err)
	}
	if exists {
		return nil, ErrSlotNotAvailable
	}

	depositAmount, balanceAmount := s.depositAndBalance(space.PricePerSlot)

	booking := &domain.Booking{
		CreatedBy:     createdBy,
		CustomerName:  &customerName,
		CustomerPhone: &customerPhone,
		SpaceID:       spaceID,
		SlotID:        slotID,
		BookingDate:   date,
		Status:        domain.BookingStatusConfirmed, // manual va directo a confirmed
		TotalPrice:    space.PricePerSlot,
		DepositAmount: depositAmount,
		DepositStatus: domain.PaymentStatusUnpaid,
		BalanceAmount: balanceAmount,
		BalanceStatus: domain.PaymentStatusUnpaid,
	}

	id, err := s.repo.Create(ctx, tx, booking)
	if err != nil {
		return nil, fmt.Errorf("BookingService.CreateManual: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("BookingService.CreateManual: commit: %w", err)
	}

	booking.ID = id
	return booking, nil
}

// GetMyBookings devuelve las reservas del usuario logueado.
func (s *BookingService) GetMyBookings(ctx context.Context, userID int64) ([]domain.BookingDetail, error) {
	rows, err := s.repo.GetAllWithDetails(ctx, &userID, nil)
	if err != nil {
		return nil, fmt.Errorf("BookingService.GetMyBookings: %w", err)
	}
	return rows, nil
}

// GetAll devuelve todas las reservas. Solo para recepcionista/admin.
func (s *BookingService) GetAll(ctx context.Context, date *time.Time) ([]domain.BookingDetail, error) {
	rows, err := s.repo.GetAllWithDetails(ctx, nil, date)
	if err != nil {
		return nil, fmt.Errorf("BookingService.GetAll: %w", err)
	}
	return rows, nil
}

// Cancel cancela una reserva.
// Un customer solo puede cancelar las suyas. Recepcionista/admin puede cancelar cualquiera.
func (s *BookingService) Cancel(ctx context.Context, bookingID, requesterID int64, requesterRole string) error {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("BookingService.Cancel: %w", err)
	}
	if booking == nil {
		return ErrNotFound
	}

	// Customer solo puede cancelar sus propias reservas
	isOwner := booking.CustomerUserID != nil && *booking.CustomerUserID == requesterID
	isStaff := requesterRole == domain.RoleReceptionist || requesterRole == domain.RoleAdmin
	if !isOwner && !isStaff {
		return ErrUnauthorized
	}

	if booking.DepositStatus == domain.PaymentStatusPaid && !isStaff {
		return ErrCancelRequiresStaffAfterPayment
	}

	if booking.Status == domain.BookingStatusCancelled {
		return ErrBookingAlreadyCancelled
	}
	if booking.Status == domain.BookingStatusCompleted {
		return ErrBookingAlreadyCompleted
	}

	updated, err := s.repo.UpdateStatus(ctx, bookingID, domain.BookingStatusCancelled)
	if err != nil {
		return fmt.Errorf("BookingService.Cancel: actualizando estado: %w", err)
	}
	if !updated {
		return ErrNotFound
	}

	return nil
}

func (s *BookingService) GetByID(ctx context.Context, bookingID, requesterID int64, requesterRole string) (*domain.BookingDetail, error) {
	row, err := s.repo.GetByIDWithDetails(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("BookingService.GetByID: %w", err)
	}
	if row == nil {
		return nil, ErrNotFound
	}

	isStaff := requesterRole == domain.RoleReceptionist || requesterRole == domain.RoleAdmin
	isOwner := row.CustomerUserID != nil && *row.CustomerUserID == requesterID
	if !isStaff && !isOwner {
		return nil, ErrUnauthorized
	}

	return row, nil
}
