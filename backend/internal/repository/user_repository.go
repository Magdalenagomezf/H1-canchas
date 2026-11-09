package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"H1-canchas/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

// userColumns is the column list shared by every user SELECT.
const userColumns = `id, name, email, phone, password_hash, google_sub, role, is_active, created_at`

// mapUniqueViolation turns a unique-index violation (pq 23505) into the
// matching domain sentinel. It returns nil for any other error.
func mapUniqueViolation(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		switch pqErr.Constraint {
		case "users_email_lower_key":
			return domain.ErrDuplicateEmail
		case "users_phone_key":
			return domain.ErrDuplicatePhone
		case "users_google_sub_key":
			return domain.ErrDuplicateGoogleSub
		}
	}
	return nil
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) (int64, error) {
	query := `
        INSERT INTO users (name, email, phone, password_hash, google_sub, role)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.GoogleSub,
		user.Role,
	).Scan(&id)

	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return 0, mapped
		}
		return 0, fmt.Errorf("userRepo.Create: %w", err)
	}
	return id, nil
}

func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
        SELECT `+userColumns+`
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

// FindByEmail busca por lower(email), igual que el índice único
// users_email_lower_key, así que matchea sin importar mayúsculas.
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT `+userColumns+`
        FROM users
        WHERE lower(email) = lower($1)`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.FindByEmail: %w", err)
	}
	return &user, nil
}

func (r *userRepo) FindByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	query := `
        SELECT ` + userColumns + `
        FROM users
        WHERE google_sub = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, sub)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.FindByGoogleSub: %w", err)
	}
	return &user, nil
}

// SetGoogleSub links an existing account to a Google identity.
func (r *userRepo) SetGoogleSub(ctx context.Context, id int64, sub string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET google_sub = $1 WHERE id = $2`, sub, id)
	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return mapped
		}
		return fmt.Errorf("userRepo.SetGoogleSub: %w", err)
	}
	return nil
}

// UpdateProfile overwrites name, email and phone. The caller passes the
// final values (merging absent fields beforehand); nil email/phone store NULL.
func (r *userRepo) UpdateProfile(ctx context.Context, id int64, name string, email, phone *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name = $1, email = $2, phone = $3 WHERE id = $4`,
		name, email, phone, id)
	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return mapped
		}
		return fmt.Errorf("userRepo.UpdateProfile: %w", err)
	}
	return nil
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
        SELECT `+userColumns+`
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
        SELECT `+userColumns+`
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
