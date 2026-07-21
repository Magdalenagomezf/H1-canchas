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
	"H1-canchas/pkg/mercadopago"

	"github.com/gin-gonic/gin"
)

// mockPaymentService implements paymentServiceI for controller tests.
type mockPaymentService struct {
	generateForBookingFn func(ctx context.Context, bookingID int64, kind string, requesterID int64, requesterRole string) (*mercadopago.PreferenceResponse, error)
	handleWebhookFn      func(ctx context.Context, xSignature, xRequestID, dataID string) error
	markPaidManuallyFn   func(ctx context.Context, bookingID int64, kind, method string, recordedByID int64) error
	markUnpaidManuallyFn func(ctx context.Context, bookingID int64, kind string) error
}

func (m *mockPaymentService) GenerateForBooking(ctx context.Context, bookingID int64, kind string, requesterID int64, requesterRole string) (*mercadopago.PreferenceResponse, error) {
	if m.generateForBookingFn != nil {
		return m.generateForBookingFn(ctx, bookingID, kind, requesterID, requesterRole)
	}
	return &mercadopago.PreferenceResponse{ID: "pref-1", InitPoint: "https://mp/init", SandboxInitPoint: "https://mp/sandbox"}, nil
}

func (m *mockPaymentService) HandleWebhook(ctx context.Context, xSignature, xRequestID, dataID string) error {
	if m.handleWebhookFn != nil {
		return m.handleWebhookFn(ctx, xSignature, xRequestID, dataID)
	}
	return nil
}

func (m *mockPaymentService) MarkPaidManually(ctx context.Context, bookingID int64, kind, method string, recordedByID int64) error {
	if m.markPaidManuallyFn != nil {
		return m.markPaidManuallyFn(ctx, bookingID, kind, method, recordedByID)
	}
	return nil
}

func (m *mockPaymentService) MarkUnpaidManually(ctx context.Context, bookingID int64, kind string) error {
	if m.markUnpaidManuallyFn != nil {
		return m.markUnpaidManuallyFn(ctx, bookingID, kind)
	}
	return nil
}

func newPaymentRouter(svc paymentServiceI, userID int64, role string) *gin.Engine {
	r := gin.New()
	r.Use(withAuth(userID, role))
	ctrl := NewPaymentController(svc)
	r.POST("/bookings/:id/payments", ctrl.GeneratePreference)
	r.PATCH("/bookings/:id/payment", ctrl.MarkPayment)
	r.POST("/payments/webhook", ctrl.Webhook)
	return r
}

// --- GeneratePreference ---

func TestPaymentController_GeneratePreference_Success(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleCustomer)
	body, _ := json.Marshal(map[string]any{"kind": "deposit"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings/1/payments", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp["preference_id"] != "pref-1" {
		t.Errorf("got preference_id %v, want pref-1", resp["preference_id"])
	}
}

func TestPaymentController_GeneratePreference_InvalidID(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleCustomer)
	body, _ := json.Marshal(map[string]any{"kind": "deposit"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings/abc/payments", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_GeneratePreference_MalformedJSON(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleCustomer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings/1/payments", bytes.NewBufferString(`{bad}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_GeneratePreference_InvalidKindBinding(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleCustomer)
	body, _ := json.Marshal(map[string]any{"kind": "not-a-kind"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings/1/payments", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_GeneratePreference_ErrorMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"InvalidPaymentKind", service.ErrInvalidPaymentKind, http.StatusBadRequest},
		{"NotFound", service.ErrNotFound, http.StatusNotFound},
		{"Unauthorized", service.ErrUnauthorized, http.StatusForbidden},
		{"BookingExpired", service.ErrBookingExpired, http.StatusConflict},
		{"BookingNotPayable", service.ErrBookingNotPayable, http.StatusConflict},
		{"PaymentAlreadyCompleted", service.ErrPaymentAlreadyCompleted, http.StatusConflict},
		{"MercadoPagoUnavailable", service.ErrMercadoPagoUnavailable, http.StatusBadGateway},
		{"Unknown", context.DeadlineExceeded, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockPaymentService{
				generateForBookingFn: func(_ context.Context, _ int64, _ string, _ int64, _ string) (*mercadopago.PreferenceResponse, error) {
					return nil, tc.err
				},
			}
			r := newPaymentRouter(svc, 1, domain.RoleCustomer)
			body, _ := json.Marshal(map[string]any{"kind": "deposit"})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/bookings/1/payments", bytes.NewReader(body)))
			if w.Code != tc.wantStatus {
				t.Errorf("got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

// --- MarkPayment ---

func TestPaymentController_MarkPayment_PaidSuccess(t *testing.T) {
	var gotMethod string
	svc := &mockPaymentService{
		markPaidManuallyFn: func(_ context.Context, _ int64, _, method string, _ int64) error {
			gotMethod = method
			return nil
		},
	}
	r := newPaymentRouter(svc, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "paid", "payment_method": "efectivo"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	if gotMethod != "efectivo" {
		t.Errorf("got method %q, want efectivo", gotMethod)
	}
}

func TestPaymentController_MarkPayment_UnpaidSuccess(t *testing.T) {
	called := false
	svc := &mockPaymentService{
		markUnpaidManuallyFn: func(_ context.Context, _ int64, _ string) error {
			called = true
			return nil
		},
	}
	r := newPaymentRouter(svc, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "unpaid"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	if !called {
		t.Error("expected MarkUnpaidManually to be called")
	}
}

func TestPaymentController_MarkPayment_MissingMethodWhenPaid(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "paid"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400 (body=%s)", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != service.ErrInvalidPaymentMethod.Error() {
		t.Errorf("got error %q, want %q", resp["error"], service.ErrInvalidPaymentMethod.Error())
	}
}

func TestPaymentController_MarkPayment_EmptyMethodWhenPaid(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "paid", "payment_method": ""})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_MarkPayment_InvalidID(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "unpaid"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/abc/payment", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_MarkPayment_MalformedJSON(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleAdmin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewBufferString(`{bad}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_MarkPayment_InvalidStatusBinding(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "not-a-status"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestPaymentController_MarkPayment_ErrorMapping(t *testing.T) {
	svc := &mockPaymentService{
		markPaidManuallyFn: func(_ context.Context, _ int64, _, _ string, _ int64) error {
			return service.ErrPaymentAlreadyCompleted
		},
	}
	r := newPaymentRouter(svc, 1, domain.RoleAdmin)
	body, _ := json.Marshal(map[string]any{"kind": "deposit", "payment_status": "paid", "payment_method": "efectivo"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/bookings/1/payment", bytes.NewReader(body)))
	if w.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", w.Code)
	}
}

// --- Webhook ---

func TestPaymentController_Webhook_Success(t *testing.T) {
	r := newPaymentRouter(&mockPaymentService{}, 0, "")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/payments/webhook?data.id=123", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
}

func TestPaymentController_Webhook_OtherFailure(t *testing.T) {
	svc := &mockPaymentService{
		handleWebhookFn: func(_ context.Context, _, _, _ string) error {
			return context.DeadlineExceeded
		},
	}
	r := newPaymentRouter(svc, 0, "")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/payments/webhook?data.id=123", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", w.Code)
	}
}
