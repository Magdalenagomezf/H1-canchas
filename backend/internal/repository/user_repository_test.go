package repository_test

import (
	"context"
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
