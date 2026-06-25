package service

import (
	"context"
	"errors"
	"testing"

	"H1-canchas/internal/domain"
)

// mockSpaceRepo is a test double for SpaceRepo.
// Each field is a function so each test can inject the behavior it needs.
type mockSpaceRepo struct {
	createFn            func(ctx context.Context, space *domain.Space) (int64, error)
	getAllFn             func(ctx context.Context) ([]domain.Space, error)
	getByIDFn           func(ctx context.Context, id int64) (*domain.Space, error)
	getSlotsBySpaceIDFn func(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error)
	updateFn            func(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error
	createSlotFn        func(ctx context.Context, slot *domain.SpaceSlot) (int64, error)
	deactivateFn        func(ctx context.Context, id int64) error
}

func (m *mockSpaceRepo) Create(ctx context.Context, space *domain.Space) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, space)
	}
	return 1, nil
}

func (m *mockSpaceRepo) GetAll(ctx context.Context) ([]domain.Space, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}

func (m *mockSpaceRepo) GetByID(ctx context.Context, id int64) (*domain.Space, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSpaceRepo) GetSlotsBySpaceID(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error) {
	if m.getSlotsBySpaceIDFn != nil {
		return m.getSlotsBySpaceIDFn(ctx, spaceID)
	}
	return nil, nil
}

func (m *mockSpaceRepo) Update(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, name, description, pricePerSlot)
	}
	return nil
}

func (m *mockSpaceRepo) CreateSlot(ctx context.Context, slot *domain.SpaceSlot) (int64, error) {
	if m.createSlotFn != nil {
		return m.createSlotFn(ctx, slot)
	}
	return 1, nil
}

func (m *mockSpaceRepo) Deactivate(ctx context.Context, id int64) error {
	if m.deactivateFn != nil {
		return m.deactivateFn(ctx, id)
	}
	return nil
}

// ptr is a helper to get a pointer to a string literal.
func ptr(s string) *string { return &s }

// activeSpace is a helper that returns a valid active space.
func activeSpace() *domain.Space {
	return &domain.Space{ID: 1, Name: "Cancha 1", Type: domain.SpaceTypePadel, IsActive: true}
}

// --- SpaceService.Create ---

func TestSpaceService_Create_Validations(t *testing.T) {
	cases := []struct {
		name         string
		spaceName    string
		spaceType    string
		price        float64
		wantErr      error
	}{
		{"empty name", "", domain.SpaceTypePadel, 100, ErrSpaceNameRequired},
		{"whitespace name", "   ", domain.SpaceTypePadel, 100, ErrSpaceNameRequired},
		{"invalid type", "Cancha 1", "voley", 100, ErrInvalidSpaceType},
		{"zero price", "Cancha 1", domain.SpaceTypePadel, 0, ErrInvalidSpacePrice},
		{"negative price", "Cancha 1", domain.SpaceTypePadel, -50, ErrInvalidSpacePrice},
	}

	svc := NewSpaceService(&mockSpaceRepo{})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.spaceName, tc.spaceType, nil, tc.price)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSpaceService_Create_Success(t *testing.T) {
	repo := &mockSpaceRepo{
		createFn: func(_ context.Context, _ *domain.Space) (int64, error) { return 42, nil },
	}
	svc := NewSpaceService(repo)

	id, err := svc.Create(context.Background(), "Cancha 1", domain.SpaceTypePadel, nil, 1500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("got id %d, want 42", id)
	}
}

func TestSpaceService_Create_EmptyDescription_TreatedAsNil(t *testing.T) {
	var captured *domain.Space
	repo := &mockSpaceRepo{
		createFn: func(_ context.Context, s *domain.Space) (int64, error) {
			captured = s
			return 1, nil
		},
	}
	svc := NewSpaceService(repo)

	_, err := svc.Create(context.Background(), "Cancha", domain.SpaceTypeFutbol, ptr("  "), 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Description != nil {
		t.Error("expected description to be nil after trimming whitespace")
	}
}

// --- SpaceService.GetByID ---

func TestSpaceService_GetByID_InvalidID(t *testing.T) {
	svc := NewSpaceService(&mockSpaceRepo{})

	for _, id := range []int64{0, -1, -100} {
		_, err := svc.GetByID(context.Background(), id)
		if !errors.Is(err, ErrInvalidSpaceID) {
			t.Errorf("id=%d: got %v, want ErrInvalidSpaceID", id, err)
		}
	}
}

func TestSpaceService_GetByID_NotFound(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := NewSpaceService(repo)

	_, err := svc.GetByID(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_GetByID_Success(t *testing.T) {
	want := activeSpace()
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return want, nil },
	}
	svc := NewSpaceService(repo)

	got, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got id %d, want %d", got.ID, want.ID)
	}
}

// --- SpaceService.GetSlots ---

func TestSpaceService_GetSlots_InvalidID(t *testing.T) {
	svc := NewSpaceService(&mockSpaceRepo{})
	_, err := svc.GetSlots(context.Background(), 0)
	if !errors.Is(err, ErrInvalidSpaceID) {
		t.Errorf("got %v, want ErrInvalidSpaceID", err)
	}
}

func TestSpaceService_GetSlots_SpaceNotFound(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := NewSpaceService(repo)

	_, err := svc.GetSlots(context.Background(), 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_GetSlots_Success(t *testing.T) {
	wantSlots := []domain.SpaceSlot{{ID: 10, SpaceID: 1, Label: "Mañana"}}
	repo := &mockSpaceRepo{
		getByIDFn:           func(_ context.Context, _ int64) (*domain.Space, error) { return activeSpace(), nil },
		getSlotsBySpaceIDFn: func(_ context.Context, _ int64) ([]domain.SpaceSlot, error) { return wantSlots, nil },
	}
	svc := NewSpaceService(repo)

	slots, err := svc.GetSlots(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 1 || slots[0].ID != 10 {
		t.Errorf("unexpected slots: %+v", slots)
	}
}

// --- SpaceService.CreateSlot ---

func TestSpaceService_CreateSlot_InvalidSpaceID(t *testing.T) {
	svc := NewSpaceService(&mockSpaceRepo{})
	_, err := svc.CreateSlot(context.Background(), 0, "Mañana", nil, nil, nil)
	if !errors.Is(err, ErrInvalidSpaceID) {
		t.Errorf("got %v, want ErrInvalidSpaceID", err)
	}
}

func TestSpaceService_CreateSlot_EmptyLabel(t *testing.T) {
	svc := NewSpaceService(&mockSpaceRepo{})
	_, err := svc.CreateSlot(context.Background(), 1, "  ", nil, nil, nil)
	if !errors.Is(err, ErrSlotLabelRequired) {
		t.Errorf("got %v, want ErrSlotLabelRequired", err)
	}
}

func TestSpaceService_CreateSlot_SpaceNotFound(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := NewSpaceService(repo)

	_, err := svc.CreateSlot(context.Background(), 1, "Mañana", nil, nil, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_CreateSlot_SpaceInactive(t *testing.T) {
	inactive := &domain.Space{ID: 1, IsActive: false}
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return inactive, nil },
	}
	svc := NewSpaceService(repo)

	_, err := svc.CreateSlot(context.Background(), 1, "Mañana", nil, nil, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_CreateSlot_Success(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn:    func(_ context.Context, _ int64) (*domain.Space, error) { return activeSpace(), nil },
		createSlotFn: func(_ context.Context, _ *domain.SpaceSlot) (int64, error) { return 7, nil },
	}
	svc := NewSpaceService(repo)

	id, err := svc.CreateSlot(context.Background(), 1, "Tarde", nil, ptr("14:00"), ptr("18:00"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("got id %d, want 7", id)
	}
}

// --- SpaceService.Update ---

func TestSpaceService_Update_Validations(t *testing.T) {
	activeRepo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return activeSpace(), nil },
	}

	cases := []struct {
		name    string
		id      int64
		svcName string
		price   float64
		wantErr error
	}{
		{"invalid id", 0, "Cancha", 100, ErrInvalidSpaceID},
		{"empty name", 1, "", 100, ErrSpaceNameRequired},
		{"whitespace name", 1, "  ", 100, ErrSpaceNameRequired},
		{"zero price", 1, "Cancha", 0, ErrInvalidSpacePrice},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewSpaceService(activeRepo)
			err := svc.Update(context.Background(), tc.id, tc.svcName, nil, tc.price)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSpaceService_Update_SpaceNotFound(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := NewSpaceService(repo)

	err := svc.Update(context.Background(), 1, "Cancha", nil, 100)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_Update_Success(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return activeSpace(), nil },
		updateFn:  func(_ context.Context, _ int64, _ string, _ *string, _ float64) error { return nil },
	}
	svc := NewSpaceService(repo)

	err := svc.Update(context.Background(), 1, "Cancha Nueva", ptr("Descripción"), 2000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SpaceService.Deactivate ---

func TestSpaceService_Deactivate_InvalidID(t *testing.T) {
	svc := NewSpaceService(&mockSpaceRepo{})
	err := svc.Deactivate(context.Background(), 0)
	if !errors.Is(err, ErrInvalidSpaceID) {
		t.Errorf("got %v, want ErrInvalidSpaceID", err)
	}
}

func TestSpaceService_Deactivate_SpaceNotFound(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return nil, nil },
	}
	svc := NewSpaceService(repo)

	err := svc.Deactivate(context.Background(), 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_Deactivate_AlreadyInactive(t *testing.T) {
	inactive := &domain.Space{ID: 1, IsActive: false}
	repo := &mockSpaceRepo{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) { return inactive, nil },
	}
	svc := NewSpaceService(repo)

	err := svc.Deactivate(context.Background(), 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestSpaceService_Deactivate_Success(t *testing.T) {
	repo := &mockSpaceRepo{
		getByIDFn:    func(_ context.Context, _ int64) (*domain.Space, error) { return activeSpace(), nil },
		deactivateFn: func(_ context.Context, _ int64) error { return nil },
	}
	svc := NewSpaceService(repo)

	err := svc.Deactivate(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
