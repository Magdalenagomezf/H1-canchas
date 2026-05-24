package service

import (
	"context"
	"fmt"
	"strings"

	"H1-canchas/internal/domain"
)

// SpaceRepo es la interfaz que el service necesita del repository.
// La define el service, la implementa el repository.
type SpaceRepo interface {
	Create(ctx context.Context, space *domain.Space) (int64, error)
	GetAll(ctx context.Context) ([]domain.Space, error)
	GetByID(ctx context.Context, id int64) (*domain.Space, error)
	GetSlotsBySpaceID(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error)
	Update(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error
	CreateSlot(ctx context.Context, slot *domain.SpaceSlot) (int64, error)
	Deactivate(ctx context.Context, id int64) error
}

type SpaceService struct {
	repo SpaceRepo
}

func NewSpaceService(repo SpaceRepo) *SpaceService {
	return &SpaceService{repo: repo}
}

func (s *SpaceService) Create(ctx context.Context, name, spaceType string, description *string, pricePerSlot float64) (int64, error) {
	name = strings.TrimSpace(name)
	spaceType = strings.TrimSpace(strings.ToLower(spaceType))

	if description != nil {
		cleanDescription := strings.TrimSpace(*description)
		if cleanDescription == "" {
			description = nil
		} else {
			description = &cleanDescription
		}
	}

	if name == "" {
		return 0, ErrSpaceNameRequired
	}

	if !isValidSpaceType(spaceType) {
		return 0, ErrInvalidSpaceType
	}

	if pricePerSlot <= 0 {
		return 0, ErrInvalidSpacePrice
	}

	space := &domain.Space{
		Name:         name,
		Type:         spaceType,
		Description:  description,
		PricePerSlot: pricePerSlot,
	}

	id, err := s.repo.Create(ctx, space)
	if err != nil {
		return 0, fmt.Errorf("SpaceService.Create: %w", err)
	}

	return id, nil
}

// GetAll devuelve todos los espacios activos.
func (s *SpaceService) GetAll(ctx context.Context) ([]domain.Space, error) {
	spaces, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("SpaceService.GetAll: %w", err)
	}
	return spaces, nil
}

// GetByID devuelve un espacio por ID.
// Devuelve ErrNotFound si no existe o está inactivo.
func (s *SpaceService) GetByID(ctx context.Context, id int64) (*domain.Space, error) {
	if id <= 0 {
		return nil, ErrInvalidSpaceID
	}
	space, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("SpaceService.GetByID: %w", err)
	}
	if space == nil {
		return nil, ErrNotFound
	}
	return space, nil
}

// GetSlots devuelve los turnos activos de un espacio.
// Valida primero que el espacio exista.
func (s *SpaceService) GetSlots(ctx context.Context, spaceID int64) ([]domain.SpaceSlot, error) {
	if spaceID <= 0 {
		return nil, ErrInvalidSpaceID
	}

	space, err := s.repo.GetByID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("SpaceService.GetSlots: %w", err)
	}
	if space == nil {
		return nil, ErrNotFound
	}

	slots, err := s.repo.GetSlotsBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("SpaceService.GetSlots: %w", err)
	}
	return slots, nil
}

func (s *SpaceService) CreateSlot(ctx context.Context, spaceID int64, label string, description, startTime, endTime *string) (int64, error) {
	if spaceID <= 0 {
		return 0, ErrInvalidSpaceID
	}

	label = strings.TrimSpace(label)
	if label == "" {
		return 0, ErrSlotLabelRequired
	}

	space, err := s.repo.GetByID(ctx, spaceID)
	if err != nil {
		return 0, fmt.Errorf("SpaceService.CreateSlot: %w", err)
	}
	if space == nil || !space.IsActive {
		return 0, ErrNotFound
	}

	slot := &domain.SpaceSlot{
		SpaceID:     spaceID,
		Label:       label,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	id, err := s.repo.CreateSlot(ctx, slot)
	if err != nil {
		return 0, fmt.Errorf("SpaceService.CreateSlot: %w", err)
	}

	return id, nil
}

func (s *SpaceService) Update(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error {
	if id <= 0 {
		return ErrInvalidSpaceID
	}

	name = strings.TrimSpace(name)
	if description != nil {
		clean := strings.TrimSpace(*description)
		if clean == "" {
			description = nil
		} else {
			description = &clean
		}
	}

	if name == "" {
		return ErrSpaceNameRequired
	}
	if pricePerSlot <= 0 {
		return ErrInvalidSpacePrice
	}

	space, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("SpaceService.Update: %w", err)
	}
	if space == nil || !space.IsActive {
		return ErrNotFound
	}

	if err := s.repo.Update(ctx, id, name, description, pricePerSlot); err != nil {
		return fmt.Errorf("SpaceService.Update: %w", err)
	}

	return nil
}

func (s *SpaceService) Deactivate(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidSpaceID
	}

	space, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("SpaceService.Deactivate: %w", err)
	}

	if space == nil || !space.IsActive {
		return ErrNotFound
	}

	if err := s.repo.Deactivate(ctx, id); err != nil {
		return fmt.Errorf("SpaceService.Deactivate: %w", err)
	}

	return nil
}

func isValidSpaceType(spaceType string) bool {
	switch spaceType {
	case domain.SpaceTypePadel, domain.SpaceTypeFutbol, domain.SpaceTypeQuincho:
		return true
	default:
		return false
	}
}
