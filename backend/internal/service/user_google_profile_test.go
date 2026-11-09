package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"H1-canchas/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// fakeGoogleVerifier implements GoogleTokenVerifier; Google is never called.
type fakeGoogleVerifier struct {
	identity *GoogleIdentity
	err      error
	got      string
}

func (f *fakeGoogleVerifier) Verify(_ context.Context, credential string) (*GoogleIdentity, error) {
	f.got = credential
	return f.identity, f.err
}

func newGoogleSvc(repo *mockUserRepo, v GoogleTokenVerifier) *UserService {
	return NewUserService(repo, "test-secret", v)
}

func okIdentity() *GoogleIdentity {
	return &GoogleIdentity{Sub: "sub-1", Email: "Ana@Example.com", EmailVerified: true, Name: "Ana Pérez"}
}

// --- LoginWithGoogle ---

func TestUserService_LoginWithGoogle_NotConfigured(t *testing.T) {
	svc := newGoogleSvc(&mockUserRepo{}, nil)

	_, _, err := svc.LoginWithGoogle(context.Background(), "tok")
	if !errors.Is(err, ErrGoogleNotConfigured) {
		t.Errorf("got %v, want ErrGoogleNotConfigured", err)
	}
}

func TestUserService_LoginWithGoogle_InvalidToken(t *testing.T) {
	cases := []struct {
		name string
		v    *fakeGoogleVerifier
	}{
		{"verifier error", &fakeGoogleVerifier{err: errors.New("bad signature")}},
		{"empty email", &fakeGoogleVerifier{identity: &GoogleIdentity{Sub: "s", Email: "", EmailVerified: true}}},
		{"malformed email", &fakeGoogleVerifier{identity: &GoogleIdentity{Sub: "s", Email: "nope", EmailVerified: true}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findByGoogleSubFn: func(context.Context, string) (*domain.User, error) {
					t.Error("repo must not be touched for an invalid token")
					return nil, nil
				},
			}
			_, _, err := newGoogleSvc(repo, tc.v).LoginWithGoogle(context.Background(), "tok")
			if !errors.Is(err, ErrInvalidGoogleToken) {
				t.Errorf("got %v, want ErrInvalidGoogleToken", err)
			}
		})
	}
}

func TestUserService_LoginWithGoogle_InvalidTokenKeepsCauseForLogs(t *testing.T) {
	v := &fakeGoogleVerifier{err: errors.New("audience mismatch")}

	_, _, err := newGoogleSvc(&mockUserRepo{}, v).LoginWithGoogle(context.Background(), "tok")
	if !errors.Is(err, ErrInvalidGoogleToken) {
		t.Fatalf("got %v, want ErrInvalidGoogleToken", err)
	}
	if !strings.Contains(err.Error(), "audience mismatch") {
		t.Errorf("error %q must keep the verifier cause so the controller can log it", err)
	}
}

func TestUserService_LoginWithGoogle_EmailNotVerified(t *testing.T) {
	v := &fakeGoogleVerifier{identity: &GoogleIdentity{Sub: "s", Email: "a@b.com", EmailVerified: false}}
	repo := &mockUserRepo{
		findByGoogleSubFn: func(context.Context, string) (*domain.User, error) {
			t.Error("repo must not be touched when the email is not verified")
			return nil, nil
		},
	}

	_, _, err := newGoogleSvc(repo, v).LoginWithGoogle(context.Background(), "tok")
	if !errors.Is(err, ErrGoogleEmailNotVerified) {
		t.Errorf("got %v, want ErrGoogleEmailNotVerified", err)
	}
}

func TestUserService_LoginWithGoogle_CreatesCustomerWithoutPhoneOrPassword(t *testing.T) {
	var created *domain.User
	repo := &mockUserRepo{
		createFn: func(_ context.Context, u *domain.User) (int64, error) {
			created = u
			return 21, nil
		},
	}
	v := &fakeGoogleVerifier{identity: okIdentity()}

	user, token, err := newGoogleSvc(repo, v).LoginWithGoogle(context.Background(), "the-credential")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.got != "the-credential" {
		t.Errorf("verifier got %q", v.got)
	}
	if token == "" || user.ID != 21 {
		t.Errorf("token=%q id=%d, want token and id 21", token, user.ID)
	}
	if created.Role != domain.RoleCustomer {
		t.Errorf("role = %q, want customer", created.Role)
	}
	if created.Phone != nil || created.PasswordHash != nil {
		t.Errorf("phone/password must be nil, got %v / %v", created.Phone, created.PasswordHash)
	}
	if created.GoogleSub == nil || *created.GoogleSub != "sub-1" {
		t.Errorf("google sub = %v, want sub-1", created.GoogleSub)
	}
	if created.Email == nil || *created.Email != "ana@example.com" {
		t.Errorf("email = %v, want lowercase ana@example.com", created.Email)
	}
	if created.Name != "Ana Pérez" {
		t.Errorf("name = %q", created.Name)
	}
}

func TestUserService_LoginWithGoogle_NameFallbackAndTruncation(t *testing.T) {
	cases := []struct {
		name       string
		googleName string
		wantName   string
	}{
		{"empty name uses email local part", "  ", "ana"},
		{"long name is cut to 100 characters", strings.Repeat("ñ", 150), strings.Repeat("ñ", 100)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var created *domain.User
			repo := &mockUserRepo{
				createFn: func(_ context.Context, u *domain.User) (int64, error) {
					created = u
					return 1, nil
				},
			}
			id := okIdentity()
			id.Name = tc.googleName
			if _, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: id}).LoginWithGoogle(context.Background(), "t"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if created.Name != tc.wantName {
				t.Errorf("name = %q, want %q", created.Name, tc.wantName)
			}
		})
	}
}

func TestUserService_LoginWithGoogle_SameSubTwiceLogsInWithoutCreating(t *testing.T) {
	existing := &domain.User{ID: 8, Role: domain.RoleCustomer, IsActive: true, GoogleSub: ptr("sub-1")}
	repo := &mockUserRepo{
		findByGoogleSubFn: func(_ context.Context, sub string) (*domain.User, error) {
			if sub == "sub-1" {
				return existing, nil
			}
			return nil, nil
		},
		createFn: func(context.Context, *domain.User) (int64, error) {
			t.Error("Create must not be called for a known sub")
			return 0, nil
		},
		findByEmailFn: func(context.Context, string) (*domain.User, error) {
			t.Error("FindByEmail must not be called for a known sub")
			return nil, nil
		},
	}
	svc := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()})

	for i := 0; i < 2; i++ {
		user, token, err := svc.LoginWithGoogle(context.Background(), "t")
		if err != nil {
			t.Fatalf("attempt %d: unexpected error: %v", i, err)
		}
		if user.ID != 8 || token == "" {
			t.Errorf("attempt %d: id=%d token=%q", i, user.ID, token)
		}
	}
}

func TestUserService_LoginWithGoogle_LinksExistingCustomer(t *testing.T) {
	var linkedID int64
	var linkedSub string
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			if email != "ana@example.com" {
				t.Errorf("FindByEmail got %q, want normalized email", email)
			}
			return &domain.User{ID: 5, Role: domain.RoleCustomer, IsActive: true}, nil
		},
		setGoogleSubFn: func(_ context.Context, id int64, sub string) error {
			linkedID, linkedSub = id, sub
			return nil
		},
		createFn: func(context.Context, *domain.User) (int64, error) {
			t.Error("Create must not be called when linking")
			return 0, nil
		},
	}

	user, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if linkedID != 5 || linkedSub != "sub-1" {
		t.Errorf("linked id=%d sub=%q, want 5 / sub-1", linkedID, linkedSub)
	}
	if user.ID != 5 {
		t.Errorf("returned user %d, want 5", user.ID)
	}
}

func TestUserService_LoginWithGoogle_StaffAccountIsNotLinked(t *testing.T) {
	for _, role := range []string{domain.RoleAdmin, domain.RoleReceptionist} {
		t.Run(role, func(t *testing.T) {
			repo := &mockUserRepo{
				findByEmailFn: func(context.Context, string) (*domain.User, error) {
					return &domain.User{ID: 2, Role: role, IsActive: true}, nil
				},
				setGoogleSubFn: func(context.Context, int64, string) error {
					t.Error("staff accounts must never be linked")
					return nil
				},
			}
			_, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
			if !errors.Is(err, ErrGoogleStaffAccount) {
				t.Errorf("got %v, want ErrGoogleStaffAccount", err)
			}
		})
	}
}

func TestUserService_LoginWithGoogle_CustomerLinkedToAnotherSub(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFn: func(context.Context, string) (*domain.User, error) {
			return &domain.User{ID: 5, Role: domain.RoleCustomer, IsActive: true, GoogleSub: ptr("other-sub")}, nil
		},
		setGoogleSubFn: func(context.Context, int64, string) error {
			t.Error("an existing google_sub must never be overwritten")
			return nil
		},
	}
	_, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
	if !errors.Is(err, ErrGoogleAccountConflict) {
		t.Errorf("got %v, want ErrGoogleAccountConflict", err)
	}
}

func TestUserService_LoginWithGoogle_InactiveUser(t *testing.T) {
	t.Run("known sub", func(t *testing.T) {
		repo := &mockUserRepo{
			findByGoogleSubFn: func(context.Context, string) (*domain.User, error) {
				return &domain.User{ID: 8, Role: domain.RoleCustomer, IsActive: false}, nil
			},
		}
		_, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
		if !errors.Is(err, ErrUserInactive) {
			t.Errorf("got %v, want ErrUserInactive", err)
		}
	})
	t.Run("linked by email", func(t *testing.T) {
		repo := &mockUserRepo{
			findByEmailFn: func(context.Context, string) (*domain.User, error) {
				return &domain.User{ID: 8, Role: domain.RoleCustomer, IsActive: false}, nil
			},
		}
		_, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
		if !errors.Is(err, ErrUserInactive) {
			t.Errorf("got %v, want ErrUserInactive", err)
		}
	})
}

func TestUserService_LoginWithGoogle_ConcurrentInsertIsEmailConflict(t *testing.T) {
	for _, repoErr := range []error{domain.ErrDuplicateEmail, domain.ErrDuplicateGoogleSub} {
		t.Run(repoErr.Error(), func(t *testing.T) {
			calls := 0
			repo := &mockUserRepo{
				createFn: func(context.Context, *domain.User) (int64, error) {
					calls++
					return 0, repoErr
				},
			}
			_, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
			if !errors.Is(err, ErrEmailAlreadyExists) {
				t.Errorf("got %v, want ErrEmailAlreadyExists", err)
			}
			if calls != 1 {
				t.Errorf("Create called %d times, want 1 (no retry)", calls)
			}
		})
	}
}

func TestUserService_LoginWithGoogle_RepoError(t *testing.T) {
	dbErr := errors.New("db down")
	repo := &mockUserRepo{
		findByGoogleSubFn: func(context.Context, string) (*domain.User, error) { return nil, dbErr },
	}
	_, _, err := newGoogleSvc(repo, &fakeGoogleVerifier{identity: okIdentity()}).LoginWithGoogle(context.Background(), "t")
	if !errors.Is(err, dbErr) {
		t.Errorf("got %v, want wrapped dbErr", err)
	}
}

// --- Login with phone on an account without password ---

func TestUserService_Login_AccountWithoutPassword(t *testing.T) {
	repo := &mockUserRepo{
		findByPhoneFn: func(context.Context, string) (*domain.User, error) {
			return &domain.User{ID: 3, IsActive: true, Phone: ptr("+5493834123456"), PasswordHash: nil, GoogleSub: ptr("s")}, nil
		},
	}
	_, _, err := newUserSvc(repo).Login(context.Background(), "3834123456", "whatever1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

// --- GetProfile / UpdateProfile ---

func profileUser() *domain.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	return &domain.User{
		ID: 4, Name: "Ana", Email: ptr("ana@example.com"), Phone: ptr("+5493834123456"),
		PasswordHash: ptr(string(hash)), Role: domain.RoleCustomer, IsActive: true,
	}
}

type profileCall struct {
	called       bool
	name         string
	email, phone *string
}

func profileRepo(user *domain.User, call *profileCall) *mockUserRepo {
	return &mockUserRepo{
		findByIDFn: func(context.Context, int64) (*domain.User, error) { return user, nil },
		updateProfileFn: func(_ context.Context, _ int64, name string, email, phone *string) error {
			*call = profileCall{true, name, email, phone}
			return nil
		},
	}
}

func TestUserService_GetProfile(t *testing.T) {
	svc := newUserSvc(&mockUserRepo{findByIDFn: func(context.Context, int64) (*domain.User, error) { return profileUser(), nil }})
	got, err := svc.GetProfile(context.Background(), 4)
	if err != nil || got.ID != 4 {
		t.Fatalf("got %+v, %v", got, err)
	}

	svc = newUserSvc(&mockUserRepo{})
	if _, err := svc.GetProfile(context.Background(), 4); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestUserService_UpdateProfile_AbsentFieldsChangeNothing(t *testing.T) {
	var call profileCall
	svc := newUserSvc(profileRepo(profileUser(), &call))

	got, err := svc.UpdateProfile(context.Background(), 4, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if call.name != "Ana" || *call.email != "ana@example.com" || *call.phone != "+5493834123456" {
		t.Errorf("stored values changed: %+v", call)
	}
	if got.Name != "Ana" || *got.Email != "ana@example.com" || *got.Phone != "+5493834123456" {
		t.Errorf("returned user changed: %+v", got)
	}
}

func TestUserService_UpdateProfile_OnlyPhoneKeepsRest(t *testing.T) {
	user := profileUser()
	user.Phone = nil // Google account completing its profile
	var call profileCall
	svc := newUserSvc(profileRepo(user, &call))

	got, err := svc.UpdateProfile(context.Background(), 4, nil, nil, ptr("0383 15-412-3456"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if call.phone == nil || *call.phone != "+5493834123456" {
		t.Errorf("stored phone = %v, want +5493834123456", call.phone)
	}
	if call.name != "Ana" || *call.email != "ana@example.com" {
		t.Errorf("name/email must be untouched: %+v", call)
	}
	if got.Phone == nil || *got.Phone != "+5493834123456" {
		t.Errorf("returned phone = %v", got.Phone)
	}
}

func TestUserService_UpdateProfile_NormalizesEmailAndTrimsName(t *testing.T) {
	var call profileCall
	svc := newUserSvc(profileRepo(profileUser(), &call))

	if _, err := svc.UpdateProfile(context.Background(), 4, ptr("  Ana María  "), ptr("  ANA.M@Example.COM "), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if call.name != "Ana María" {
		t.Errorf("name = %q", call.name)
	}
	if call.email == nil || *call.email != "ana.m@example.com" {
		t.Errorf("email = %v, want lowercase", call.email)
	}
}

func TestUserService_UpdateProfile_Validations(t *testing.T) {
	cases := []struct {
		name         string
		n, e, p      *string
		wantErr      error
	}{
		{"empty name", ptr("  "), nil, nil, ErrNameRequired},
		{"name too long", ptr(strings.Repeat("a", 101)), nil, nil, ErrNameTooLong},
		{"empty email", nil, ptr(""), nil, ErrEmailRequired},
		{"blank email", nil, ptr("   "), nil, ErrEmailRequired},
		{"invalid email", nil, ptr("nope"), nil, ErrInvalidEmail},
		{"empty phone", nil, nil, ptr(""), ErrInvalidPhone},
		{"invalid phone", nil, nil, ptr("abc"), ErrInvalidPhone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var call profileCall
			svc := newUserSvc(profileRepo(profileUser(), &call))

			_, err := svc.UpdateProfile(context.Background(), 4, tc.n, tc.e, tc.p)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
			if call.called {
				t.Error("nothing must be written when validation fails")
			}
		})
	}
}

func TestUserService_UpdateProfile_NameOf100CharactersIsAllowed(t *testing.T) {
	var call profileCall
	svc := newUserSvc(profileRepo(profileUser(), &call))

	if _, err := svc.UpdateProfile(context.Background(), 4, ptr(strings.Repeat("ñ", 100)), nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_UpdateProfile_ContactUsedByAnotherUser(t *testing.T) {
	other := &domain.User{ID: 99}
	t.Run("email", func(t *testing.T) {
		var call profileCall
		repo := profileRepo(profileUser(), &call)
		repo.findByEmailFn = func(context.Context, string) (*domain.User, error) { return other, nil }

		_, err := newUserSvc(repo).UpdateProfile(context.Background(), 4, nil, ptr("taken@example.com"), nil)
		if !errors.Is(err, ErrEmailAlreadyExists) {
			t.Errorf("got %v, want ErrEmailAlreadyExists", err)
		}
		if call.called {
			t.Error("must not write")
		}
	})
	t.Run("phone", func(t *testing.T) {
		var call profileCall
		repo := profileRepo(profileUser(), &call)
		repo.findByPhoneFn = func(context.Context, string) (*domain.User, error) { return other, nil }

		_, err := newUserSvc(repo).UpdateProfile(context.Background(), 4, nil, nil, ptr("3834000000"))
		if !errors.Is(err, ErrPhoneAlreadyExists) {
			t.Errorf("got %v, want ErrPhoneAlreadyExists", err)
		}
		if call.called {
			t.Error("must not write")
		}
	})
}

func TestUserService_UpdateProfile_OwnContactIsNotAConflict(t *testing.T) {
	var call profileCall
	repo := profileRepo(profileUser(), &call)
	repo.findByEmailFn = func(context.Context, string) (*domain.User, error) { return &domain.User{ID: 4}, nil }
	repo.findByPhoneFn = func(context.Context, string) (*domain.User, error) { return &domain.User{ID: 4}, nil }

	if _, err := newUserSvc(repo).UpdateProfile(context.Background(), 4, nil, ptr("ANA@example.com"), ptr("3834123456")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_UpdateProfile_MapsUniqueViolationOnUpdate(t *testing.T) {
	tests := []struct {
		repoErr error
		want    error
	}{
		{domain.ErrDuplicateEmail, ErrEmailAlreadyExists},
		{domain.ErrDuplicatePhone, ErrPhoneAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.repoErr.Error(), func(t *testing.T) {
			repo := &mockUserRepo{
				findByIDFn:      func(context.Context, int64) (*domain.User, error) { return profileUser(), nil },
				updateProfileFn: func(context.Context, int64, string, *string, *string) error { return tt.repoErr },
			}
			_, err := newUserSvc(repo).UpdateProfile(context.Background(), 4, ptr("Ana"), nil, nil)
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	_, err := newUserSvc(&mockUserRepo{}).UpdateProfile(context.Background(), 4, ptr("Ana"), nil, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestUserService_UpdateProfile_GoogleAccountMayChangeEmail(t *testing.T) {
	user := profileUser()
	user.GoogleSub = ptr("sub-1")
	user.PasswordHash = nil
	var call profileCall
	svc := newUserSvc(profileRepo(user, &call))

	if _, err := svc.UpdateProfile(context.Background(), 4, nil, ptr("new@example.com"), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *call.email != "new@example.com" {
		t.Errorf("email = %v", call.email)
	}
}
