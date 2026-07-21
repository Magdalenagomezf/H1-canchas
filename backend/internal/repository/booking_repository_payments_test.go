package repository_test

import (
	"context"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/repository"
	"H1-canchas/pkg/database"
)

// bookingPaymentTestRepo is a local interface that captures the payment-related
// methods under test. *bookingRepo (unexported) satisfies it structurally.
type bookingPaymentTestRepo interface {
	BeginTx(ctx context.Context) (database.Tx, error)
	Create(ctx context.Context, tx database.Tx, b *domain.Booking) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id int64, status string) (bool, error)
	ConfirmAfterPayment(ctx context.Context, tx database.Tx, bookingID int64, kind string) (bool, error)
	MarkPaidManually(ctx context.Context, id int64, kind, method string, recordedBy int64) (bool, error)
	MarkUnpaidManually(ctx context.Context, id int64, kind string) (bool, error)
	ExpirePendingBookings(ctx context.Context, spaceID, slotID int64, date time.Time) (int64, error)
	ExpireAllStalePending(ctx context.Context) (int64, error)
}

// setExpiresAt writes expires_at directly since Create does not accept it
// (it relies on the DB default of NULL — the hold TTL is set by the
// service layer in a later phase). Tests set it directly to simulate a hold.
func setExpiresAt(t *testing.T, bookingID int64, expiresAt *time.Time) {
	t.Helper()
	_, err := testDB.Exec(`UPDATE bookings SET expires_at = $1 WHERE id = $2`, expiresAt, bookingID)
	if err != nil {
		t.Fatalf("setExpiresAt: %v", err)
	}
}

// --- ConfirmAfterPayment ---

func TestBookingRepo_ConfirmAfterPayment_Deposit_HappyPath(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	future := time.Now().UTC().Add(20 * time.Minute)
	setExpiresAt(t, id, &future)

	tx, err := repo.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	updated, err := repo.ConfirmAfterPayment(context.Background(), tx, id, domain.PaymentKindDeposit)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("ConfirmAfterPayment: %v", err)
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
	if got.Status != domain.BookingStatusConfirmed {
		t.Errorf("got status %q, want %q", got.Status, domain.BookingStatusConfirmed)
	}
	if got.DepositStatus != domain.PaymentStatusPaid {
		t.Errorf("got deposit_status %q, want %q", got.DepositStatus, domain.PaymentStatusPaid)
	}
	if got.ExpiresAt != nil {
		t.Errorf("expected expires_at to be cleared, got %v", got.ExpiresAt)
	}
	if got.BalanceStatus == domain.PaymentStatusPaid {
		t.Error("expected balance_status to remain unaffected by deposit confirmation")
	}
}

func TestBookingRepo_ConfirmAfterPayment_Balance_HappyPath(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	// Confirm the deposit first so the booking is confirmed, then pay the balance.
	if _, err := repo.UpdateStatus(context.Background(), id, domain.BookingStatusConfirmed); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	tx, err := repo.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	updated, err := repo.ConfirmAfterPayment(context.Background(), tx, id, domain.PaymentKindBalance)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("ConfirmAfterPayment: %v", err)
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
	if got.BalanceStatus != domain.PaymentStatusPaid {
		t.Errorf("got balance_status %q, want %q", got.BalanceStatus, domain.PaymentStatusPaid)
	}
	// Balance has no hold/expiry and must never touch status.
	if got.Status != domain.BookingStatusConfirmed {
		t.Errorf("got status %q, want unchanged %q", got.Status, domain.BookingStatusConfirmed)
	}
	if got.DepositStatus == domain.PaymentStatusPaid {
		t.Error("expected deposit_status to remain unaffected by balance confirmation")
	}
}

func TestBookingRepo_ConfirmAfterPayment_NoOpOnTerminalStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		kind   string
	}{
		{"already cancelled, deposit", domain.BookingStatusCancelled, domain.PaymentKindDeposit},
		{"already expired, deposit", domain.BookingStatusExpired, domain.PaymentKindDeposit},
		{"already cancelled, balance", domain.BookingStatusCancelled, domain.PaymentKindBalance},
		{"already expired, balance", domain.BookingStatusExpired, domain.PaymentKindBalance},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateAll(t)
			repo := repository.NewBookingRepository(testDB)

			userID := seedUser(t)
			spaceID := seedSpace(t)
			slotID := seedSlot(t, spaceID)
			id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

			if _, err := repo.UpdateStatus(context.Background(), id, tt.status); err != nil {
				t.Fatalf("UpdateStatus: %v", err)
			}

			tx, err := repo.BeginTx(context.Background())
			if err != nil {
				t.Fatalf("BeginTx: %v", err)
			}
			updated, err := repo.ConfirmAfterPayment(context.Background(), tx, id, tt.kind)
			_ = tx.Rollback()
			if err != nil {
				t.Fatalf("ConfirmAfterPayment: %v", err)
			}

			if updated {
				t.Error("expected false (no-op), got true")
			}

			got, err := repo.GetByID(context.Background(), id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Status != tt.status {
				t.Errorf("got status %q, want unchanged %q", got.Status, tt.status)
			}
			if got.DepositStatus == domain.PaymentStatusPaid {
				t.Error("expected deposit_status to remain unpaid, got paid")
			}
			if got.BalanceStatus == domain.PaymentStatusPaid {
				t.Error("expected balance_status to remain unpaid, got paid")
			}
		})
	}
}

// --- MarkPaidManually / MarkUnpaidManually ---

func TestBookingRepo_MarkPaidManually_Deposit_SetsFields(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	staffID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	updated, err := repo.MarkPaidManually(context.Background(), id, domain.PaymentKindDeposit, domain.PaymentMethodCash, staffID)
	if err != nil {
		t.Fatalf("MarkPaidManually: %v", err)
	}
	if !updated {
		t.Error("expected true, got false")
	}

	got, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DepositStatus != domain.PaymentStatusPaid {
		t.Errorf("got deposit_status %q, want %q", got.DepositStatus, domain.PaymentStatusPaid)
	}
	if got.DepositMethod == nil || *got.DepositMethod != domain.PaymentMethodCash {
		t.Errorf("got deposit_method %v, want %q", got.DepositMethod, domain.PaymentMethodCash)
	}
	if got.DepositRecordedBy == nil || *got.DepositRecordedBy != staffID {
		t.Errorf("got deposit_recorded_by %v, want %d", got.DepositRecordedBy, staffID)
	}
	// Marking the deposit must never touch the balance columns.
	if got.BalanceStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got balance_status %q, want unaffected %q", got.BalanceStatus, domain.PaymentStatusUnpaid)
	}
	if got.BalanceMethod != nil {
		t.Errorf("expected balance_method nil, got %v", *got.BalanceMethod)
	}
	if got.BalanceRecordedBy != nil {
		t.Errorf("expected balance_recorded_by nil, got %v", *got.BalanceRecordedBy)
	}
}

func TestBookingRepo_MarkPaidManually_Balance_SetsFields(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	staffID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	updated, err := repo.MarkPaidManually(context.Background(), id, domain.PaymentKindBalance, domain.PaymentMethodPosnet, staffID)
	if err != nil {
		t.Fatalf("MarkPaidManually: %v", err)
	}
	if !updated {
		t.Error("expected true, got false")
	}

	got, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.BalanceStatus != domain.PaymentStatusPaid {
		t.Errorf("got balance_status %q, want %q", got.BalanceStatus, domain.PaymentStatusPaid)
	}
	if got.BalanceMethod == nil || *got.BalanceMethod != domain.PaymentMethodPosnet {
		t.Errorf("got balance_method %v, want %q", got.BalanceMethod, domain.PaymentMethodPosnet)
	}
	if got.BalanceRecordedBy == nil || *got.BalanceRecordedBy != staffID {
		t.Errorf("got balance_recorded_by %v, want %d", got.BalanceRecordedBy, staffID)
	}
	// Marking the balance must never touch the deposit columns or status.
	if got.DepositStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got deposit_status %q, want unaffected %q", got.DepositStatus, domain.PaymentStatusUnpaid)
	}
	if got.Status != domain.BookingStatusPending {
		t.Errorf("got status %q, want unaffected %q", got.Status, domain.BookingStatusPending)
	}
}

func TestBookingRepo_MarkPaidManually_ReturnsFalseOnMiss(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	updated, err := repo.MarkPaidManually(context.Background(), 99999, domain.PaymentKindDeposit, domain.PaymentMethodCash, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Error("expected false, got true")
	}
}

func TestBookingRepo_MarkUnpaidManually_Deposit_ClearsFields(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	staffID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	if _, err := repo.MarkPaidManually(context.Background(), id, domain.PaymentKindDeposit, domain.PaymentMethodTransfer, staffID); err != nil {
		t.Fatalf("MarkPaidManually: %v", err)
	}

	updated, err := repo.MarkUnpaidManually(context.Background(), id, domain.PaymentKindDeposit)
	if err != nil {
		t.Fatalf("MarkUnpaidManually: %v", err)
	}
	if !updated {
		t.Error("expected true, got false")
	}

	got, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DepositStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got deposit_status %q, want %q", got.DepositStatus, domain.PaymentStatusUnpaid)
	}
	if got.DepositMethod != nil {
		t.Errorf("expected deposit_method nil, got %v", *got.DepositMethod)
	}
	if got.DepositRecordedBy != nil {
		t.Errorf("expected deposit_recorded_by nil, got %v", *got.DepositRecordedBy)
	}
}

func TestBookingRepo_MarkUnpaidManually_Balance_ClearsFields(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	staffID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	if _, err := repo.MarkPaidManually(context.Background(), id, domain.PaymentKindBalance, domain.PaymentMethodTransfer, staffID); err != nil {
		t.Fatalf("MarkPaidManually: %v", err)
	}

	updated, err := repo.MarkUnpaidManually(context.Background(), id, domain.PaymentKindBalance)
	if err != nil {
		t.Fatalf("MarkUnpaidManually: %v", err)
	}
	if !updated {
		t.Error("expected true, got false")
	}

	got, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.BalanceStatus != domain.PaymentStatusUnpaid {
		t.Errorf("got balance_status %q, want %q", got.BalanceStatus, domain.PaymentStatusUnpaid)
	}
	if got.BalanceMethod != nil {
		t.Errorf("expected balance_method nil, got %v", *got.BalanceMethod)
	}
	if got.BalanceRecordedBy != nil {
		t.Errorf("expected balance_recorded_by nil, got %v", *got.BalanceRecordedBy)
	}
}

func TestBookingRepo_MarkUnpaidManually_ReturnsFalseOnMiss(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	updated, err := repo.MarkUnpaidManually(context.Background(), 99999, domain.PaymentKindDeposit)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated {
		t.Error("expected false, got true")
	}
}

// --- ExpirePendingBookings ---

func TestBookingRepo_ExpirePendingBookings_ExpiresOnlyPastTTL(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	date := tomorrow()

	// Booking A: expires_at in the past -> should expire.
	slotA := seedSlot(t, spaceID)
	idA := createBooking(t, repo, newPendingBooking(spaceID, slotA, userID))
	past := time.Now().UTC().Add(-1 * time.Hour)
	setExpiresAt(t, idA, &past)

	// Booking B: expires_at in the future -> should stay pending.
	slotB := seedSlot(t, spaceID)
	idB := createBooking(t, repo, newPendingBooking(spaceID, slotB, userID))
	future := time.Now().UTC().Add(1 * time.Hour)
	setExpiresAt(t, idB, &future)

	// Booking C: no expires_at (manual booking) -> should stay pending.
	slotC := seedSlot(t, spaceID)
	idC := createBooking(t, repo, newPendingBooking(spaceID, slotC, userID))

	countA, err := repo.ExpirePendingBookings(context.Background(), spaceID, slotA, date)
	if err != nil {
		t.Fatalf("ExpirePendingBookings (A): %v", err)
	}
	if countA != 1 {
		t.Errorf("expected 1 row expired for slot A, got %d", countA)
	}

	countB, err := repo.ExpirePendingBookings(context.Background(), spaceID, slotB, date)
	if err != nil {
		t.Fatalf("ExpirePendingBookings (B): %v", err)
	}
	if countB != 0 {
		t.Errorf("expected 0 rows expired for slot B (future TTL), got %d", countB)
	}

	countC, err := repo.ExpirePendingBookings(context.Background(), spaceID, slotC, date)
	if err != nil {
		t.Fatalf("ExpirePendingBookings (C): %v", err)
	}
	if countC != 0 {
		t.Errorf("expected 0 rows expired for slot C (no expires_at), got %d", countC)
	}

	gotA, _ := repo.GetByID(context.Background(), idA)
	if gotA.Status != domain.BookingStatusExpired {
		t.Errorf("booking A: got status %q, want %q", gotA.Status, domain.BookingStatusExpired)
	}
	gotB, _ := repo.GetByID(context.Background(), idB)
	if gotB.Status != domain.BookingStatusPending {
		t.Errorf("booking B: got status %q, want %q", gotB.Status, domain.BookingStatusPending)
	}
	gotC, _ := repo.GetByID(context.Background(), idC)
	if gotC.Status != domain.BookingStatusPending {
		t.Errorf("booking C: got status %q, want %q", gotC.Status, domain.BookingStatusPending)
	}
}

// --- ExpireAllStalePending ---

func TestBookingRepo_ExpireAllStalePending_SweepsGlobally(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)

	slotStale := seedSlot(t, spaceID)
	idStale := createBooking(t, repo, newPendingBooking(spaceID, slotStale, userID))
	past := time.Now().UTC().Add(-1 * time.Hour)
	setExpiresAt(t, idStale, &past)

	slotFresh := seedSlot(t, spaceID)
	idFresh := createBooking(t, repo, newPendingBooking(spaceID, slotFresh, userID))
	future := time.Now().UTC().Add(1 * time.Hour)
	setExpiresAt(t, idFresh, &future)

	count, err := repo.ExpireAllStalePending(context.Background())
	if err != nil {
		t.Fatalf("ExpireAllStalePending: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row expired globally, got %d", count)
	}

	gotStale, _ := repo.GetByID(context.Background(), idStale)
	if gotStale.Status != domain.BookingStatusExpired {
		t.Errorf("stale booking: got status %q, want %q", gotStale.Status, domain.BookingStatusExpired)
	}
	gotFresh, _ := repo.GetByID(context.Background(), idFresh)
	if gotFresh.Status != domain.BookingStatusPending {
		t.Errorf("fresh booking: got status %q, want %q", gotFresh.Status, domain.BookingStatusPending)
	}
}
