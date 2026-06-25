package repository_test

import (
	"context"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/repository"
	"H1-canchas/pkg/database"
)

// bookingTestRepo is a local interface that captures the methods under test.
// *bookingRepo (unexported) satisfies it automatically via Go structural typing.
type bookingTestRepo interface {
	BeginTx(ctx context.Context) (database.Tx, error)
	Create(ctx context.Context, tx database.Tx, b *domain.Booking) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id int64, status string) (bool, error)
	ExistsActiveBooking(ctx context.Context, spaceID, slotID int64, date time.Time) (bool, error)
	SlotBelongsToSpace(ctx context.Context, spaceID, slotID int64) (bool, error)
	GetAllWithDetails(ctx context.Context, userID *int64, date *time.Time) ([]domain.BookingDetail, error)
	GetByIDWithDetails(ctx context.Context, id int64) (*domain.BookingDetail, error)
}

// createBooking opens a tx, inserts, and commits. Returns the new booking ID.
func createBooking(t *testing.T, repo bookingTestRepo, b *domain.Booking) int64 {
	t.Helper()
	tx, err := repo.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	id, err := repo.Create(context.Background(), tx, b)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("Create: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return id
}

func newPendingBooking(spaceID, slotID, createdBy int64) *domain.Booking {
	name := "Walk-in"
	phone := "0000000000"
	return &domain.Booking{
		SpaceID:       spaceID,
		SlotID:        slotID,
		CreatedBy:     createdBy,
		BookingDate:   tomorrow(),
		Status:        domain.BookingStatusPending,
		TotalPrice:    1000.0,
		CustomerName:  &name,
		CustomerPhone: &phone,
	}
}

// --- ExistsActiveBooking ---

func TestBookingRepo_ExistsActiveBooking_FalseWhenEmpty(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	exists, err := repo.ExistsActiveBooking(context.Background(), spaceID, slotID, tomorrow())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false, got true")
	}
}

func TestBookingRepo_ExistsActiveBooking_TrueAfterInsert(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	exists, err := repo.ExistsActiveBooking(context.Background(), spaceID, slotID, tomorrow())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true, got false")
	}
}

// --- SlotBelongsToSpace ---

func TestBookingRepo_SlotBelongsToSpace_True(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	ok, err := repo.SlotBelongsToSpace(context.Background(), spaceID, slotID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true, got false")
	}
}

func TestBookingRepo_SlotBelongsToSpace_FalseForWrongSpace(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	spaceA := seedSpace(t)
	spaceB := seedSpace(t)
	slotOfA := seedSlot(t, spaceA)

	// slotOfA belongs to spaceA, not spaceB.
	ok, err := repo.SlotBelongsToSpace(context.Background(), spaceB, slotOfA)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false, got true")
	}
}

// --- Create ---

func TestBookingRepo_Create_ReturnsPositiveID(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

// TestBookingRepo_Create_UniqueIndexRejectsDouble verifies that the partial unique index
// unique_active_booking (space_id, slot_id, booking_date) WHERE status IN ('pending','confirmed')
// prevents two active bookings for the same slot on the same date.
func TestBookingRepo_Create_UniqueIndexRejectsDouble(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	// Second booking for the same space+slot+date must fail at DB level.
	tx, err := repo.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	_, err = repo.Create(context.Background(), tx, newPendingBooking(spaceID, slotID, userID))
	_ = tx.Rollback()

	if err == nil {
		t.Error("expected unique constraint error, got nil")
	}
}

// TestBookingRepo_Create_CancelledDoesNotBlock verifies the partial index covers only
// pending/confirmed — a cancelled booking must NOT block a new one for the same slot.
func TestBookingRepo_Create_CancelledDoesNotBlock(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	firstID := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	// Cancel the first booking.
	_, err := repo.UpdateStatus(context.Background(), firstID, domain.BookingStatusCancelled)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	// A new booking for the same slot on the same date must succeed.
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))
	if id <= 0 {
		t.Errorf("expected second booking to succeed, got ID %d", id)
	}
}

// --- GetByID ---

func TestBookingRepo_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	got, err := repo.GetByID(context.Background(), 99999)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestBookingRepo_GetByID_Found(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	got, err := repo.GetByID(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected booking, got nil")
	}
	if got.Status != domain.BookingStatusPending {
		t.Errorf("got status %q, want %q", got.Status, domain.BookingStatusPending)
	}
}

// --- UpdateStatus ---

func TestBookingRepo_UpdateStatus_ReturnsTrueOnMatch(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	updated, err := repo.UpdateStatus(context.Background(), id, domain.BookingStatusCancelled)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Error("expected true, got false")
	}

	got, _ := repo.GetByID(context.Background(), id)
	if got.Status != domain.BookingStatusCancelled {
		t.Errorf("got status %q, want %q", got.Status, domain.BookingStatusCancelled)
	}
}

func TestBookingRepo_UpdateStatus_ReturnsFalseOnMiss(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	updated, err := repo.UpdateStatus(context.Background(), 99999, domain.BookingStatusCancelled)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Error("expected false, got true")
	}
}

// --- GetAllWithDetails ---

func TestBookingRepo_GetAllWithDetails_FilterByUser(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slot1 := seedSlot(t, spaceID)
	slot2 := seedSlot(t, spaceID)

	// Booking linked to userID as customer (slot 1).
	b1 := newPendingBooking(spaceID, slot1, userID)
	b1.CustomerUserID = &userID
	createBooking(t, repo, b1)

	// Walk-in booking with no customer_user_id — staff manual (slot 2, different slot).
	createBooking(t, repo, newPendingBooking(spaceID, slot2, userID))

	// Filter by userID — only the one with customer_user_id = userID should appear.
	rows, err := repo.GetAllWithDetails(context.Background(), &userID, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 booking for user, got %d", len(rows))
	}
}

// --- GetByIDWithDetails ---

func TestBookingRepo_GetByIDWithDetails_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	got, err := repo.GetByIDWithDetails(context.Background(), 99999)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestBookingRepo_GetByIDWithDetails_JoinsWork(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	got, err := repo.GetByIDWithDetails(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected detail, got nil")
	}
	if got.SpaceID != spaceID {
		t.Errorf("got space_id %d, want %d", got.SpaceID, spaceID)
	}
	if got.SlotID != slotID {
		t.Errorf("got slot_id %d, want %d", got.SlotID, slotID)
	}
}

// TestBookingRepo_GetAllWithDetails_FilterByDate verifies date filtering.
func TestBookingRepo_GetAllWithDetails_FilterByDate(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)

	createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	// Filter by a different date — should return 0 results.
	otherDate := tomorrow().AddDate(0, 0, 7)
	rows, err := repo.GetAllWithDetails(context.Background(), nil, &otherDate)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 results for different date, got %d", len(rows))
	}
}
