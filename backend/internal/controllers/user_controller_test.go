package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/service"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockUserService implements userServiceI for controller tests.
type mockUserService struct {
	registerFn    func(ctx context.Context, name, phone, password string, email *string) (*domain.User, string, error)
	createStaffFn func(ctx context.Context, name, phone, password, role string, email *string) (int64, error)
	loginFn       func(ctx context.Context, phone, password string) (*domain.User, string, error)
	listUsersFn   func(ctx context.Context) ([]domain.User, error)
	updateRoleFn  func(ctx context.Context, id int64, newRole string) error
	deleteUserFn  func(ctx context.Context, targetID, requesterID int64) error
}

func (m *mockUserService) Register(ctx context.Context, name, phone, password string, email *string) (*domain.User, string, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, name, phone, password, email)
	}
	return &domain.User{ID: 1, Name: name, Phone: phone, Email: email, Role: domain.RoleCustomer}, "tok", nil
}

func (m *mockUserService) CreateStaff(ctx context.Context, name, phone, password, role string, email *string) (int64, error) {
	if m.createStaffFn != nil {
		return m.createStaffFn(ctx, name, phone, password, role, email)
	}
	return 1, nil
}

func (m *mockUserService) Login(ctx context.Context, phone, password string) (*domain.User, string, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, phone, password)
	}
	return &domain.User{ID: 1, Phone: phone, Role: domain.RoleCustomer}, "tok", nil
}

func (m *mockUserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx)
	}
	return nil, nil
}

func (m *mockUserService) UpdateRole(ctx context.Context, id int64, newRole string) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, id, newRole)
	}
	return nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, targetID, requesterID int64) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, targetID, requesterID)
	}
	return nil
}

func newAuthRouter(svc userServiceI) *gin.Engine {
	r := gin.New()
	ctrl := NewAuthController(svc)
	r.POST("/auth/register", ctrl.Register)
	r.POST("/auth/login", ctrl.Login)
	r.POST("/admin/users", ctrl.CreateStaff)
	return r
}

// --- Register ---

func TestAuthController_Register_InvalidJSON(t *testing.T) {
	r := newAuthRouter(&mockUserService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{bad json}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestAuthController_Register_PhoneConflict(t *testing.T) {
	svc := &mockUserService{
		registerFn: func(_ context.Context, _, _, _ string, _ *string) (*domain.User, string, error) {
			return nil, "", service.ErrPhoneAlreadyExists
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"name": "Juan", "phone": "123", "password": "secret1", "email": "juan@example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", w.Code)
	}
}

func TestAuthController_Register_Success(t *testing.T) {
	svc := &mockUserService{
		registerFn: func(_ context.Context, _, _, _ string, _ *string) (*domain.User, string, error) {
			email := "juan@example.com"
			return &domain.User{ID: 5, Name: "Juan", Phone: "123", Email: &email, Role: domain.RoleCustomer}, "jwt-token", nil
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"name": "Juan", "phone": "123", "password": "secret1", "email": "juan@example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Errorf("got %d, want 201", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] != "jwt-token" {
		t.Errorf("expected token in response, got %v", resp)
	}
	user, _ := resp["user"].(map[string]any)
	if user["email"] != "juan@example.com" {
		t.Errorf("expected email in user response, got %v", user)
	}
}

func TestAuthController_Register_MissingEmail(t *testing.T) {
	r := newAuthRouter(&mockUserService{})
	body, _ := json.Marshal(map[string]string{"name": "Juan", "phone": "123", "password": "secret1"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestAuthController_Register_ErrorStatus(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{service.ErrNameRequired, http.StatusBadRequest},
		{service.ErrPhoneRequired, http.StatusBadRequest},
		{service.ErrPasswordRequired, http.StatusBadRequest},
		{service.ErrPasswordTooShort, http.StatusBadRequest},
		{service.ErrEmailRequired, http.StatusBadRequest},
		{service.ErrInvalidEmail, http.StatusBadRequest},
		{service.ErrInvalidPhone, http.StatusBadRequest},
		{service.ErrPhoneAlreadyExists, http.StatusConflict},
		{service.ErrEmailAlreadyExists, http.StatusConflict},
		{errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			svc := &mockUserService{
				registerFn: func(_ context.Context, _, _, _ string, _ *string) (*domain.User, string, error) {
					return nil, "", tc.err
				},
			}
			r := newAuthRouter(svc)
			body, _ := json.Marshal(map[string]string{"name": "Juan", "phone": "123", "password": "secret1", "email": "juan@example.com"})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
			if w.Code != tc.want {
				t.Errorf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

// --- Login ---

func TestAuthController_Login_InvalidJSON(t *testing.T) {
	r := newAuthRouter(&mockUserService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{bad}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	svc := &mockUserService{
		loginFn: func(_ context.Context, _, _ string) (*domain.User, string, error) {
			return nil, "", service.ErrInvalidCredentials
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"phone": "123", "password": "wrong"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuthController_Login_UserInactive(t *testing.T) {
	svc := &mockUserService{
		loginFn: func(_ context.Context, _, _ string) (*domain.User, string, error) {
			return nil, "", service.ErrUserInactive
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"phone": "123", "password": "secret1"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuthController_Login_Success(t *testing.T) {
	svc := &mockUserService{
		loginFn: func(_ context.Context, _, _ string) (*domain.User, string, error) {
			return &domain.User{ID: 1, Name: "Juan", Phone: "123", Role: domain.RoleCustomer}, "jwt-token", nil
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"phone": "123", "password": "secret1"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] != "jwt-token" {
		t.Errorf("expected token in response, got %v", resp)
	}
}

// --- CreateStaff ---

func TestAuthController_CreateStaff_InvalidJSON(t *testing.T) {
	r := newAuthRouter(&mockUserService{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBufferString(`{bad}`)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestAuthController_CreateStaff_InvalidRole(t *testing.T) {
	svc := &mockUserService{
		createStaffFn: func(_ context.Context, _, _, _, _ string, _ *string) (int64, error) {
			return 0, service.ErrInvalidRole
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"name": "Ana", "phone": "456", "password": "secret1", "role": "superadmin"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
}

func TestAuthController_CreateStaff_Success(t *testing.T) {
	svc := &mockUserService{
		createStaffFn: func(_ context.Context, _, _, _, _ string, _ *string) (int64, error) {
			return 10, nil
		},
	}
	r := newAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"name": "Ana", "phone": "456", "password": "secret1", "role": "receptionist"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Errorf("got %d, want 201", w.Code)
	}
}
