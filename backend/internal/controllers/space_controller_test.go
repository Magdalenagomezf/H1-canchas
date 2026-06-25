package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

// mockSpaceService implements spaceServiceI for controller tests.
type mockSpaceService struct {
	createFn     func(ctx context.Context, name, spaceType string, description *string, pricePerSlot float64) (int64, error)
	getAllFn      func(ctx context.Context) ([]domain.Space, error)
	getByIDFn    func(ctx context.Context, id int64) (*domain.Space, error)
	getSlotsFn   func(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error)
	createSlotFn func(ctx context.Context, spaceID int64, label string, description *string, startTime, endTime *string) (int64, error)
	updateFn     func(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error
	deactivateFn func(ctx context.Context, id int64) error
}

func (m *mockSpaceService) Create(ctx context.Context, name, spaceType string, description *string, pricePerSlot float64) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, name, spaceType, description, pricePerSlot)
	}
	return 1, nil
}
func (m *mockSpaceService) GetAll(ctx context.Context) ([]domain.Space, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}
func (m *mockSpaceService) GetByID(ctx context.Context, id int64) (*domain.Space, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockSpaceService) GetSlots(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error) {
	if m.getSlotsFn != nil {
		return m.getSlotsFn(ctx, spaceID)
	}
	return nil, nil
}
func (m *mockSpaceService) CreateSlot(ctx context.Context, spaceID int64, label string, description *string, startTime, endTime *string) (int64, error) {
	if m.createSlotFn != nil {
		return m.createSlotFn(ctx, spaceID, label, description, startTime, endTime)
	}
	return 1, nil
}
func (m *mockSpaceService) Update(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, name, description, pricePerSlot)
	}
	return nil
}
func (m *mockSpaceService) Deactivate(ctx context.Context, id int64) error {
	if m.deactivateFn != nil {
		return m.deactivateFn(ctx, id)
	}
	return nil
}

func newSpaceRouter(svc spaceServiceI) *gin.Engine {
	r := gin.New()
	ctrl := NewSpaceController(svc)
	r.GET("/spaces", ctrl.GetAll)
	r.GET("/spaces/:id", ctrl.GetByID)
	r.GET("/spaces/:id/slots", ctrl.GetSlots)
	r.POST("/spaces", ctrl.Create)
	r.DELETE("/spaces/:id", ctrl.Deactivate)
	return r
}

// --- GetAll ---

func TestSpaceController_GetAll_ReturnsEmptyList(t *testing.T) {
	r := newSpaceRouter(&mockSpaceService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/spaces", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	var resp []any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %v", resp)
	}
}

func TestSpaceController_GetAll_ReturnsList(t *testing.T) {
	svc := &mockSpaceService{
		getAllFn: func(_ context.Context) ([]domain.Space, error) {
			return []domain.Space{{ID: 1, Name: "Cancha 1"}, {ID: 2, Name: "Cancha 2"}}, nil
		},
	}
	r := newSpaceRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/spaces", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	var resp []any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("got %d items, want 2", len(resp))
	}
}

// --- GetByID ---

func TestSpaceController_GetByID_InvalidID(t *testing.T) {
	r := newSpaceRouter(&mockSpaceService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/spaces/abc", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestSpaceController_GetByID_NotFound(t *testing.T) {
	svc := &mockSpaceService{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return nil, service.ErrNotFound
		},
	}
	r := newSpaceRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/spaces/99", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", w.Code)
	}
}

func TestSpaceController_GetByID_Success(t *testing.T) {
	svc := &mockSpaceService{
		getByIDFn: func(_ context.Context, _ int64) (*domain.Space, error) {
			return &domain.Space{ID: 1, Name: "Cancha Padel", Type: domain.SpaceTypePadel, IsActive: true}, nil
		},
	}
	r := newSpaceRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/spaces/1", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != "Cancha Padel" {
		t.Errorf("expected name %q, got %v", "Cancha Padel", resp["name"])
	}
}

// --- Create ---

func TestSpaceController_Create_InvalidJSON(t *testing.T) {
	r := newSpaceRouter(&mockSpaceService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/spaces", bytes.NewBufferString(`{bad}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestSpaceController_Create_ValidationError(t *testing.T) {
	svc := &mockSpaceService{
		createFn: func(_ context.Context, _, _ string, _ *string, _ float64) (int64, error) {
			return 0, service.ErrInvalidSpaceType
		},
	}
	r := newSpaceRouter(svc)
	body, _ := json.Marshal(map[string]any{"name": "X", "type": "invalido", "price_per_slot": 100.0})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/spaces", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestSpaceController_Create_Success(t *testing.T) {
	svc := &mockSpaceService{
		createFn: func(_ context.Context, _, _ string, _ *string, _ float64) (int64, error) {
			return 7, nil
		},
	}
	r := newSpaceRouter(svc)
	body, _ := json.Marshal(map[string]any{"name": "Cancha", "type": "padel", "price_per_slot": 1500.0})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/spaces", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Errorf("got %d, want 201", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"] != float64(7) {
		t.Errorf("expected id 7, got %v", resp["id"])
	}
}

// --- Deactivate ---

func TestSpaceController_Deactivate_NotFound(t *testing.T) {
	svc := &mockSpaceService{
		deactivateFn: func(_ context.Context, _ int64) error {
			return service.ErrNotFound
		},
	}
	r := newSpaceRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/spaces/99", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", w.Code)
	}
}

func TestSpaceController_Deactivate_Success(t *testing.T) {
	r := newSpaceRouter(&mockSpaceService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/spaces/1", nil))
	if w.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", w.Code)
	}
}
