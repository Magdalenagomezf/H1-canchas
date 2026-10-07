package repository_test

import (
	"context"
	"testing"

	"H1-canchas/internal/repository"
)

func TestBookingRepo_ClaimConfirmationEmail_OnlyFirstClaimWins(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	first, err := repo.ClaimConfirmationEmail(context.Background(), id)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if !first {
		t.Error("first claim: got false, want true")
	}

	second, err := repo.ClaimConfirmationEmail(context.Background(), id)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if second {
		t.Error("second claim: got true, want false")
	}
}

func TestBookingRepo_ReleaseConfirmationEmail_AllowsClaimAgain(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	id := createBooking(t, repo, newPendingBooking(spaceID, slotID, userID))

	if ok, err := repo.ClaimConfirmationEmail(context.Background(), id); err != nil || !ok {
		t.Fatalf("claim: ok=%v err=%v", ok, err)
	}
	if err := repo.ReleaseConfirmationEmail(context.Background(), id); err != nil {
		t.Fatalf("release: %v", err)
	}

	ok, err := repo.ClaimConfirmationEmail(context.Background(), id)
	if err != nil {
		t.Fatalf("claim after release: %v", err)
	}
	if !ok {
		t.Error("claim after release: got false, want true")
	}
}

func TestBookingRepo_ClaimConfirmationEmail_UnknownBooking(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	ok, err := repo.ClaimConfirmationEmail(context.Background(), 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("got true for a non-existent booking, want false")
	}
}

func TestBookingRepo_GetByIDWithDetails_IncludesUserEmail(t *testing.T) {
	truncateAll(t)
	repo := repository.NewBookingRepository(testDB)

	userID := seedUser(t)
	if _, err := testDB.Exec(`UPDATE users SET email = 'ana@example.com' WHERE id = $1`, userID); err != nil {
		t.Fatalf("set email: %v", err)
	}
	spaceID := seedSpace(t)
	slotID := seedSlot(t, spaceID)
	b := newPendingBooking(spaceID, slotID, userID)
	b.CustomerUserID = &userID
	id := createBooking(t, repo, b)

	got, err := repo.GetByIDWithDetails(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.CustomerUserEmail == nil || *got.CustomerUserEmail != "ana@example.com" {
		t.Errorf("got CustomerUserEmail %v, want ana@example.com", got)
	}
}
