package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/repository"
)

func TestSpaceRepo_Create_ReturnsID(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	space := &domain.Space{Name: "Padel Norte", Type: domain.SpaceTypePadel, PricePerSlot: 1500.0}
	id, err := repo.Create(context.Background(), space)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestSpaceRepo_GetAll_Empty(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	spaces, err := repo.GetAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spaces) != 0 {
		t.Errorf("expected empty slice, got %d items", len(spaces))
	}
}

func TestSpaceRepo_GetAll_OnlyActive(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	// Insert two active spaces then deactivate one.
	id1, _ := repo.Create(context.Background(), &domain.Space{Name: "A", Type: domain.SpaceTypePadel, PricePerSlot: 1000})
	_, _ = repo.Create(context.Background(), &domain.Space{Name: "B", Type: domain.SpaceTypePadel, PricePerSlot: 1000})
	_ = repo.Deactivate(context.Background(), id1)

	spaces, err := repo.GetAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spaces) != 1 {
		t.Errorf("expected 1 active space, got %d", len(spaces))
	}
	if spaces[0].Name != "B" {
		t.Errorf("expected space B, got %q", spaces[0].Name)
	}
}

func TestSpaceRepo_GetByID_Found(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	id, _ := repo.Create(context.Background(), &domain.Space{Name: "Futbol Sur", Type: domain.SpaceTypeFutbol, PricePerSlot: 2000})

	got, err := repo.GetByID(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected space, got nil")
	}
	if got.Name != "Futbol Sur" {
		t.Errorf("got name %q, want %q", got.Name, "Futbol Sur")
	}
}

func TestSpaceRepo_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	got, err := repo.GetByID(context.Background(), 99999)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestSpaceRepo_CreateSlot_ReturnsID(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	spaceID := seedSpace(t)
	slot := &domain.SpaceSlot{SpaceID: spaceID, Label: "09:00 - 10:00"}
	id, err := repo.CreateSlot(context.Background(), slot)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestSpaceRepo_GetSlotsBySpaceID_ReturnsActiveSlots(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	spaceID := seedSpace(t)
	_, _ = repo.CreateSlot(context.Background(), &domain.SpaceSlot{SpaceID: spaceID, Label: "08:00 - 09:00"})
	_, _ = repo.CreateSlot(context.Background(), &domain.SpaceSlot{SpaceID: spaceID, Label: "09:00 - 10:00"})

	slots, err := repo.GetSlotsBySpaceID(context.Background(), spaceID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(slots))
	}
}

func TestSpaceRepo_GetSlotsBySpaceID_EmptyForUnknownSpace(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	slots, err := repo.GetSlotsBySpaceID(context.Background(), 99999)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("expected empty slice, got %d items", len(slots))
	}
}

func TestSpaceRepo_Deactivate_Success(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	id, _ := repo.Create(context.Background(), &domain.Space{Name: "Quincho Eventos", Type: domain.SpaceTypeQuincho, PricePerSlot: 500})

	if err := repo.Deactivate(context.Background(), id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), id)
	if got == nil || got.IsActive {
		t.Errorf("expected space to be inactive")
	}
}

func TestSpaceRepo_Deactivate_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewSpaceRepository(testDB)

	err := repo.Deactivate(context.Background(), 99999)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}
