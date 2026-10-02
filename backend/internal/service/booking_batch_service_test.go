package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"H1-canchas/internal/domain"
)

// mockBookingBatchRepo implements BookingBatchRepo for unit tests.
type mockBookingBatchRepo struct {
	createFn         func(ctx context.Context, b *domain.BookingBatch) (int64, error)
	getByIDFn        func(ctx context.Context, id int64) (*domain.BookingBatch, error)
	getAllDetailedFn func(ctx context.Context, batchType *string) ([]domain.BookingBatchDetail, error)
	updateStatusFn   func(ctx context.Context, id int64, status string) error
	createCalls      int
	updatedStatus    string
}

func (m *mockBookingBatchRepo) Create(ctx context.Context, b *domain.BookingBatch) (int64, error) {
	m.createCalls++
	if m.createFn != nil {
		return m.createFn(ctx, b)
	}
	return 100, nil
}
func (m *mockBookingBatchRepo) GetByID(ctx context.Context, id int64) (*domain.BookingBatch, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockBookingBatchRepo) GetAllDetailed(ctx context.Context, batchType *string) ([]domain.BookingBatchDetail, error) {
	if m.getAllDetailedFn != nil {
		return m.getAllDetailedFn(ctx, batchType)
	}
	return nil, nil
}
func (m *mockBookingBatchRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	m.updatedStatus = status
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

// mockBookingRepoForBatch implements BookingRepoForBatch for unit tests.
type mockBookingRepoForBatch struct {
	setBatchIDFn         func(ctx context.Context, bookingID, batchID int64) error
	cancelFutureByBatchFn func(ctx context.Context, batchID int64, fromDate time.Time) (int64, error)
	setBatchIDCalls      int
	cancelFutureCalled   bool
}

func (m *mockBookingRepoForBatch) SetBatchID(ctx context.Context, bookingID, batchID int64) error {
	m.setBatchIDCalls++
	if m.setBatchIDFn != nil {
		return m.setBatchIDFn(ctx, bookingID, batchID)
	}
	return nil
}
func (m *mockBookingRepoForBatch) CancelFutureByBatch(ctx context.Context, batchID int64, fromDate time.Time) (int64, error) {
	m.cancelFutureCalled = true
	if m.cancelFutureByBatchFn != nil {
		return m.cancelFutureByBatchFn(ctx, batchID, fromDate)
	}
	return 0, nil
}

// mockSlotLister implements spaceSlotLister for unit tests.
type mockSlotLister struct {
	getSlotsBySpaceIDFn func(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error)
}

func (m *mockSlotLister) GetSlotsBySpaceID(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error) {
	if m.getSlotsBySpaceIDFn != nil {
		return m.getSlotsBySpaceIDFn(ctx, spaceID)
	}
	return nil, nil
}

// daysFromNow devuelve la fecha de hoy (Argentina) + n días, en el mismo
// formato que usa el resto del código (medianoche UTC = fecha calendario).
func daysFromNow(n int) time.Time {
	return argToday().AddDate(0, 0, n)
}

func newBookingBatchSvc(
	batchRepo *mockBookingBatchRepo, repoForBatch *mockBookingRepoForBatch,
	slotLister *mockSlotLister, bookingRepo *mockBookingRepo, spaceRepo *mockSpaceRepoForBooking,
) *BookingBatchService {
	bookingSvc := newBookingSvc(bookingRepo, spaceRepo)
	return NewBookingBatchService(batchRepo, repoForBatch, slotLister, bookingSvc)
}

// --- CreateRecurringTeacherBatch: validaciones ---

func TestBookingBatchService_CreateRecurring_EmptyName(t *testing.T) {
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	_, _, _, err := svc.CreateRecurringTeacherBatch(context.Background(), 1, 1, 1, 1, "", "1155551234", daysFromNow(1), daysFromNow(30), "")
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("got %v, want ErrNameRequired", err)
	}
}

func TestBookingBatchService_CreateRecurring_InvalidWeekday(t *testing.T) {
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	_, _, _, err := svc.CreateRecurringTeacherBatch(context.Background(), 1, 1, 1, 7, "Prof", "1155551234", daysFromNow(1), daysFromNow(30), "")
	if !errors.Is(err, ErrInvalidWeekday) {
		t.Errorf("got %v, want ErrInvalidWeekday", err)
	}
}

func TestBookingBatchService_CreateRecurring_EndBeforeStart(t *testing.T) {
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	weekday := int(daysFromNow(10).Weekday())
	_, _, _, err := svc.CreateRecurringTeacherBatch(context.Background(), 1, 1, 1, weekday, "Prof", "1155551234", daysFromNow(10), daysFromNow(1), "")
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Errorf("got %v, want ErrInvalidDateRange", err)
	}
}

func TestBookingBatchService_CreateRecurring_StartInPast(t *testing.T) {
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	weekday := int(daysFromNow(-5).Weekday())
	_, _, _, err := svc.CreateRecurringTeacherBatch(context.Background(), 1, 1, 1, weekday, "Prof", "1155551234", daysFromNow(-5), daysFromNow(30), "")
	if !errors.Is(err, ErrInvalidBookingDate) {
		t.Errorf("got %v, want ErrInvalidBookingDate", err)
	}
}

func TestBookingBatchService_CreateRecurring_RangeTooLong(t *testing.T) {
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	weekday := int(daysFromNow(1).Weekday())
	_, _, _, err := svc.CreateRecurringTeacherBatch(context.Background(), 1, 1, 1, weekday, "Prof", "1155551234", daysFromNow(1), daysFromNow(400), "")
	if !errors.Is(err, ErrDateRangeTooLong) {
		t.Errorf("got %v, want ErrDateRangeTooLong", err)
	}
}

// --- CreateRecurringTeacherBatch: happy path + skip on conflict ---

func TestBookingBatchService_CreateRecurring_HappyPath(t *testing.T) {
	batchRepo := &mockBookingBatchRepo{}
	repoForBatch := &mockBookingRepoForBatch{}
	bookingRepo := &mockBookingRepo{}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Space, error) {
			return activeSpaceWithPrice(10000), nil
		},
	}
	svc := newBookingBatchSvc(batchRepo, repoForBatch, &mockSlotLister{}, bookingRepo, spaceRepo)

	start := daysFromNow(1)
	weekday := int(start.Weekday())
	end := start.AddDate(0, 0, 14) // 3 ocurrencias: día 0, 7 y 14

	batch, created, skipped, err := svc.CreateRecurringTeacherBatch(
		context.Background(), 1, 1, 1, weekday, "Prof. Fernández", "1155551234", start, end, "",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 3 {
		t.Errorf("got %d created dates, want 3", len(created))
	}
	if len(skipped) != 0 {
		t.Errorf("got %d skipped dates, want 0", len(skipped))
	}
	if batchRepo.createCalls != 1 {
		t.Errorf("batch Create called %d times, want 1 (una sola receta para todas las fechas)", batchRepo.createCalls)
	}
	if repoForBatch.setBatchIDCalls != 3 {
		t.Errorf("SetBatchID called %d times, want 3", repoForBatch.setBatchIDCalls)
	}
	if batch.Type != domain.BatchTypeRecurringTeacher {
		t.Errorf("got batch type %q, want %q", batch.Type, domain.BatchTypeRecurringTeacher)
	}
}

func TestBookingBatchService_CreateRecurring_SkipsConflictingDate(t *testing.T) {
	start := daysFromNow(1)
	weekday := int(start.Weekday())
	conflictingDate := start.AddDate(0, 0, 7) // la segunda ocurrencia

	bookingRepo := &mockBookingRepo{
		existsActiveBookingFn: func(ctx context.Context, spaceID, slotID int64, date time.Time) (bool, error) {
			return date.Equal(conflictingDate), nil
		},
	}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Space, error) {
			return activeSpaceWithPrice(10000), nil
		},
	}
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, bookingRepo, spaceRepo)

	end := start.AddDate(0, 0, 14)
	_, created, skipped, err := svc.CreateRecurringTeacherBatch(
		context.Background(), 1, 1, 1, weekday, "Prof. Fernández", "1155551234", start, end, "",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 2 {
		t.Errorf("got %d created dates, want 2", len(created))
	}
	if len(skipped) != 1 || !skipped[0].Equal(conflictingDate) {
		t.Errorf("got skipped=%v, want [%v]", skipped, conflictingDate)
	}
}

// --- CreateMaintenanceBatch ---

func TestBookingBatchService_CreateMaintenance_ReasonRequired(t *testing.T) {
	svc := newBookingBatchSvc(&mockBookingBatchRepo{}, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	_, _, _, err := svc.CreateMaintenanceBatch(context.Background(), 1, 1, int64ptr(1), daysFromNow(1), daysFromNow(3), "  ")
	if !errors.Is(err, ErrReasonRequired) {
		t.Errorf("got %v, want ErrReasonRequired", err)
	}
}

func TestBookingBatchService_CreateMaintenance_AllSlotsWhenSlotIDNil(t *testing.T) {
	batchRepo := &mockBookingBatchRepo{}
	repoForBatch := &mockBookingRepoForBatch{}
	bookingRepo := &mockBookingRepo{}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(ctx context.Context, id int64) (*domain.Space, error) {
			return activeSpaceWithPrice(10000), nil
		},
	}
	slotLister := &mockSlotLister{
		getSlotsBySpaceIDFn: func(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error) {
			return []domain.SpaceSlot{{ID: 10}, {ID: 11}}, nil
		},
	}
	svc := newBookingBatchSvc(batchRepo, repoForBatch, slotLister, bookingRepo, spaceRepo)

	start := daysFromNow(1)
	end := start.AddDate(0, 0, 2) // 3 días × 2 turnos = 6 reservas

	batch, created, skipped, err := svc.CreateMaintenanceBatch(context.Background(), 1, 1, nil, start, end, "Cambio de red")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 6 {
		t.Errorf("got %d created, want 6 (3 días x 2 turnos)", len(created))
	}
	if len(skipped) != 0 {
		t.Errorf("got %d skipped, want 0", len(skipped))
	}
	if batchRepo.createCalls != 1 {
		t.Errorf("batch Create called %d times, want 1", batchRepo.createCalls)
	}
	if batch.Type != domain.BatchTypeMaintenance {
		t.Errorf("got batch type %q, want %q", batch.Type, domain.BatchTypeMaintenance)
	}
}

// --- CancelBatch ---

func TestBookingBatchService_CancelBatch_NotFound(t *testing.T) {
	batchRepo := &mockBookingBatchRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.BookingBatch, error) { return nil, nil },
	}
	svc := newBookingBatchSvc(batchRepo, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	err := svc.CancelBatch(context.Background(), 1)
	if !errors.Is(err, ErrBatchNotFound) {
		t.Errorf("got %v, want ErrBatchNotFound", err)
	}
}

func TestBookingBatchService_CancelBatch_AlreadyCancelled(t *testing.T) {
	batchRepo := &mockBookingBatchRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.BookingBatch, error) {
			return &domain.BookingBatch{ID: 1, Status: domain.BatchStatusCancelled}, nil
		},
	}
	svc := newBookingBatchSvc(batchRepo, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})
	err := svc.CancelBatch(context.Background(), 1)
	if !errors.Is(err, ErrBatchAlreadyCancelled) {
		t.Errorf("got %v, want ErrBatchAlreadyCancelled", err)
	}
}

func TestBookingBatchService_CancelBatch_HappyPath(t *testing.T) {
	batchRepo := &mockBookingBatchRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.BookingBatch, error) {
			return &domain.BookingBatch{ID: 1, Status: domain.BatchStatusActive}, nil
		},
	}
	repoForBatch := &mockBookingRepoForBatch{}
	svc := newBookingBatchSvc(batchRepo, repoForBatch, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})

	if err := svc.CancelBatch(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repoForBatch.cancelFutureCalled {
		t.Error("CancelFutureByBatch no fue llamado")
	}
	if batchRepo.updatedStatus != domain.BatchStatusCancelled {
		t.Errorf("got updated status %q, want %q", batchRepo.updatedStatus, domain.BatchStatusCancelled)
	}
}

// --- helpers de fechas ---

func TestDatesMatchingWeekday(t *testing.T) {
	start := daysFromNow(1)
	weekday := int(start.Weekday())
	end := start.AddDate(0, 0, 14)

	dates := datesMatchingWeekday(start, end, weekday)
	if len(dates) != 3 {
		t.Fatalf("got %d dates, want 3", len(dates))
	}
	for _, d := range dates {
		if int(d.Weekday()) != weekday {
			t.Errorf("date %v has weekday %d, want %d", d, d.Weekday(), weekday)
		}
	}
}

func TestAllDatesInRange(t *testing.T) {
	start := daysFromNow(1)
	end := start.AddDate(0, 0, 4)
	dates := allDatesInRange(start, end)
	if len(dates) != 5 {
		t.Fatalf("got %d dates, want 5", len(dates))
	}
}

func TestBookingBatchService_CreateRecurring_InvalidPhoneFailsBeforeCreatingBatch(t *testing.T) {
	batchRepo := &mockBookingBatchRepo{}
	svc := newBookingBatchSvc(batchRepo, &mockBookingRepoForBatch{}, &mockSlotLister{}, &mockBookingRepo{}, &mockSpaceRepoForBooking{})

	_, _, _, err := svc.CreateRecurringTeacherBatch(context.Background(), 1, 1, 1, 1, "Prof", "abc", daysFromNow(1), daysFromNow(30), "")
	if !errors.Is(err, ErrInvalidPhone) {
		t.Errorf("got %v, want ErrInvalidPhone", err)
	}
	if batchRepo.createCalls != 0 {
		t.Errorf("batch Create called %d times, want 0", batchRepo.createCalls)
	}
}
