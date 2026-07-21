package repository_test

import (
	"context"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/repository"
	"H1-canchas/pkg/database"
)

// paymentTestRepo is a local interface that captures the methods under test.
// *paymentRepo (unexported) satisfies it automatically via Go structural typing.
type paymentTestRepo interface {
	Create(ctx context.Context, tx database.Tx, p *domain.Payment) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Payment, error)
	GetLatestByBookingID(ctx context.Context, bookingID int64, kind string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, tx database.Tx, id int64, status string, rawDetail *string, externalPaymentID *string) (bool, error)
}

// beginTx opens a real transaction against testDB.
// database.Tx is satisfied structurally by *sqlx.Tx, no wrapping needed.
func beginTx(t *testing.T) database.Tx {
	t.Helper()
	tx, err := testDB.BeginTxx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTxx: %v", err)
	}
	return tx
}

// createPayment opens a tx, inserts, and commits. Returns the new payment ID.
func createPayment(t *testing.T, repo paymentTestRepo, p *domain.Payment) int64 {
	t.Helper()
	tx := beginTx(t)
	id, err := repo.Create(context.Background(), tx, p)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("Create: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return id
}

func newPendingPayment(bookingID int64) *domain.Payment {
	return &domain.Payment{
		BookingID: bookingID,
		Kind:      domain.PaymentKindDeposit,
		Provider:  domain.PaymentProviderMercadoPago,
		Status:    domain.PaymentTxStatusPending,
		Amount:    150.0,
	}
}

// seedBookingForPayment creates a user, space, slot, and pending booking,
// returning the booking ID so payment tests have a valid FK target.
func seedBookingForPayment(t *testing.T) int64 {
	t.Helper()
	bookingRepo := repository.NewBookingRepository(testDB)
	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	return createBooking(t, bookingRepo, newPendingBooking(spaceID, slotID, userID))
}

// --- Create ---

func TestPaymentRepo_Create_ReturnsPositiveID(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	bookingID := seedBookingForPayment(t)

	id := createPayment(t, repo, newPendingPayment(bookingID))

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

// --- GetByID ---

func TestPaymentRepo_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	got, err := repo.GetByID(context.Background(), 99999)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestPaymentRepo_GetByID_Found(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	bookingID := seedBookingForPayment(t)
	id := createPayment(t, repo, newPendingPayment(bookingID))

	got, err := repo.GetByID(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected payment, got nil")
	}
	if got.BookingID != bookingID {
		t.Errorf("got booking_id %d, want %d", got.BookingID, bookingID)
	}
	if got.Kind != domain.PaymentKindDeposit {
		t.Errorf("got kind %q, want %q", got.Kind, domain.PaymentKindDeposit)
	}
	if got.Status != domain.PaymentTxStatusPending {
		t.Errorf("got status %q, want %q", got.Status, domain.PaymentTxStatusPending)
	}
	if got.Provider != domain.PaymentProviderMercadoPago {
		t.Errorf("got provider %q, want %q", got.Provider, domain.PaymentProviderMercadoPago)
	}
}

// --- GetLatestByBookingID ---

func TestPaymentRepo_GetLatestByBookingID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	bookingID := seedBookingForPayment(t)

	got, err := repo.GetLatestByBookingID(context.Background(), bookingID, domain.PaymentKindDeposit)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestPaymentRepo_GetLatestByBookingID_ReturnsMostRecent(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	bookingID := seedBookingForPayment(t)

	firstID := createPayment(t, repo, newPendingPayment(bookingID))
	// Ensure a distinct created_at for the second attempt (retry after rejection).
	time.Sleep(10 * time.Millisecond)
	secondPayment := newPendingPayment(bookingID)
	secondPayment.Amount = 200.0
	secondID := createPayment(t, repo, secondPayment)

	got, err := repo.GetLatestByBookingID(context.Background(), bookingID, domain.PaymentKindDeposit)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected payment, got nil")
	}
	if got.ID != secondID {
		t.Errorf("got ID %d, want most recent ID %d (first was %d)", got.ID, secondID, firstID)
	}
	if got.Amount != 200.0 {
		t.Errorf("got amount %v, want 200.0", got.Amount)
	}
}

func TestPaymentRepo_GetLatestByBookingID_ScopedToKind(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	bookingID := seedBookingForPayment(t)

	depositPayment := newPendingPayment(bookingID)
	depositID := createPayment(t, repo, depositPayment)

	time.Sleep(10 * time.Millisecond)
	balancePayment := newPendingPayment(bookingID)
	balancePayment.Kind = domain.PaymentKindBalance
	balancePayment.Amount = 850.0
	balanceID := createPayment(t, repo, balancePayment)

	// Even though the balance attempt is more recent overall, asking for
	// "deposit" must return the deposit attempt, not the balance one.
	gotDeposit, err := repo.GetLatestByBookingID(context.Background(), bookingID, domain.PaymentKindDeposit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDeposit == nil || gotDeposit.ID != depositID {
		t.Errorf("got %+v, want deposit payment ID %d", gotDeposit, depositID)
	}

	gotBalance, err := repo.GetLatestByBookingID(context.Background(), bookingID, domain.PaymentKindBalance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBalance == nil || gotBalance.ID != balanceID {
		t.Errorf("got %+v, want balance payment ID %d", gotBalance, balanceID)
	}
}

// --- UpdateStatus ---

func TestPaymentRepo_UpdateStatus_ReturnsTrueAndUpdates(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	bookingID := seedBookingForPayment(t)
	id := createPayment(t, repo, newPendingPayment(bookingID))

	detail := "accredited"
	externalID := "mp-123456"

	tx := beginTx(t)
	updated, err := repo.UpdateStatus(context.Background(), tx, id, domain.PaymentTxStatusApproved, &detail, &externalID)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if !updated {
		t.Error("expected true, got false")
	}

	got, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != domain.PaymentTxStatusApproved {
		t.Errorf("got status %q, want %q", got.Status, domain.PaymentTxStatusApproved)
	}
	if got.RawStatusDetail == nil || *got.RawStatusDetail != detail {
		t.Errorf("got raw_status_detail %v, want %q", got.RawStatusDetail, detail)
	}
	if got.ExternalPaymentID == nil || *got.ExternalPaymentID != externalID {
		t.Errorf("got external_payment_id %v, want %q", got.ExternalPaymentID, externalID)
	}
}

func TestPaymentRepo_UpdateStatus_ReturnsFalseOnMiss(t *testing.T) {
	truncateAll(t)
	repo := repository.NewPaymentRepository(testDB)

	tx := beginTx(t)
	updated, err := repo.UpdateStatus(context.Background(), tx, 99999, domain.PaymentTxStatusApproved, nil, nil)
	_ = tx.Rollback()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Error("expected false, got true")
	}
}
