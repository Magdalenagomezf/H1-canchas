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
		Phone:        "3511234567",
		PasswordHash: "somehash",
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

	user := &domain.User{Name: "Carlos", Phone: "3519876543", PasswordHash: "hash", Role: domain.RoleCustomer}
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
	if got.Phone != "3519876543" {
		t.Errorf("got phone %q, want %q", got.Phone, "3519876543")
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

	user := &domain.User{Name: "María", Phone: "3510001111", PasswordHash: "hash", Role: domain.RoleReceptionist}
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
		Name: "Ana", Email: &email, Phone: "+5493834123456", PasswordHash: "hash", Role: domain.RoleCustomer,
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
		Name: "Ana", Email: &lower, Phone: "+5493834123456", PasswordHash: "hash", Role: domain.RoleCustomer,
	}); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana 2", Email: &upper, Phone: "+5493834123457", PasswordHash: "hash", Role: domain.RoleCustomer,
	})
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Fatalf("got %v, want domain.ErrDuplicateEmail", err)
	}
}

func TestUserRepo_Create_RejectsDuplicatePhone(t *testing.T) {
	truncateAll(t)
	repo := repository.NewUserRepository(testDB)

	if _, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana", Phone: "+5493834123456", PasswordHash: "hash", Role: domain.RoleCustomer,
	}); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(context.Background(), &domain.User{
		Name: "Ana 2", Phone: "+5493834123456", PasswordHash: "hash", Role: domain.RoleCustomer,
	})
	if !errors.Is(err, domain.ErrDuplicatePhone) {
		t.Fatalf("got %v, want domain.ErrDuplicatePhone", err)
	}
}
