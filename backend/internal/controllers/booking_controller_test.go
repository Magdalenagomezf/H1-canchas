package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

// mockBookingService implements bookingServiceI for controller tests.
type mockBookingService struct {
	createFn        func(ctx context.Context, customerUserID int64, spaceID, slotID int64, date time.Time) (*domain.Booking, error)
	createManualFn  func(ctx context.Context, createdBy int64, spaceID, slotID int64, date time.Time, customerName, customerPhone string) (*domain.Booking, error)
	getMyBookingsFn func(ctx context.Context, userID int64) ([]domain.BookingDetail, error)
	getAllFn         func(ctx context.Context, date *time.Time) ([]domain.BookingDetail, error)
	cancelFn        func(ctx context.Context, bookingID, requesterID int64, requesterRole string) error
	getByIDFn       func(ctx context.Context, bookingID, requesterID int64, requesterRole string) (*domain.BookingDetail, error)
}

func (m *mockBookingService) Create(ctx context.Context, customerUserID int64, spaceID, slotID int64, date time.Time) (*domain.Booking, error) {
	if m.createFn != nil {
		return m.createFn(ctx, customerUserID, spaceID, slotID, date)
	}
	return &domain.Booking{ID: 1}, nil
}
func (m *mockBookingService) CreateManual(ctx context.Context, createdBy int64, spaceID, slotID int64, date time.Time, customerName, customerPhone string) (*domain.Booking, error) {
	if m.createManualFn != nil {
		return m.createManualFn(ctx, createdBy, spaceID, slotID, date, customerName, customerPhone)
	}
	return &domain.Booking{ID: 1}, nil
}
func (m *mockBookingService) GetMyBookings(ctx context.Context, userID int64) ([]domain.BookingDetail, error) {
	if m.getMyBookingsFn != nil {
		return m.getMyBookingsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockBookingService) GetAll(ctx context.Context, date *time.Time) ([]domain.BookingDetail, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, date)
	}
	return nil, nil
}
func (m *mockBookingService) Cancel(ctx context.Context, bookingID, requesterID int64, requesterRole string) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, bookingID, requesterID, requesterRole)
	}
	return nil
}
func (m *mockBookingService) GetByID(ctx context.Context, bookingID, requesterID int64, requesterRole string) (*domain.BookingDetail, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, bookingID, requesterID, requesterRole)
	}
	return &domain.BookingDetail{ID: bookingID}, nil
}

// withAuth injects user_id and role into the Gin context (simulates JWT middleware).
func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func newBookingRouter(svc bookingServiceI, userID int64, role string) *gin.Engine {
	r := gin.New()
	r.Use(withAuth(userID, role))
	ctrl := NewBookingController(svc)
	r.POST("/bookings", ctrl.Create)
	r.POST("/bookings/manual", ctrl.CreateManual)
	r.GET("/bookings", ctrl.GetMyBookings)
	r.GET("/bookings/:id", ctrl.GetByID)
	r.PATCH("/bookings/:id/cancel", ctrl.Cancel)
	return r
}

// --- Cancel ---

func TestBookingController_Cancel_InvalidID(t *testing.T) {
	r := newBookingRouter(&mockBookingService{}, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/abc/cancel", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestBookingController_Cancel_NotFound(t *testing.T) {
	svc := &mockBookingService{
		cancelFn: func(_ context.Context, _, _ int64, _ string) error { return service.ErrNotFound },
	}
	r := newBookingRouter(svc, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/99/cancel", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", w.Code)
	}
}

func TestBookingController_Cancel_Forbidden(t *testing.T) {
	svc := &mockBookingService{
		cancelFn: func(_ context.Context, _, _ int64, _ string) error { return service.ErrUnauthorized },
	}
	r := newBookingRouter(svc, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/cancel", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", w.Code)
	}
}

func TestBookingController_Cancel_AlreadyCancelled(t *testing.T) {
	svc := &mockBookingService{
		cancelFn: func(_ context.Context, _, _ int64, _ string) error { return service.ErrBookingAlreadyCancelled },
	}
	r := newBookingRouter(svc, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/cancel", nil))
	if w.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", w.Code)
	}
}

func TestBookingController_Cancel_RequiresStaffAfterPayment(t *testing.T) {
	svc := &mockBookingService{
		cancelFn: func(_ context.Context, _, _ int64, _ string) error { return service.ErrCancelRequiresStaffAfterPayment },
	}
	r := newBookingRouter(svc, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/cancel", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", w.Code)
	}
}

func TestBookingController_Cancel_Success(t *testing.T) {
	r := newBookingRouter(&mockBookingService{}, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/cancel", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}

// --- GetMyBookings ---

func TestBookingController_GetMyBookings_NoAuth(t *testing.T) {
	r := gin.New()
	ctrl := NewBookingController(&mockBookingService{})
	r.GET("/bookings", ctrl.GetMyBookings)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/bookings", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestBookingController_GetMyBookings_CustomerGetsOwnBookings(t *testing.T) {
	svc := &mockBookingService{
		getMyBookingsFn: func(_ context.Context, userID int64) ([]domain.BookingDetail, error) {
			return []domain.BookingDetail{{ID: 1}, {ID: 2}}, nil
		},
	}
	r := newBookingRouter(svc, 7, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/bookings", nil))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	var resp []any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("got %d items, want 2", len(resp))
	}
}

// --- Create ---

func TestBookingController_Create_InvalidJSON(t *testing.T) {
	r := newBookingRouter(&mockBookingService{}, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBufferString(`{bad}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestBookingController_Create_InvalidDateFormat(t *testing.T) {
	r := newBookingRouter(&mockBookingService{}, 1, domain.RoleCustomer)
	body, _ := json.Marshal(map[string]any{"space_id": 1, "slot_id": 1, "booking_date": "15/06/2026"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestBookingController_Create_SlotNotAvailable(t *testing.T) {
	svc := &mockBookingService{
		createFn: func(_ context.Context, _ int64, _, _ int64, _ time.Time) (*domain.Booking, error) {
			return nil, service.ErrSlotNotAvailable
		},
	}
	r := newBookingRouter(svc, 1, domain.RoleCustomer)
	body, _ := json.Marshal(map[string]any{"space_id": 1, "slot_id": 1, "booking_date": "2027-01-15"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body)))
	if w.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", w.Code)
	}
}

func TestBookingController_Create_Success(t *testing.T) {
	svc := &mockBookingService{
		createFn: func(_ context.Context, _ int64, _, _ int64, _ time.Time) (*domain.Booking, error) {
			return &domain.Booking{ID: 42}, nil
		},
		getByIDFn: func(_ context.Context, bookingID, _ int64, _ string) (*domain.BookingDetail, error) {
			return &domain.BookingDetail{ID: bookingID, Status: domain.BookingStatusPending}, nil
		},
	}
	r := newBookingRouter(svc, 1, domain.RoleCustomer)
	body, _ := json.Marshal(map[string]any{"space_id": 1, "slot_id": 1, "booking_date": "2027-01-15"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Errorf("got %d, want 201", w.Code)
	}
}
