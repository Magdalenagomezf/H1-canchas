package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // embebe la base de zonas horarias: LoadLocation funciona aunque el host no tenga tzdata instalado (ej. contenedores slim)

	"H1-canchas/internal/domain"
)

const argentinaTZ = "America/Argentina/Buenos_Aires"

// maxBatchRange es el máximo rango de fechas permitido para un turno fijo
// o un bloqueo de mantenimiento. Sin este límite, "fecha de fin" indefinida
// requeriría un motor de reglas recurrentes evaluado en tiempo real (y un
// cron para mantenerlo) en vez de materializar bookings reales — ver
// decisión registrada en la conversación con el usuario.
const maxBatchRangeDays = 366

// BookingBatchRepo es lo que el service necesita de la tabla booking_batches.
type BookingBatchRepo interface {
	Create(ctx context.Context, b *domain.BookingBatch) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.BookingBatch, error)
	GetAllDetailed(ctx context.Context, batchType *string) ([]domain.BookingBatchDetail, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

// BookingRepoForBatch es lo que el service necesita de bookings para
// etiquetar y cancelar en bloque las reservas generadas por un batch.
type BookingRepoForBatch interface {
	SetBatchID(ctx context.Context, bookingID, batchID int64) error
	CancelFutureByBatch(ctx context.Context, batchID int64, fromDate time.Time) (int64, error)
}

// spaceSlotLister es lo mínimo que el service necesita de spaces para
// resolver "todos los turnos activos" cuando un bloqueo de mantenimiento
// no especifica un slot puntual.
type spaceSlotLister interface {
	GetSlotsBySpaceID(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error)
}

type BookingBatchService struct {
	repo        BookingBatchRepo
	bookingRepo BookingRepoForBatch
	slotLister  spaceSlotLister
	bookingSvc  *BookingService
}

func NewBookingBatchService(repo BookingBatchRepo, bookingRepo BookingRepoForBatch, slotLister spaceSlotLister, bookingSvc *BookingService) *BookingBatchService {
	return &BookingBatchService{repo: repo, bookingRepo: bookingRepo, slotLister: slotLister, bookingSvc: bookingSvc}
}

// argToday devuelve la fecha de hoy en horario de Argentina, representada
// como medianoche UTC — la misma convención que usa el resto del código
// para booking_date (que sale de time.Parse("2006-01-02", ...), sin
// componente horario real). Da igual qué hora es en UTC: lo que importa
// es qué día calendario es en Buenos Aires.
func argToday() time.Time {
	loc, err := time.LoadLocation(argentinaTZ)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// CreateRecurringTeacherBatch reserva un turno fijo semanal (mismo día de
// la semana, mismo horario) para un profesor, entre startDate y endDate.
// Genera una reserva 'confirmed' real por cada fecha que coincide con el
// día de la semana pedido; las fechas que ya estaban ocupadas se saltean
// y se reportan, no abortan el resto del batch.
func (s *BookingBatchService) CreateRecurringTeacherBatch(
	ctx context.Context, createdBy int64, spaceID, slotID int64, weekday int,
	customerName, customerPhone string, startDate, endDate time.Time, note string,
) (*domain.BookingBatch, []time.Time, []time.Time, error) {
	customerName = strings.TrimSpace(customerName)
	customerPhone = strings.TrimSpace(customerPhone)
	if customerName == "" {
		return nil, nil, nil, ErrNameRequired
	}
	if customerPhone == "" {
		return nil, nil, nil, ErrPhoneRequired
	}
	if weekday < 0 || weekday > 6 {
		return nil, nil, nil, ErrInvalidWeekday
	}
	if err := validateDateRange(startDate, endDate); err != nil {
		return nil, nil, nil, err
	}

	reason := strings.TrimSpace(note)
	if reason == "" {
		reason = fmt.Sprintf("Turno fijo — %s", customerName)
	}

	wd := weekday
	batch := &domain.BookingBatch{
		Type:      domain.BatchTypeRecurringTeacher,
		SpaceID:   spaceID,
		SlotID:    &slotID,
		Weekday:   &wd,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
		CreatedBy: createdBy,
		Status:    domain.BatchStatusActive,
	}

	dates := datesMatchingWeekday(startDate, endDate, weekday)

	created, skipped, err := s.materializeBookings(ctx, batch, createdBy, spaceID, slotID, dates, customerName, customerPhone)
	if err != nil {
		return nil, nil, nil, err
	}
	return batch, created, skipped, nil
}

// CreateMaintenanceBatch bloquea uno o todos los turnos activos de un
// espacio durante un rango de días (ej. cancha en refacción). Igual que
// el turno fijo, genera reservas 'confirmed' reales por cada combinación
// fecha × turno, con el nombre del recepcionista y el motivo en vez de
// un cliente real.
func (s *BookingBatchService) CreateMaintenanceBatch(
	ctx context.Context, createdBy int64, spaceID int64, slotID *int64,
	startDate, endDate time.Time, reason string,
) (*domain.BookingBatch, []time.Time, []time.Time, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, nil, nil, ErrReasonRequired
	}
	if err := validateDateRange(startDate, endDate); err != nil {
		return nil, nil, nil, err
	}

	slotIDs, err := s.resolveSlotIDs(ctx, spaceID, slotID)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(slotIDs) == 0 {
		return nil, nil, nil, ErrNotFound
	}

	batch := &domain.BookingBatch{
		Type:      domain.BatchTypeMaintenance,
		SpaceID:   spaceID,
		SlotID:    slotID,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
		CreatedBy: createdBy,
		Status:    domain.BatchStatusActive,
	}

	dates := allDatesInRange(startDate, endDate)

	var created, skipped []time.Time
	for _, slID := range slotIDs {
		c, sk, err := s.materializeBookings(ctx, batch, createdBy, spaceID, slID, dates, "Mantenimiento", "-")
		if err != nil {
			return nil, nil, nil, err
		}
		created = append(created, c...)
		skipped = append(skipped, sk...)
	}
	return batch, created, skipped, nil
}

// materializeBookings inserta la fila del batch (una sola vez, la primera
// vez que se llama con batch.ID == 0) y luego una reserva manual real por
// cada fecha, etiquetada con el batch_id. Las fechas ya ocupadas
// (ErrSlotNotAvailable) se saltean sin abortar el resto.
func (s *BookingBatchService) materializeBookings(
	ctx context.Context, batch *domain.BookingBatch, createdBy, spaceID, slotID int64,
	dates []time.Time, customerName, customerPhone string,
) ([]time.Time, []time.Time, error) {
	if batch.ID == 0 {
		id, err := s.repo.Create(ctx, batch)
		if err != nil {
			return nil, nil, fmt.Errorf("BookingBatchService.materializeBookings: creando batch: %w", err)
		}
		batch.ID = id
	}

	var created, skipped []time.Time
	for _, date := range dates {
		booking, err := s.bookingSvc.CreateManual(ctx, createdBy, spaceID, slotID, date, customerName, customerPhone)
		if err != nil {
			if errors.Is(err, ErrSlotNotAvailable) {
				skipped = append(skipped, date)
				continue
			}
			return nil, nil, fmt.Errorf("BookingBatchService.materializeBookings: %w", err)
		}
		if err := s.bookingRepo.SetBatchID(ctx, booking.ID, batch.ID); err != nil {
			return nil, nil, fmt.Errorf("BookingBatchService.materializeBookings: etiquetando batch: %w", err)
		}
		created = append(created, date)
	}
	return created, skipped, nil
}

// resolveSlotIDs devuelve [*slotID] si vino un slot puntual, o todos los
// turnos activos del espacio si slotID es nil.
func (s *BookingBatchService) resolveSlotIDs(ctx context.Context, spaceID int64, slotID *int64) ([]int64, error) {
	if slotID != nil {
		return []int64{*slotID}, nil
	}
	slots, err := s.slotLister.GetSlotsBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("BookingBatchService.resolveSlotIDs: %w", err)
	}
	ids := make([]int64, len(slots))
	for i, sl := range slots {
		ids[i] = sl.ID
	}
	return ids, nil
}

// CancelBatch cancela todas las ocurrencias futuras (>= hoy en Argentina)
// de un turno fijo o bloqueo, y marca el batch como cancelado. Las
// ocurrencias pasadas quedan como historial, sin tocar.
func (s *BookingBatchService) CancelBatch(ctx context.Context, batchID int64) error {
	batch, err := s.repo.GetByID(ctx, batchID)
	if err != nil {
		return fmt.Errorf("BookingBatchService.CancelBatch: %w", err)
	}
	if batch == nil {
		return ErrBatchNotFound
	}
	if batch.Status == domain.BatchStatusCancelled {
		return ErrBatchAlreadyCancelled
	}

	if _, err := s.bookingRepo.CancelFutureByBatch(ctx, batchID, argToday()); err != nil {
		return fmt.Errorf("BookingBatchService.CancelBatch: cancelando reservas: %w", err)
	}
	if err := s.repo.UpdateStatus(ctx, batchID, domain.BatchStatusCancelled); err != nil {
		return fmt.Errorf("BookingBatchService.CancelBatch: %w", err)
	}
	return nil
}

// GetAll devuelve los batches (turnos fijos y/o bloqueos), opcionalmente
// filtrados por tipo.
func (s *BookingBatchService) GetAll(ctx context.Context, batchType *string) ([]domain.BookingBatchDetail, error) {
	rows, err := s.repo.GetAllDetailed(ctx, batchType)
	if err != nil {
		return nil, fmt.Errorf("BookingBatchService.GetAll: %w", err)
	}
	return rows, nil
}

func validateDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return ErrInvalidDateRange
	}
	if startDate.Before(argToday()) {
		return ErrInvalidBookingDate
	}
	if endDate.Sub(startDate) > maxBatchRangeDays*24*time.Hour {
		return ErrDateRangeTooLong
	}
	return nil
}

// datesMatchingWeekday devuelve todas las fechas entre startDate y endDate
// (ambas inclusive) cuyo día de la semana coincide con weekday.
func datesMatchingWeekday(startDate, endDate time.Time, weekday int) []time.Time {
	var dates []time.Time
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		if int(d.Weekday()) == weekday {
			dates = append(dates, d)
		}
	}
	return dates
}

// allDatesInRange devuelve cada fecha entre startDate y endDate, ambas inclusive.
func allDatesInRange(startDate, endDate time.Time) []time.Time {
	var dates []time.Time
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}
