package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"H1-canchas/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// mockUserRepo implements UserRepo for unit tests.
type mockUserRepo struct {
	createFn      func(ctx context.Context, user *domain.User) (int64, error)
	findByPhoneFn func(ctx context.Context, phone string) (*domain.User, error)
	findByIDFn    func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return 1, nil
}

func (m *mockUserRepo) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	if m.findByPhoneFn != nil {
		return m.findByPhoneFn(ctx, phone)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func newUserSvc(repo *mockUserRepo) *UserService {
	return NewUserService(repo, "test-secret")
}

// --- UserService.Register ---

func TestUserService_Register_Validations(t *testing.T) {
	svc := newUserSvc(&mockUserRepo{})

	cases := []struct {
		name     string
		userName string
		phone    string
		password string
		wantErr  error
	}{
		{"empty name", "", "1122334455", "password1", ErrNameRequired},
		{"whitespace name", "   ", "1122334455", "password1", ErrNameRequired},
		{"empty phone", "Juan", "", "password1", ErrPhoneRequired},
		{"whitespace phone", "Juan", "   ", "password1", ErrPhoneRequired},
		{"empty password", "Juan", "1122334455", "", ErrPasswordRequired},
		{"short password", "Juan", "1122334455", "12345", ErrPasswordTooShort},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Register(context.Background(), tc.userName, tc.phone, tc.password, nil)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserService_Register_PhoneAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: 1}, nil
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", nil)
	if !errors.Is(err, ErrPhoneAlreadyExists) {
		t.Errorf("got %v, want ErrPhoneAlreadyExists", err)
	}
}

func TestUserService_Register_RepoFindError(t *testing.T) {
	dbErr := errors.New("db timeout")
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, dbErr
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("got %v, want wrapped dbErr", err)
	}
}

func TestUserService_Register_HappyPath(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, u *domain.User) (int64, error) {
			if u.Role != domain.RoleCustomer {
				t.Errorf("expected role %q, got %q", domain.RoleCustomer, u.Role)
			}
			if u.PasswordHash == "" {
				t.Error("expected non-empty password hash")
			}
			return 42, nil
		},
	}
	svc := newUserSvc(repo)

	user, token, err := svc.Register(context.Background(), "  Juan  ", "1122334455", "password1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 42 {
		t.Errorf("got ID %d, want 42", user.ID)
	}
	if user.Name != "Juan" {
		t.Errorf("got name %q, want %q", user.Name, "Juan")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

// --- UserService.CreateStaff ---

func TestUserService_CreateStaff_InvalidRole(t *testing.T) {
	svc := newUserSvc(&mockUserRepo{})

	_, err := svc.CreateStaff(context.Background(), "Juan", "1122334455", "password1", "superadmin", nil)
	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("got %v, want ErrInvalidRole", err)
	}
}

func TestUserService_CreateStaff_Validations(t *testing.T) {
	svc := newUserSvc(&mockUserRepo{})

	cases := []struct {
		name     string
		userName string
		phone    string
		password string
		wantErr  error
	}{
		{"empty name", "", "1122334455", "password1", ErrNameRequired},
		{"empty phone", "Juan", "", "password1", ErrPhoneRequired},
		{"empty password", "Juan", "1122334455", "", ErrPasswordRequired},
		{"short password", "Juan", "1122334455", "12345", ErrPasswordTooShort},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateStaff(context.Background(), tc.userName, tc.phone, tc.password, domain.RoleReceptionist, nil)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserService_CreateStaff_PhoneAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: 1}, nil
		},
	}
	svc := newUserSvc(repo)

	_, err := svc.CreateStaff(context.Background(), "Juan", "1122334455", "password1", domain.RoleReceptionist, nil)
	if !errors.Is(err, ErrPhoneAlreadyExists) {
		t.Errorf("got %v, want ErrPhoneAlreadyExists", err)
	}
}

func TestUserService_CreateStaff_HappyPath(t *testing.T) {
	cases := []struct {
		role string
	}{
		{domain.RoleReceptionist},
		{domain.RoleAdmin},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			repo := &mockUserRepo{
				createFn: func(_ context.Context, u *domain.User) (int64, error) {
					if u.Role != tc.role {
						t.Errorf("expected role %q, got %q", tc.role, u.Role)
					}
					return 10, nil
				},
			}
			svc := newUserSvc(repo)

			id, err := svc.CreateStaff(context.Background(), "Ana", "9988776655", "password1", tc.role, nil)
			if err != nil {
				t.Fatalf("role %s: unexpected error: %v", tc.role, err)
			}
			if id != 10 {
				t.Errorf("got ID %d, want 10", id)
			}
		})
	}
}

// --- UserService.Login ---

func TestUserService_Login_EmptyCredentials(t *testing.T) {
	svc := newUserSvc(&mockUserRepo{})

	cases := []struct{ phone, password string }{
		{"", "password1"},
		{"1122334455", ""},
		{"", ""},
	}

	for _, tc := range cases {
		_, _, err := svc.Login(context.Background(), tc.phone, tc.password)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("phone=%q password=%q: got %v, want ErrInvalidCredentials", tc.phone, tc.password, err)
		}
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) { return nil, nil },
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Login(context.Background(), "1122334455", "password1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestUserService_Login_UserInactive(t *testing.T) {
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: 1, IsActive: false}, nil
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Login(context.Background(), "1122334455", "password1")
	if !errors.Is(err, ErrUserInactive) {
		t.Errorf("got %v, want ErrUserInactive", err)
	}
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: 1, IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Login(context.Background(), "1122334455", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestUserService_Login_HappyPath(t *testing.T) {
	const password = "secret123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{
				ID:           7,
				Name:         "Juan",
				Phone:        "1122334455",
				Role:         domain.RoleCustomer,
				IsActive:     true,
				PasswordHash: string(hash),
			}, nil
		},
	}
	svc := newUserSvc(repo)

	user, token, err := svc.Login(context.Background(), "1122334455", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 7 {
		t.Errorf("got ID %d, want 7", user.ID)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if strings.Contains(token, password) {
		t.Error("token must not contain plain password")
	}
}
