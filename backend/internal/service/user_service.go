// service/user_service.go
package service

import (
	"H1-canchas/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"

	jwtutil "H1-canchas/pkg"
	"H1-canchas/pkg/phone"

	"golang.org/x/crypto/bcrypt"
)

// UserRepo es lo que el service necesita del repository.
// Lo define el service, lo implementa el repository.
type UserRepo interface {
	Create(ctx context.Context, user *domain.User) (int64, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
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

// normalizeOptionalEmail applies normalizeEmail to an optional pointer input.
func normalizeOptionalEmail(email *string) (*string, error) {
	if email == nil {
		return nil, nil
	}
	return normalizeEmail(*email)
}

// checkContactAvailable normalizes rawPhone (E.164) and verifies that neither
// the phone nor the email (when not nil) belong to another user. It returns
// the normalized phone. Domain errors are returned unwrapped so controllers
// can show them as-is; infrastructure errors are wrapped with op.
func (s *UserService) checkContactAvailable(ctx context.Context, op, rawPhone string, email *string) (string, error) {
	normalizedPhone, err := phone.Normalize(rawPhone)
	if err != nil {
		return "", ErrInvalidPhone
	}

	byPhone, err := s.repo.FindByPhone(ctx, normalizedPhone)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if byPhone != nil {
		return "", ErrPhoneAlreadyExists
	}

	if email != nil {
		byEmail, err := s.repo.FindByEmail(ctx, *email)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
		if byEmail != nil {
			return "", ErrEmailAlreadyExists
		}
	}

	return normalizedPhone, nil
}

// Register crea un usuario nuevo.
// Exige email, normaliza teléfono y email, valida que no estén en uso y
// hashea la password.
func (s *UserService) Register(ctx context.Context, name, phone, password string, email *string) (*domain.User, string, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)

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

	cleanEmail, err := normalizeOptionalEmail(email)
	if err != nil {
		return nil, "", err
	}
	if cleanEmail == nil {
		return nil, "", ErrEmailRequired
	}

	phone, err = s.checkContactAvailable(ctx, "Register", phone, cleanEmail)
	if err != nil {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("Register: hasheando password: %w", err)
	}

	user := &domain.User{
		Name:         name,
		Phone:        phone,
		Email:        cleanEmail,
		PasswordHash: string(hash),
		Role:         domain.RoleCustomer,
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			return nil, "", ErrEmailAlreadyExists
		}
		if errors.Is(err, domain.ErrDuplicatePhone) {
			return nil, "", ErrPhoneAlreadyExists
		}
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

	cleanEmail, err := normalizeOptionalEmail(email)
	if err != nil {
		return 0, err
	}

	phone, err = s.checkContactAvailable(ctx, "CreateStaff", phone, cleanEmail)
	if err != nil {
		return 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, fmt.Errorf("CreateStaff: hasheando password: %w", err)
	}

	user := &domain.User{
		Name:         name,
		Phone:        phone,
		Email:        cleanEmail,
		PasswordHash: string(hash),
		Role:         role,
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			return 0, ErrEmailAlreadyExists
		}
		if errors.Is(err, domain.ErrDuplicatePhone) {
			return 0, ErrPhoneAlreadyExists
		}
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
func (s *UserService) Login(ctx context.Context, rawPhone, password string) (*domain.User, string, error) {
	rawPhone = strings.TrimSpace(rawPhone)
	password = strings.TrimSpace(password)

	if rawPhone == "" || password == "" {
		return nil, "", ErrInvalidCredentials
	}

	// The user may type the number in any format; a value that is not a
	// phone at all gets the same answer as a wrong password (no leak).
	normalizedPhone, err := phone.Normalize(rawPhone)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.repo.FindByPhone(ctx, normalizedPhone)
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
