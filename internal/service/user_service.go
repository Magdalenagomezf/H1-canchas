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
func (s *UserService) Register(ctx context.Context, name, phone, password string, email *string) (string, error) {
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
		return "", ErrNameRequired
	}

	if phone == "" {
		return "", ErrPhoneRequired
	}

	if password == "" {
		return "", ErrPasswordRequired
	}
	if len(password) < 6 {
		return "", ErrPasswordTooShort
	}

	existing, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("Register: %w", err)
	}

	if existing != nil {
		return "", ErrPhoneAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("Register: hasheando password: %w", err)
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
		return "", fmt.Errorf("Register: %w", err)
	}

	token, err := jwtutil.GenerateToken(id, user.Role, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("Register: generando token: %w", err)
	}

	return token, nil
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

// Login valida las credenciales y devuelve un JWT si son correctas.
func (s *UserService) Login(ctx context.Context, phone, password string) (string, error) {
	phone = strings.TrimSpace(phone)
	password = strings.TrimSpace(password)

	if phone == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("Login: %w", err)
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("Login: generando token: %w", err)
	}

	return token, nil
}
