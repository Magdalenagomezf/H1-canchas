package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

// newMeRouter wires GET/PATCH /me behind a fake authenticated user.
func newMeRouter(svc userServiceI, userID int64) *gin.Engine {
	r := gin.New()
	r.Use(withAuth(userID, domain.RoleCustomer))
	ctrl := NewAuthController(svc)
	r.GET("/me", ctrl.GetMe)
	r.PATCH("/me", ctrl.UpdateMe)
	return r
}

func doJSON(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, &buf))
	return w
}

// --- POST /auth/google ---

func TestAuthController_GoogleLogin_ErrorStatus(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{service.ErrGoogleNotConfigured, http.StatusServiceUnavailable},
		{service.ErrInvalidGoogleToken, http.StatusUnauthorized},
		{service.ErrUserInactive, http.StatusUnauthorized},
		{service.ErrGoogleStaffAccount, http.StatusForbidden},
		{service.ErrGoogleAccountConflict, http.StatusConflict},
		{service.ErrEmailAlreadyExists, http.StatusConflict},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			svc := &mockUserService{
				loginWithGoogleFn: func(_ context.Context, _ string) (*domain.User, string, error) {
					return nil, "", tc.err
				},
			}
			w := doJSON(newAuthRouter(svc), http.MethodPost, "/auth/google", map[string]string{"credential": "x"})
			if w.Code != tc.want {
				t.Errorf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestAuthController_GoogleLogin_NotConfiguredMessage(t *testing.T) {
	svc := &mockUserService{
		loginWithGoogleFn: func(_ context.Context, _ string) (*domain.User, string, error) {
			return nil, "", service.ErrGoogleNotConfigured
		},
	}
	w := doJSON(newAuthRouter(svc), http.MethodPost, "/auth/google", map[string]string{"credential": "x"})

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "El inicio de sesión con Google no está configurado" {
		t.Errorf("unexpected message %q", resp["error"])
	}
}

func TestAuthController_GoogleLogin_MissingCredential(t *testing.T) {
	w := doJSON(newAuthRouter(&mockUserService{}), http.MethodPost, "/auth/google", map[string]string{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestAuthController_GoogleLogin_Success_PhoneIsNull(t *testing.T) {
	svc := &mockUserService{
		loginWithGoogleFn: func(_ context.Context, cred string) (*domain.User, string, error) {
			if cred != "id-token" {
				t.Errorf("credential = %q, want id-token", cred)
			}
			return &domain.User{ID: 3, Name: "Ana", Email: ptr("ana@example.com"), Role: domain.RoleCustomer}, "jwt", nil
		},
	}
	w := doJSON(newAuthRouter(svc), http.MethodPost, "/auth/google", map[string]string{"credential": "id-token"})
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] != "jwt" {
		t.Errorf("expected token, got %v", resp)
	}
	user, _ := resp["user"].(map[string]any)
	phone, present := user["phone"]
	if !present || phone != nil {
		t.Errorf("phone must be present and null, got present=%v value=%v", present, phone)
	}
}

// --- GET /me ---

func TestAuthController_GetMe_UsesTokenUser(t *testing.T) {
	svc := &mockUserService{
		getProfileFn: func(_ context.Context, id int64) (*domain.User, error) {
			return &domain.User{ID: id, Name: "Ana", Phone: ptr("+5493834123456"), Role: domain.RoleReceptionist}, nil
		},
	}
	w := doJSON(newMeRouter(svc, 9), http.MethodGet, "/me", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"] != float64(9) || resp["phone"] != "+5493834123456" {
		t.Errorf("unexpected body %v", resp)
	}
}

func TestAuthController_GetMe_NotFound(t *testing.T) {
	svc := &mockUserService{
		getProfileFn: func(_ context.Context, _ int64) (*domain.User, error) { return nil, service.ErrNotFound },
	}
	w := doJSON(newMeRouter(svc, 9), http.MethodGet, "/me", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", w.Code)
	}
}

// --- PATCH /me ---

func TestAuthController_UpdateMe_AbsentFieldsArePassedAsNil(t *testing.T) {
	var gotName, gotEmail, gotPhone *string
	svc := &mockUserService{
		updateProfileFn: func(_ context.Context, id int64, name, email, phone *string) (*domain.User, error) {
			gotName, gotEmail, gotPhone = name, email, phone
			return &domain.User{ID: id, Name: "Ana", Role: domain.RoleCustomer}, nil
		},
	}
	w := doJSON(newMeRouter(svc, 4), http.MethodPatch, "/me", map[string]string{"phone": "3834123456"})
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	if gotName != nil || gotEmail != nil {
		t.Errorf("absent name/email must be nil, got %v / %v", gotName, gotEmail)
	}
	if gotPhone == nil || *gotPhone != "3834123456" {
		t.Errorf("phone = %v, want 3834123456", gotPhone)
	}
}

func TestAuthController_UpdateMe_ErrorStatus(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{service.ErrNameRequired, http.StatusBadRequest},
		{service.ErrNameTooLong, http.StatusBadRequest},
		{service.ErrEmailRequired, http.StatusBadRequest},
		{service.ErrInvalidEmail, http.StatusBadRequest},
		{service.ErrInvalidPhone, http.StatusBadRequest},
		{service.ErrEmailAlreadyExists, http.StatusConflict},
		{service.ErrPhoneAlreadyExists, http.StatusConflict},
		{service.ErrNotFound, http.StatusNotFound},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			svc := &mockUserService{
				updateProfileFn: func(_ context.Context, _ int64, _, _, _ *string) (*domain.User, error) {
					return nil, tc.err
				},
			}
			w := doJSON(newMeRouter(svc, 4), http.MethodPatch, "/me", map[string]string{"name": "x"})
			if w.Code != tc.want {
				t.Errorf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestAuthController_UpdateMe_InvalidJSON(t *testing.T) {
	r := newMeRouter(&mockUserService{}, 4)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/me", bytes.NewBufferString(`{bad`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

// --- POST /bookings: phone required ---

func TestBookingController_Create_ProfilePhoneMissing(t *testing.T) {
	svc := &mockBookingService{
		createFn: func(_ context.Context, _ int64, _, _ int64, _ time.Time) (*domain.Booking, error) {
			return nil, service.ErrProfilePhoneMissing
		},
	}
	r := newBookingRouter(svc, 5, domain.RoleCustomer)
	w := doJSON(r, http.MethodPost, "/bookings", map[string]any{"space_id": 1, "slot_id": 1, "booking_date": "2030-01-01"})

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422", w.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "phone_required" {
		t.Errorf("code = %q, want phone_required", resp["code"])
	}
	if resp["error"] != service.ErrProfilePhoneMissing.Error() {
		t.Errorf("error = %q", resp["error"])
	}
}
