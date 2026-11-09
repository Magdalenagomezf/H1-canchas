package repository_test

import (
	"context"
	"errors"
	"testing"

	"H1-canchas/internal/domain"
	"H1-canchas/internal/repository"
)

func TestUserRepo_Create_ReturnsID(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	user := &domain.User{
		Name:         "Ana García",
		Phone:        strPtr("3511234567"),
		PasswordHash: strPtr("somehash"),
		Role:         domain.RoleCustomer,
	}
	id, err := repo.Create(context.Background(), user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestUserRepo_FindByPhone_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	got, err := repo.FindByPhone(context.Background(), "0000000000")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestUserRepo_FindByPhone_Found(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	user := &domain.User{Name: "Carlos", Phone: strPtr("3519876543"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer}
	id, _ := repo.Create(context.Background(), user)

	got, err := repo.FindByPhone(context.Background(), "3519876543")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.ID != id {
		t.Errorf("got ID %d, want %d", got.ID, id)
	}
	if got.Phone == nil || *got.Phone != "3519876543" {
		t.Errorf("got phone %v, want %q", got.Phone, "3519876543")
	}
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	got, err := repo.FindByID(context.Background(), 99999)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestUserRepo_FindByID_Found(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	user := &domain.User{Name: "María", Phone: strPtr("3510001111"), PasswordHash: strPtr("hash"), Role: domain.RoleReceptionist}
	id, _ := repo.Create(context.Background(), user)

	got, err := repo.FindByID(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.Name != "María" {
		t.Errorf("got name %q, want %q", got.Name, "María")
	}
	if got.Role != domain.RoleReceptionist {
		t.Errorf("got role %q, want %q", got.Role, domain.RoleReceptionist)
	}
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	got, err := repo.FindByEmail(context.Background(), "nobody@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestUserRepo_FindByEmail_IsCaseInsensitive(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	email := "ana@example.com"
	id, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana", Email: &email, Phone: strPtr("+5493834123456"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.FindByEmail(context.Background(), "ANA@Example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != id {
		t.Fatalf("expected user %d, got %+v", id, got)
	}
}

func TestUserRepo_Create_RejectsEmailDifferingOnlyByCase(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	lower, upper := "ana@example.com", "ANA@example.com"
	if _, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana", Email: &lower, Phone: strPtr("+5493834123456"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer,
	}); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana 2", Email: &upper, Phone: strPtr("+5493834123457"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer,
	})
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Fatalf("got %v, want domain.ErrDuplicateEmail", err)
	}
}

func TestUserRepo_Create_RejectsDuplicatePhone(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	if _, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana", Phone: strPtr("+5493834123456"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer,
	}); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana 2", Phone: strPtr("+5493834123456"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer,
	})
	if !errors.Is(err, domain.ErrDuplicatePhone) {
		t.Fatalf("got %v, want domain.ErrDuplicatePhone", err)
	}
}

func strPtr(s string) *string { return &s }

// --- Phase 4a: Google login and profile ---

func TestUserRepo_Create_GoogleUserWithNullPhoneAndPassword(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)
	ctx := context.Background()

	email := "g@example.com"
	id, err := repo.Create(ctx, &domain.User{
		Name: "Google User", Email: &email, GoogleSub: strPtr("sub-1"), Role: domain.RoleCustomer,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Every read path must scan NULL phone / password_hash without failing.
	byID, err := repo.FindByID(ctx, id)
	if err != nil || byID == nil {
		t.Fatalf("FindByID: %v, %+v", err, byID)
	}
	if byID.Phone != nil || byID.PasswordHash != nil {
		t.Errorf("expected NULL phone and password, got %v / %v", byID.Phone, byID.PasswordHash)
	}
	if byID.GoogleSub == nil || *byID.GoogleSub != "sub-1" {
		t.Errorf("google sub = %v", byID.GoogleSub)
	}
	if byID.Role != domain.RoleCustomer || !byID.IsActive {
		t.Errorf("unexpected role/active: %+v", byID)
	}

	if byEmail, err := repo.FindByEmail(ctx, "G@Example.com"); err != nil || byEmail == nil || byEmail.ID != id {
		t.Errorf("FindByEmail: %v, %+v", err, byEmail)
	}
	if bySub, err := repo.FindByGoogleSub(ctx, "sub-1"); err != nil || bySub == nil || bySub.ID != id {
		t.Errorf("FindByGoogleSub: %v, %+v", err, bySub)
	}
	all, err := repo.GetAll(ctx)
	if err != nil || len(all) != 1 {
		t.Errorf("GetAll: %v, %d users", err, len(all))
	}
}

func TestUserRepo_Create_AllowsSeveralUsersWithoutPhone(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	for i, sub := range []string{"sub-a", "sub-b"} {
		email := []string{"a@example.com", "b@example.com"}[i]
		if _, err := repo.Create(context.Background(), &domain.User{
			Name: "G", Email: &email, GoogleSub: strPtr(sub), Role: domain.RoleCustomer,
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}
}

func TestUserRepo_Create_RejectsAccountWithoutLoginMethod(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	_, err := repo.Create(context.Background(), &domain.User{
		Name: "Nobody", Phone: strPtr("+5493834123456"), Role: domain.RoleCustomer,
	})
	if err == nil {
		t.Fatal("expected the users_login_method_chk constraint to reject the row")
	}
}

func TestUserRepo_Create_RejectsDuplicateGoogleSub(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	e1, e2 := "one@example.com", "two@example.com"
	if _, err := repo.Create(context.Background(), &domain.User{
		Name: "One", Email: &e1, GoogleSub: strPtr("same"), Role: domain.RoleCustomer,
	}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := repo.Create(context.Background(), &domain.User{
		Name: "Two", Email: &e2, GoogleSub: strPtr("same"), Role: domain.RoleCustomer,
	})
	if !errors.Is(err, domain.ErrDuplicateGoogleSub) {
		t.Fatalf("got %v, want domain.ErrDuplicateGoogleSub", err)
	}
}

func TestUserRepo_FindByGoogleSub_NotFound(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	got, err := repo.FindByGoogleSub(context.Background(), "missing")
	if err != nil || got != nil {
		t.Errorf("got %+v, %v; want nil, nil", got, err)
	}
}

func TestUserRepo_SetGoogleSub(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)
	ctx := context.Background()

	id, err := repo.Create(ctx, &domain.User{Name: "Ana", Phone: strPtr("+5493834123456"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.SetGoogleSub(ctx, id, "sub-x"); err != nil {
		t.Fatalf("SetGoogleSub: %v", err)
	}
	got, _ := repo.FindByGoogleSub(ctx, "sub-x")
	if got == nil || got.ID != id {
		t.Fatalf("expected user %d, got %+v", id, got)
	}
	if got.PasswordHash == nil || got.Phone == nil {
		t.Error("linking must keep password and phone")
	}

	other, _ := repo.Create(ctx, &domain.User{Name: "Bea", Phone: strPtr("+5493834123457"), PasswordHash: strPtr("hash"), Role: domain.RoleCustomer})
	if err := repo.SetGoogleSub(ctx, other, "sub-x"); !errors.Is(err, domain.ErrDuplicateGoogleSub) {
		t.Errorf("got %v, want domain.ErrDuplicateGoogleSub", err)
	}
}

func TestUserRepo_UpdateProfile(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)
	ctx := context.Background()

	old := "old@example.com"
	id, err := repo.Create(ctx, &domain.User{Name: "Ana", Email: &old, GoogleSub: strPtr("s"), Role: domain.RoleCustomer})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newEmail, newPhone := "new@example.com", "+5493834123456"
	if err := repo.UpdateProfile(ctx, id, "Ana María", &newEmail, &newPhone); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	got, _ := repo.FindByID(ctx, id)
	if got.Name != "Ana María" || *got.Email != newEmail || *got.Phone != newPhone {
		t.Errorf("unexpected stored profile: %+v", got)
	}
	if got.Role != domain.RoleCustomer || got.GoogleSub == nil {
		t.Errorf("role / google sub must be untouched: %+v", got)
	}
}

func TestUserRepo_UpdateProfile_MapsUniqueViolations(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)
	ctx := context.Background()

	e1, e2 := "one@example.com", "two@example.com"
	p1, p2 := "+5493834123456", "+5493834123457"
	if _, err := repo.Create(ctx, &domain.User{Name: "One", Email: &e1, Phone: &p1, PasswordHash: strPtr("h"), Role: domain.RoleCustomer}); err != nil {
		t.Fatalf("create one: %v", err)
	}
	two, err := repo.Create(ctx, &domain.User{Name: "Two", Email: &e2, Phone: &p2, PasswordHash: strPtr("h"), Role: domain.RoleCustomer})
	if err != nil {
		t.Fatalf("create two: %v", err)
	}

	upper := "ONE@example.com"
	if err := repo.UpdateProfile(ctx, two, "Two", &upper, &p2); !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Errorf("email: got %v, want domain.ErrDuplicateEmail", err)
	}
	if err := repo.UpdateProfile(ctx, two, "Two", &e2, &p1); !errors.Is(err, domain.ErrDuplicatePhone) {
		t.Errorf("phone: got %v, want domain.ErrDuplicatePhone", err)
	}
}
