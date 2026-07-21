package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/pkg/database"
)

// mockTx implements database.Tx for unit tests — no real DB needed.
type mockTx struct {
	commitErr   error
	rollbackErr error
}

func (m *mockTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}
func (m *mockTx) Commit() error   { return m.commitErr }
func (m *mockTx) Rollback() error { return m.rollbackErr }

// mockBookingRepo implements BookingRepo for unit tests.
// callLog records call order for tests that assert sequencing
// (e.g. ExpirePendingBookings must run before ExistsActiveBooking).
type mockBookingRepo struct {
	beginTxFn               func(ctx context.Context) (database.Tx, error)
	existsActiveBookingFn   func(ctx context.Context, spaceID, slotID int64, date time.Time) (bool, error)
	slotBelongsToSpaceFn    func(ctx context.Context, spaceID, slotID int64) (bool, error)
	createFn                func(ctx context.Context, tx database.Tx, b *domain.Booking) (int64, error)
	getByIDFn               func(ctx context.Context, id int64) (*domain.Booking, error)
	updateStatusFn          func(ctx context.Context, id int64, status string) (bool, error)
	getAllWithDetailsFn     func(ctx context.Context, userID *int64, date *time.Time) ([]domain.BookingDetail, error)
	getByIDWithDetailsFn    func(ctx context.Context, id int64) (*domain.BookingDetail, error)
	expirePendingBookingsFn func(ctx context.Context, spaceID, slotID int64, date time.Time) (int64, error)
	callLog                 []string
}

func (m *mockBookingRepo) BeginTx(ctx context.Context) (database.Tx, error) {
	if m.beginTxFn != nil {
		return m.beginTxFn(ctx)
	}
	return &mockTx{}, nil
}
func (m *mockBookingRepo) ExistsActiveBooking(ctx context.Context, spaceID, slotID int64, date time.Time) (bool, error) {
	m.callLog = append(m.callLog, "ExistsActiveBooking")
	if m.existsActiveBookingFn != nil {
		return m.existsActiveBookingFn(ctx, spaceID, slotID, date)
	}
	return false, nil
}
func (m *mockBookingRepo) ExpirePendingBookings(ctx context.Context, spaceID, slotID int64, date time.Time) (int64, error) {
	m.callLog = append(m.callLog, "ExpirePendingBookings")
	if m.expirePendingBookingsFn != nil {
		return m.expirePendingBookingsFn(ctx, spaceID, slotID, date)
	}
	return 0, nil
}
func (m *mockBookingRepo) SlotBelongsToSpace(ctx context.Context, spaceID, slotID int64) (bool, error) {
	if m.slotBelongsToSpaceFn != nil {
		return m.slotBelongsToSpaceFn(ctx, spaceID, slotID)
	}
	return true, nil
}
func (m *mockBookingRepo) Create(ctx context.Context, tx database.Tx, b *domain.Booking) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, tx, b)
	}
	return 1, nil
}
func (m *mockBookingRepo) GetByID(ctx context.Context, id int64) (*domain.Booking, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, id int64, status string) (bool, error) {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return true, nil
}
func (m *mockBookingRepo) GetAllWithDetails(ctx context.Context, userID *int64, date *time.Time) ([]domain.BookingDetail, error) {
	if m.getAllWithDetailsFn != nil {
		return m.getAllWithDetailsFn(ctx, userID, date)
	}
	return nil, nil
}
func (m *mockBookingRepo) GetByIDWithDetails(ctx context.Context, id int64) (*domain.BookingDetail, error) {
	if m.getByIDWithDetailsFn != nil {
		return m.getByIDWithDetailsFn(ctx, id)
	}
	return nil, nil
}

// mockSpaceRepoForBooking implements SpaceRepoForBooking.
type mockSpaceRepoForBooking struct {
	getByIDFn func(ctx context.Context, id int64) (*domain.Space, error)
}

func (m *mockSpaceRepoForBooking) GetByID(ctx context.Context, id int64) (*domain.Space, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

// helpers

func tomorrow() time.Time {
	return time.Now().UTC().AddDate(0, 0, 1)
}

func yesterday() time.Time {
	return time.Now().UTC().AddDate(0, 0, -1)
}

func activeSpaceWithPrice(price float64) *domain.Space {
	return &domain.Space{ID: 1, Name: "Cancha", Type: domain.SpaceTypePadel, IsActive: true, PricePerSlot: price}
}

func int64ptr(v int64) *int64 { return &v }

// testHoldTTL / testDepositPct match the values wired in cmd/main.go
// (20 min hold, 15% deposit) so tests exercise realistic numbers.
const (
	testHoldTTL    = 20 * time.Minute
	testDepositPct = 0.15
)

// newBookingSvc is a shortcut to build a BookingService with both mocks.
func newBookingSvc(repo *mockBookingRepo, spaceRepo *mockSpaceRepoForBooking) *BookingService {
	return NewBookingService(repo, spaceRepo, testHoldTTL, testDepositPct)
}

// --- BookingService.Create ---

func TestBookingService_Create_PastDate(t *testing.T) {
	svc := newBookingSvc(&mockBookingRepo{}, &mockSpaceRepoForBooking{})
	_, err := svc.Create(context.Background(), 1, 1, 1, yesterday())
	if !errors.Is(err, ErrInvalidBookingDate) {
		t.Errorf("got %v, want ErrInvalidBookingDate", err)
	}
}

func TestBookingService_Create_SpaceNotFound(t *testing.T) {
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := newBookingSvc(&mockBookingRepo{}, spaceRepo)

	_, err := svc.Create(context.Background(), 1, 1, 1, tomorrow())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestBookingService_Create_SpaceInactive(t *testing.T) {
	inactive := &domain.Space{ID: 1, IsActive: false}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return inactive, nil },
	}
	svc := newBookingSvc(&mockBookingRepo{}, spaceRepo)

	_, err := svc.Create(context.Background(), 1, 1, 1, tomorrow())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestBookingService_Create_BeginTxError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := &mockBookingRepo{
		beginTxFn: func(_ context.Context) (database.Tx, error) { return nil, dbErr },
	}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(1000), nil
		},
	}
	svc := newBookingSvc(repo, spaceRepo)

	_, err := svc.Create(context.Background(), 1, 1, 1, tomorrow())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBookingService_Create_SlotDoesNotBelong(t *testing.T) {
	repo := &mockBookingRepo{
		slotBelongsToSpaceFn: func(_ context.Context, _, _ int64) (bool, error) { return false, nil },
	}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(1000), nil
		},
	}
	svc := newBookingSvc(repo, spaceRepo)

	_, err := svc.Create(context.Background(), 1, 1, 1, tomorrow())
	if !errors.Is(err, ErrSlotDoesNotBelong) {
		t.Errorf("got %v, want ErrSlotDoesNotBelong", err)
	}
}

func TestBookingService_Create_SlotNotAvailable(t *testing.T) {
	repo := &mockBookingRepo{
		existsActiveBookingFn: func(_ context.Context, _, _ int64, _ time.Time) (bool, error) { return true, nil },
	}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(1000), nil
		},
	}
	svc := newBookingSvc(repo, spaceRepo)

	_, err := svc.Create(context.Background(), 1, 1, 1, tomorrow())
	if !errors.Is(err, ErrSlotNotAvailable) {
		t.Errorf("got %v, want ErrSlotNotAvailable", err)
	}
}

func TestBookingService_Create_HappyPath(t *testing.T) {
	const price = 1500.0
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(price), nil
		},
	}
	repo := &mockBookingRepo{
		createFn: func(_ context.Context, _ database.Tx, _ *domain.Booking) (int64, error) { return 42, nil },
	}
	svc := newBookingSvc(repo, spaceRepo)

	date := tomorrow()
	got, err := svc.Create(context.Background(), 7, 1, 1, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("got ID %d, want 42", got.ID)
	}
	if got.Status != domain.BookingStatusPending {
		t.Errorf("got status %q, want %q", got.Status, domain.BookingStatusPending)
	}
	if got.TotalPrice != price {
		t.Errorf("got price %v, want %v", got.TotalPrice, price)
	}
}

func TestBookingService_Create_DepositAndBalanceAmounts(t *testing.T) {
	const price = 1000.0
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(price), nil
		},
	}
	repo := &mockBookingRepo{
		createFn: func(_ context.Context, _ database.Tx, _ *domain.Booking) (int64, error) { return 1, nil },
	}
	svc := newBookingSvc(repo, spaceRepo)

	before := time.Now().UTC()
	got, err := svc.Create(context.Background(), 7, 1, 1, tomorrow())
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.DepositAmount != 150.0 {
		t.Errorf("got deposit amount %v, want 150.0", got.DepositAmount)
	}
	if got.BalanceAmount != 850.0 {
		t.Errorf("got balance amount %v, want 850.0", got.BalanceAmount)
	}
	if got.DepositStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got deposit status %q, want unpaid", got.DepositStatus)
	}
	if got.BalanceStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got balance status %q, want unpaid", got.BalanceStatus)
	}
	if got.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set, got nil")
	}
	wantMin := before.Add(testHoldTTL)
	wantMax := after.Add(testHoldTTL)
	if got.ExpiresAt.Before(wantMin) || got.ExpiresAt.After(wantMax) {
		t.Errorf("got ExpiresAt %v, want between %v and %v", got.ExpiresAt, wantMin, wantMax)
	}
}

func TestBookingService_Create_ExpiresPendingBookingsBeforeCheckingAvailability(t *testing.T) {
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(1000), nil
		},
	}
	repo := &mockBookingRepo{
		createFn: func(_ context.Context, _ database.Tx, _ *domain.Booking) (int64, error) { return 1, nil },
	}
	svc := newBookingSvc(repo, spaceRepo)

	_, err := svc.Create(context.Background(), 7, 1, 1, tomorrow())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.callLog) < 2 {
		t.Fatalf("expected at least 2 calls, got %v", repo.callLog)
	}
	expireIdx, existsIdx := -1, -1
	for i, call := range repo.callLog {
		if call == "ExpirePendingBookings" && expireIdx == -1 {
			expireIdx = i
		}
		if call == "ExistsActiveBooking" && existsIdx == -1 {
			existsIdx = i
		}
	}
	if expireIdx == -1 || existsIdx == -1 {
		t.Fatalf("expected both calls in log, got %v", repo.callLog)
	}
	if expireIdx > existsIdx {
		t.Errorf("expected ExpirePendingBookings before ExistsActiveBooking, got order %v", repo.callLog)
	}
}

// --- BookingService.CreateManual ---

func TestBookingService_CreateManual_Validations(t *testing.T) {
	svc := newBookingSvc(&mockBookingRepo{}, &mockSpaceRepoForBooking{})

	cases := []struct {
		name          string
		customerName  string
		customerPhone string
		wantErr       error
	}{
		{"empty name", "", "1122334455", ErrNameRequired},
		{"whitespace name", "   ", "1122334455", ErrNameRequired},
		{"empty phone", "Juan", "", ErrPhoneRequired},
		{"whitespace phone", "Juan", "  ", ErrPhoneRequired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateManual(context.Background(), 99, 1, 1, tomorrow(), tc.customerName, tc.customerPhone)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestBookingService_CreateManual_PastDate(t *testing.T) {
	svc := newBookingSvc(&mockBookingRepo{}, &mockSpaceRepoForBooking{})
	_, err := svc.CreateManual(context.Background(), 99, 1, 1, yesterday(), "Juan", "1122334455")
	if !errors.Is(err, ErrInvalidBookingDate) {
		t.Errorf("got %v, want ErrInvalidBookingDate", err)
	}
}

func TestBookingService_CreateManual_SpaceNotFound(t *testing.T) {
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := newBookingSvc(&mockBookingRepo{}, spaceRepo)

	_, err := svc.CreateManual(context.Background(), 99, 1, 1, tomorrow(), "Juan", "1122334455")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestBookingService_CreateManual_SpaceInactive(t *testing.T) {
	inactive := &domain.Space{ID: 1, IsActive: false}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return inactive, nil },
	}
	svc := newBookingSvc(&mockBookingRepo{}, spaceRepo)

	_, err := svc.CreateManual(context.Background(), 99, 1, 1, tomorrow(), "Juan", "1122334455")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestBookingService_CreateManual_SlotDoesNotBelong(t *testing.T) {
	repo := &mockBookingRepo{
		slotBelongsToSpaceFn: func(_ context.Context, _, _ int64) (bool, error) { return false, nil },
	}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(1000), nil
		},
	}
	svc := newBookingSvc(repo, spaceRepo)

	_, err := svc.CreateManual(context.Background(), 99, 1, 1, tomorrow(), "Juan", "1122334455")
	if !errors.Is(err, ErrSlotDoesNotBelong) {
		t.Errorf("got %v, want ErrSlotDoesNotBelong", err)
	}
}

func TestBookingService_CreateManual_SlotNotAvailable(t *testing.T) {
	repo := &mockBookingRepo{
		existsActiveBookingFn: func(_ context.Context, _, _ int64, _ time.Time) (bool, error) { return true, nil },
	}
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(1000), nil
		},
	}
	svc := newBookingSvc(repo, spaceRepo)

	_, err := svc.CreateManual(context.Background(), 99, 1, 1, tomorrow(), "Juan", "1122334455")
	if !errors.Is(err, ErrSlotNotAvailable) {
		t.Errorf("got %v, want ErrSlotNotAvailable", err)
	}
}

func TestBookingService_CreateManual_HappyPath(t *testing.T) {
	const price = 2000.0
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(price), nil
		},
	}
	repo := &mockBookingRepo{
		createFn: func(_ context.Context, _ database.Tx, _ *domain.Booking) (int64, error) { return 99, nil },
	}
	svc := newBookingSvc(repo, spaceRepo)

	got, err := svc.CreateManual(context.Background(), 5, 1, 1, tomorrow(), "Juan", "1122334455")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 99 {
		t.Errorf("got ID %d, want 99", got.ID)
	}
	if got.Status != domain.BookingStatusConfirmed {
		t.Errorf("got status %q, want %q", got.Status, domain.BookingStatusConfirmed)
	}
	if got.TotalPrice != price {
		t.Errorf("got price %v, want %v", got.TotalPrice, price)
	}
}

func TestBookingService_CreateManual_DepositAndBalanceAmountsNoHold(t *testing.T) {
	const price = 1000.0
	spaceRepo := &mockSpaceRepoForBooking{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return activeSpaceWithPrice(price), nil
		},
	}
	repo := &mockBookingRepo{
		createFn: func(_ context.Context, _ database.Tx, _ *domain.Booking) (int64, error) { return 1, nil },
	}
	svc := newBookingSvc(repo, spaceRepo)

	got, err := svc.CreateManual(context.Background(), 5, 1, 1, tomorrow(), "Juan", "1122334455")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.DepositAmount != 150.0 {
		t.Errorf("got deposit amount %v, want 150.0", got.DepositAmount)
	}
	if got.BalanceAmount != 850.0 {
		t.Errorf("got balance amount %v, want 850.0", got.BalanceAmount)
	}
	if got.DepositStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got deposit status %q, want unpaid", got.DepositStatus)
	}
	if got.BalanceStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got balance status %q, want unpaid", got.BalanceStatus)
	}
	if got.ExpiresAt != nil {
		t.Errorf("expected ExpiresAt nil for manual booking, got %v", got.ExpiresAt)
	}
}

// --- BookingService.Cancel ---

func TestBookingService_Cancel_NotFound(t *testing.T) {
	repo := &mockBookingRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return nil, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestBookingService_Cancel_CustomerCannotCancelOtherUsersBooking(t *testing.T) {
	ownerID := int64ptr(42)
	booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusPending}
	repo := &mockBookingRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	// requesterID=99 is not the owner (42)
	err := svc.Cancel(context.Background(), 1, 99, domain.RoleCustomer)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("got %v, want ErrUnauthorized", err)
	}
}

func TestBookingService_Cancel_AlreadyCancelled(t *testing.T) {
	ownerID := int64ptr(10)
	booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusCancelled}
	repo := &mockBookingRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrBookingAlreadyCancelled) {
		t.Errorf("got %v, want ErrBookingAlreadyCancelled", err)
	}
}

func TestBookingService_Cancel_AlreadyCompleted(t *testing.T) {
	ownerID := int64ptr(10)
	booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusCompleted}
	repo := &mockBookingRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrBookingAlreadyCompleted) {
		t.Errorf("got %v, want ErrBookingAlreadyCompleted", err)
	}
}

func TestBookingService_Cancel_OwnerCanCancelOwnBooking(t *testing.T) {
	ownerID := int64ptr(10)
	booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusPending}
	repo := &mockBookingRepo{
		getByIDFn:      func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
		updateStatusFn: func(_ context.Context, _ int64, _ string) (bool, error) { return true, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBookingService_Cancel_StaffCanCancelAnyBooking(t *testing.T) {
	cases := []struct {
		role string
	}{
		{domain.RoleReceptionist},
		{domain.RoleAdmin},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			otherUserID := int64ptr(99)
			booking := &domain.Booking{ID: 1, CustomerUserID: otherUserID, Status: domain.BookingStatusConfirmed}
			repo := &mockBookingRepo{
				getByIDFn:      func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
				updateStatusFn: func(_ context.Context, _ int64, _ string) (bool, error) { return true, nil },
			}
			svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

			// requesterID=1 is NOT the owner (99), but is staff
			err := svc.Cancel(context.Background(), 1, 1, tc.role)
			if err != nil {
				t.Errorf("role %s: unexpected error: %v", tc.role, err)
			}
		})
	}
}

func TestBookingService_Cancel_ManualBookingCanBeCancelledByStaff(t *testing.T) {
	// Manual bookings have no CustomerUserID (anonymous customer)
	booking := &domain.Booking{ID: 1, CustomerUserID: nil, Status: domain.BookingStatusConfirmed}
	repo := &mockBookingRepo{
		getByIDFn:      func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
		updateStatusFn: func(_ context.Context, _ int64, _ string) (bool, error) { return true, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 1, domain.RoleReceptionist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBookingService_Cancel_ManualBookingCannotBeCancelledByCustomer(t *testing.T) {
	// A customer trying to cancel a manual booking they didn't create (CustomerUserID=nil)
	booking := &domain.Booking{ID: 1, CustomerUserID: nil, Status: domain.BookingStatusPending}
	repo := &mockBookingRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("got %v, want ErrUnauthorized", err)
	}
}

func TestBookingService_Cancel_CustomerCannotCancelAfterDepositPaid(t *testing.T) {
	ownerID := int64ptr(10)
	booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusConfirmed, DepositStatus: domain.PaymentStatusPaid}
	repo := &mockBookingRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrCancelRequiresStaffAfterPayment) {
		t.Errorf("got %v, want ErrCancelRequiresStaffAfterPayment", err)
	}
}

func TestBookingService_Cancel_StaffCanCancelAfterDepositPaid(t *testing.T) {
	ownerID := int64ptr(10)
	booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusConfirmed, DepositStatus: domain.PaymentStatusPaid}
	repo := &mockBookingRepo{
		getByIDFn:      func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
		updateStatusFn: func(_ context.Context, _ int64, _ string) (bool, error) { return true, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	err := svc.Cancel(context.Background(), 1, 1, domain.RoleReceptionist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBookingService_Cancel_CustomerCanCancelWhenDepositUnpaidOrPending(t *testing.T) {
	cases := []struct {
		name          string
		depositStatus string
	}{
		{"unpaid", domain.PaymentStatusUnpaid},
		{"pending", domain.PaymentStatusPending},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ownerID := int64ptr(10)
			booking := &domain.Booking{ID: 1, CustomerUserID: ownerID, Status: domain.BookingStatusPending, DepositStatus: tc.depositStatus}
			repo := &mockBookingRepo{
				getByIDFn:      func(_ context.Context, _ int64) (*domain.Booking, error) { return booking, nil },
				updateStatusFn: func(_ context.Context, _ int64, _ string) (bool, error) { return true, nil },
			}
			svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

			err := svc.Cancel(context.Background(), 1, 10, domain.RoleCustomer)
			if err != nil {
				t.Errorf("deposit status %s: unexpected error: %v", tc.depositStatus, err)
			}
		})
	}
}

// --- BookingService.GetByID ---

func TestBookingService_GetByID_NotFound(t *testing.T) {
	repo := &mockBookingRepo{
		getByIDWithDetailsFn: func(_ context.Context, _ int64) (*domain.BookingDetail, error) { return nil, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	_, err := svc.GetByID(context.Background(), 1, 10, domain.RoleCustomer)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestBookingService_GetByID_CustomerCannotSeeOthersBooking(t *testing.T) {
	ownerID := int64ptr(42)
	detail := &domain.BookingDetail{ID: 1, CustomerUserID: ownerID}
	repo := &mockBookingRepo{
		getByIDWithDetailsFn: func(_ context.Context, _ int64) (*domain.BookingDetail, error) { return detail, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	// requesterID=99 is not the owner (42)
	_, err := svc.GetByID(context.Background(), 1, 99, domain.RoleCustomer)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("got %v, want ErrUnauthorized", err)
	}
}

func TestBookingService_GetByID_OwnerCanSeeOwnBooking(t *testing.T) {
	ownerID := int64ptr(10)
	detail := &domain.BookingDetail{ID: 1, CustomerUserID: ownerID}
	repo := &mockBookingRepo{
		getByIDWithDetailsFn: func(_ context.Context, _ int64) (*domain.BookingDetail, error) { return detail, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	got, err := svc.GetByID(context.Background(), 1, 10, domain.RoleCustomer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("got id %d, want 1", got.ID)
	}
}

func TestBookingService_GetByID_StaffCanSeeAnyBooking(t *testing.T) {
	ownerID := int64ptr(99)
	detail := &domain.BookingDetail{ID: 1, CustomerUserID: ownerID}
	repo := &mockBookingRepo{
		getByIDWithDetailsFn: func(_ context.Context, _ int64) (*domain.BookingDetail, error) { return detail, nil },
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	for _, role := range []string{domain.RoleReceptionist, domain.RoleAdmin} {
		_, err := svc.GetByID(context.Background(), 1, 1, role)
		if err != nil {
			t.Errorf("role %s: unexpected error: %v", role, err)
		}
	}
}

// --- BookingService.GetMyBookings ---

func TestBookingService_GetMyBookings_ReturnsOnlyUserBookings(t *testing.T) {
	const userID int64 = 7
	repo := &mockBookingRepo{
		getAllWithDetailsFn: func(_ context.Context, gotUserID *int64, gotDate *time.Time) ([]domain.BookingDetail, error) {
			if gotUserID == nil || *gotUserID != userID {
				t.Errorf("expected userID %d, got %v", userID, gotUserID)
			}
			if gotDate != nil {
				t.Errorf("expected nil date, got %v", gotDate)
			}
			return []domain.BookingDetail{{ID: 1}, {ID: 2}}, nil
		},
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	rows, err := svc.GetMyBookings(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("got %d rows, want 2", len(rows))
	}
}

func TestBookingService_GetMyBookings_RepoError(t *testing.T) {
	dbErr := errors.New("db timeout")
	repo := &mockBookingRepo{
		getAllWithDetailsFn: func(_ context.Context, _ *int64, _ *time.Time) ([]domain.BookingDetail, error) {
			return nil, dbErr
		},
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	_, err := svc.GetMyBookings(context.Background(), 1)
	if !errors.Is(err, dbErr) {
		t.Errorf("got %v, want wrapped dbErr", err)
	}
}

// --- BookingService.GetAll ---

func TestBookingService_GetAll_NoFilter(t *testing.T) {
	repo := &mockBookingRepo{
		getAllWithDetailsFn: func(_ context.Context, gotUserID *int64, gotDate *time.Time) ([]domain.BookingDetail, error) {
			if gotUserID != nil {
				t.Errorf("expected nil userID, got %v", gotUserID)
			}
			if gotDate != nil {
				t.Errorf("expected nil date, got %v", gotDate)
			}
			return []domain.BookingDetail{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	rows, err := svc.GetAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("got %d rows, want 3", len(rows))
	}
}

func TestBookingService_GetAll_WithDateFilter(t *testing.T) {
	date := tomorrow()
	repo := &mockBookingRepo{
		getAllWithDetailsFn: func(_ context.Context, gotUserID *int64, gotDate *time.Time) ([]domain.BookingDetail, error) {
			if gotUserID != nil {
				t.Errorf("expected nil userID, got %v", gotUserID)
			}
			if gotDate == nil || !gotDate.Equal(date) {
				t.Errorf("expected date %v, got %v", date, gotDate)
			}
			return []domain.BookingDetail{{ID: 5}}, nil
		},
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	rows, err := svc.GetAll(context.Background(), &date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("got %d rows, want 1", len(rows))
	}
}

func TestBookingService_GetAll_RepoError(t *testing.T) {
	dbErr := errors.New("connection lost")
	repo := &mockBookingRepo{
		getAllWithDetailsFn: func(_ context.Context, _ *int64, _ *time.Time) ([]domain.BookingDetail, error) {
			return nil, dbErr
		},
	}
	svc := newBookingSvc(repo, &mockSpaceRepoForBooking{})

	_, err := svc.GetAll(context.Background(), nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("got %v, want wrapped dbErr", err)
	}
}
