// service/user_service.go
package service

import (
	"H1-canchas/internal/domain"
	"context"
	"fmt"
	"strings"

	jwtutil "H1-canchas/pkg"

	"golang.org/x/crypto/bcrypt"
)

// UserRepo es lo que el service necesita del repository.
// Lo define el service, lo implementa el repository.
type UserRepo interface {
	Create(ctx context.Context, user *domain.User) (int64, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	GetAll(ctx context.Context) ([]domain.User, error)
	UpdateRole(ctx context.Context, id int64, role string) error
	HasBookings(ctx context.Context, id int64) (bool, error)
	Delete(ctx context.Context, id int64) error
}

type UserService struct {
	repo      UserRepo
	jwtSecret string
}

func NewUserService(repo UserRepo, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register crea un usuario nuevo.
// Valida que el teléfono no esté en uso y hashea la password.
func (s *UserService) Register(ctx context.Context, name, phone, password string, email *string) (*domain.User, string, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)

	if email != nil {
		cleanEmail := strings.TrimSpace(*email)
		if cleanEmail == "" {
			email = nil
		} else {
			email = &cleanEmail
		}
	}

	if name == "" {
		return nil, "", ErrNameRequired
	}

	if phone == "" {
		return nil, "", ErrPhoneRequired
	}

	if password == "" {
		return nil, "", ErrPasswordRequired
	}
	if len(password) < 6 {
		return nil, "", ErrPasswordTooShort
	}

	existing, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, "", fmt.Errorf("Register: %w", err)
	}

	if existing != nil {
		return nil, "", ErrPhoneAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("Register: hasheando password: %w", err)
	}

	user := &domain.User{
		Name:         name,
		Phone:        phone,
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleCustomer,
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, "", fmt.Errorf("Register: %w", err)
	}

	user.ID = id

	token, err := jwtutil.GenerateToken(id, user.Role, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("Register: generando token: %w", err)
	}

	return user, token, nil
}

// CreateStaff crea un usuario con rol receptionist o admin.
// Solo puede llamarlo un admin. Devuelve el ID del usuario creado.
func (s *UserService) CreateStaff(ctx context.Context, name, phone, password, role string, email *string) (int64, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)

	if email != nil {
		cleanEmail := strings.TrimSpace(*email)
		if cleanEmail == "" {
			email = nil
		} else {
			email = &cleanEmail
		}
	}

	if role != domain.RoleReceptionist && role != domain.RoleAdmin {
		return 0, ErrInvalidRole
	}

	if name == "" {
		return 0, ErrNameRequired
	}
	if phone == "" {
		return 0, ErrPhoneRequired
	}
	if password == "" {
		return 0, ErrPasswordRequired
	}
	if len(password) < 6 {
		return 0, ErrPasswordTooShort
	}

	existing, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return 0, fmt.Errorf("CreateStaff: %w", err)
	}
	if existing != nil {
		return 0, ErrPhoneAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, fmt.Errorf("CreateStaff: hasheando password: %w", err)
	}

	user := &domain.User{
		Name:         name,
		Phone:        phone,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("CreateStaff: %w", err)
	}

	return id, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) UpdateRole(ctx context.Context, id int64, newRole string) error {
	if newRole != domain.RoleCustomer && newRole != domain.RoleReceptionist && newRole != domain.RoleAdmin {
		return ErrInvalidRole
	}
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("UpdateRole: %w", err)
	}
	if user == nil {
		return ErrNotFound
	}
	return s.repo.UpdateRole(ctx, id, newRole)
}

func (s *UserService) DeleteUser(ctx context.Context, targetID, requesterID int64) error {
	if targetID == requesterID {
		return ErrCannotDeleteSelf
	}

	user, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	if user == nil {
		return ErrNotFound
	}

	hasBookings, err := s.repo.HasBookings(ctx, targetID)
	if err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	if hasBookings {
		return ErrUserHasBookings
	}

	return s.repo.Delete(ctx, targetID)
}

// Login valida las credenciales y devuelve el usuario y un JWT si son correctas.
func (s *UserService) Login(ctx context.Context, phone, password string) (*domain.User, string, error) {
	phone = strings.TrimSpace(phone)
	password = strings.TrimSpace(password)

	if phone == "" || password == "" {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, "", fmt.Errorf("Login: %w", err)
	}

	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, "", ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("Login: generando token: %w", err)
	}

	return user, token, nil
}
