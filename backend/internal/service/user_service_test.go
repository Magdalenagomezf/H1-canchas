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
	createFn       func(ctx context.Context, user *domain.User) (int64, error)
	findByPhoneFn  func(ctx context.Context, phone string) (*domain.User, error)
	findByEmailFn  func(ctx context.Context, email string) (*domain.User, error)
	findByIDFn     func(ctx context.Context, id int64) (*domain.User, error)
	getAllFn        func(ctx context.Context) ([]domain.User, error)
	updateRoleFn   func(ctx context.Context, id int64, role string) error
	hasBookingsFn  func(ctx context.Context, id int64) (bool, error)
	deleteFn       func(ctx context.Context, id int64) error
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

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]domain.User, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, id, role)
	}
	return nil
}

func (m *mockUserRepo) HasBookings(ctx context.Context, id int64) (bool, error) {
	if m.hasBookingsFn != nil {
		return m.hasBookingsFn(ctx, id)
	}
	return false, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
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
			_, _, err := svc.Register(context.Background(), tc.userName, tc.phone, tc.password, ptr("juan@example.com"))
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

	_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", ptr("juan@example.com"))
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

	_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", ptr("juan@example.com"))
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

	user, token, err := svc.Register(context.Background(), "  Juan  ", "1122334455", "password1", ptr("juan@example.com"))
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

			id, err := svc.CreateStaff(context.Background(), "Ana", "3834123456", "password1", tc.role, nil)
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

// --- Contact data normalization ---

func TestUserService_Register_StoresNormalizedContact(t *testing.T) {
	var stored *domain.User
	repo := &mockUserRepo{
		createFn: func(_ context.Context, u *domain.User) (int64, error) {
			stored = u
			return 1, nil
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Register(context.Background(), "Juan", "0383 15-412-3456", "password1", ptr("  Juan.Perez@Example.COM "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored.Email == nil || *stored.Email != "juan.perez@example.com" {
		t.Errorf("got email %v, want juan.perez@example.com", stored.Email)
	}
	if stored.Phone != "+5493834123456" {
		t.Errorf("got phone %q, want +5493834123456", stored.Phone)
	}
}

func TestUserService_Register_ChecksNormalizedValuesForDuplicates(t *testing.T) {
	var gotPhone, gotEmail string
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, phone string) (*domain.User, error) {
			gotPhone = phone
			return nil, nil
		},
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			gotEmail = email
			return nil, nil
		},
	}
	svc := newUserSvc(repo)

	if _, _, err := svc.Register(context.Background(), "Juan", "3834123456", "password1", ptr("A@B.com")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPhone != "+5493834123456" {
		t.Errorf("FindByPhone got %q, want E.164 value", gotPhone)
	}
	if gotEmail != "a@b.com" {
		t.Errorf("FindByEmail got %q, want lowercase value", gotEmail)
	}
}

func TestUserService_Register_EmailAlreadyExists(t *testing.T) {
	// The stored email is lowercase; the second attempt uses another casing.
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			if email == "juan@example.com" {
				return &domain.User{ID: 1}, nil
			}
			return nil, nil
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", ptr("JUAN@Example.com"))
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("got %v, want ErrEmailAlreadyExists", err)
	}
}

func TestUserService_Register_ContactValidations(t *testing.T) {
	svc := newUserSvc(&mockUserRepo{})

	cases := []struct {
		name    string
		phone   string
		email   *string
		wantErr error
	}{
		{"nil email", "1122334455", nil, ErrEmailRequired},
		{"empty email", "1122334455", ptr(""), ErrEmailRequired},
		{"whitespace email", "1122334455", ptr("   "), ErrEmailRequired},
		{"email without domain", "1122334455", ptr("juan"), ErrInvalidEmail},
		{"email with display name", "1122334455", ptr("Juan <juan@example.com>"), ErrInvalidEmail},
		{"invalid phone", "abc", ptr("juan@example.com"), ErrInvalidPhone},
		{"too short phone", "123", ptr("juan@example.com"), ErrInvalidPhone},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Register(context.Background(), "Juan", tc.phone, "password1", tc.email)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserService_Register_EmailRepoError(t *testing.T) {
	dbErr := errors.New("db timeout")
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) { return nil, dbErr },
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", ptr("juan@example.com"))
	if !errors.Is(err, dbErr) {
		t.Errorf("got %v, want wrapped dbErr", err)
	}
}

func TestUserService_CreateStaff_EmailOptionalButNormalized(t *testing.T) {
	var stored *domain.User
	repo := &mockUserRepo{
		createFn: func(_ context.Context, u *domain.User) (int64, error) {
			stored = u
			return 3, nil
		},
	}
	svc := newUserSvc(repo)

	if _, err := svc.CreateStaff(context.Background(), "Ana", "1122334455", "password1", domain.RoleReceptionist, nil); err != nil {
		t.Fatalf("without email: unexpected error: %v", err)
	}
	if stored.Email != nil {
		t.Errorf("without email: got %v, want nil", *stored.Email)
	}

	if _, err := svc.CreateStaff(context.Background(), "Ana", "1122334455", "password1", domain.RoleReceptionist, ptr("   ")); err != nil {
		t.Fatalf("blank email: unexpected error: %v", err)
	}
	if stored.Email != nil {
		t.Errorf("blank email: got %v, want nil", *stored.Email)
	}

	if _, err := svc.CreateStaff(context.Background(), "Ana", "011 15 2233-4455", "password1", domain.RoleAdmin, ptr("ANA@Example.com")); err != nil {
		t.Fatalf("with email: unexpected error: %v", err)
	}
	if stored.Email == nil || *stored.Email != "ana@example.com" {
		t.Errorf("got email %v, want ana@example.com", stored.Email)
	}
	if stored.Phone != "+5491122334455" {
		t.Errorf("got phone %q, want +5491122334455", stored.Phone)
	}
}

func TestUserService_CreateStaff_ContactErrors(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			if email == "taken@example.com" {
				return &domain.User{ID: 1}, nil
			}
			return nil, nil
		},
	}
	svc := newUserSvc(repo)

	cases := []struct {
		name    string
		phone   string
		email   *string
		wantErr error
	}{
		{"invalid email", "1122334455", ptr("nope"), ErrInvalidEmail},
		{"invalid phone", "abc", nil, ErrInvalidPhone},
		{"email taken", "1122334455", ptr("Taken@Example.com"), ErrEmailAlreadyExists},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateStaff(context.Background(), "Ana", tc.phone, "password1", domain.RoleReceptionist, tc.email)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserService_Login_NormalizesPhone(t *testing.T) {
	const password = "secret123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	formats := []string{"3834123456", "0383 15-412-3456", "+54 9 383 412 3456"}
	for _, input := range formats {
		t.Run(input, func(t *testing.T) {
			var lookedUp string
			repo := &mockUserRepo{
				findByPhoneFn: func(_ context.Context, phone string) (*domain.User, error) {
					lookedUp = phone
					return &domain.User{ID: 7, IsActive: true, PasswordHash: string(hash)}, nil
				},
			}
			svc := newUserSvc(repo)

			if _, _, err := svc.Login(context.Background(), input, password); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if lookedUp != "+5493834123456" {
				t.Errorf("FindByPhone got %q, want +5493834123456", lookedUp)
			}
		})
	}
}

func TestUserService_Login_GarbagePhone(t *testing.T) {
	repo := &mockUserRepo{
		findByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) {
			t.Error("FindByPhone must not be called for an invalid phone")
			return nil, nil
		},
	}
	svc := newUserSvc(repo)

	_, _, err := svc.Login(context.Background(), "abc", "password1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestUserService_Register_MapsRepoDuplicateErrors(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		want    error
	}{
		{"duplicate email from unique index", domain.ErrDuplicateEmail, ErrEmailAlreadyExists},
		{"duplicate phone from unique index", domain.ErrDuplicatePhone, ErrPhoneAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserSvc(&mockUserRepo{
				createFn: func(_ context.Context, _ *domain.User) (int64, error) { return 0, tt.repoErr },
			})

			_, _, err := svc.Register(context.Background(), "Juan", "1122334455", "password1", ptr("juan@example.com"))
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestUserService_CreateStaff_MapsRepoDuplicateErrors(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		want    error
	}{
		{"duplicate email from unique index", domain.ErrDuplicateEmail, ErrEmailAlreadyExists},
		{"duplicate phone from unique index", domain.ErrDuplicatePhone, ErrPhoneAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserSvc(&mockUserRepo{
				createFn: func(_ context.Context, _ *domain.User) (int64, error) { return 0, tt.repoErr },
			})

			_, err := svc.CreateStaff(context.Background(), "Staff", "1122334455", "password1", domain.RoleReceptionist, ptr("staff@example.com"))
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
}
