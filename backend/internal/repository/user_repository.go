package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"H1-canchas/internal/domain"

	"github.com/jmoiron/sqlx"
)

// userRepo es la implementación concreta de UserRepository.
// Es privada (minúscula) — solo se crea desde NewUserRepository.
type userRepo struct {
	db *sqlx.DB
}

// NewUserRepository crea una instancia del repository.
func NewUserRepository(db *sqlx.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) (int64, error) {
	query := `
        INSERT INTO users (name, email, phone, password_hash, role)
        VALUES ($1, $2, $3, $4, $5) 
        RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("userRepo.Create: %w", err)
	}
	return id, nil
}

func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
        SELECT id, name, email, phone, password_hash, role, is_active, created_at
        FROM users
        WHERE phone = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, phone)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // no existe, no es un error
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.FindByPhone: %w", err)
	}
	return &user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
        SELECT id, name, email, phone, password_hash, role, is_active, created_at
        FROM users
        WHERE id = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.FindByID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetAll(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := r.db.SelectContext(ctx, &users, `
        SELECT id, name, email, phone, password_hash, role, is_active, created_at
        FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetAll: %w", err)
	}
	return users, nil
}

func (r *userRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, id)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateRole: %w", err)
	}
	return nil
}

func (r *userRepo) HasBookings(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE customer_user_id = $1 OR created_by = $1 LIMIT 1)`, id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("userRepo.HasBookings: %w", err)
	}
	return exists, nil
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("userRepo.Delete: %w", err)
	}
	return nil
}
