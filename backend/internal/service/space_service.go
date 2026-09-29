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
	GetAvailableSlotsForDate(ctx context.Context, spaceID int64, date string) ([]domain.SpaceSlot, error)
	Update(ctx context.Context, id int64, name string, description *string, pricePerSlot float64) error
	CreateSlot(ctx context.Context, slot *domain.SpaceSlot) (int64, error)
	Deactivate(ctx context.Context, id int64) error
	HasBookings(ctx context.Context, id int64) (bool, error)
	HardDelete(ctx context.Context, id int64) error
}

type SpaceService struct {
	repo SpaceRepo
}

func NewSpaceService(repo SpaceRepo) *SpaceService {
	return &SpaceService{repo: repo}
}

type slotTemplate struct {
	label     string
	startTime string
	endTime   string
}

func defaultSlotsFor(spaceType string) []slotTemplate {
	switch spaceType {
	case domain.SpaceTypePadel, domain.SpaceTypeFutbol, domain.SpaceTypePadbol, domain.SpaceTypeBeachVoley:
		hours := []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}
		slots := make([]slotTemplate, len(hours))
		for i, h := range hours {
			next := (h + 1) % 24
			endLabel := fmt.Sprintf("%02d:00", next)
			endTime := fmt.Sprintf("%02d:00:00", next)
			if next == 0 {
				endTime = "24:00:00"
			}
			slots[i] = slotTemplate{
				label:     fmt.Sprintf("%02d:00–%s", h, endLabel),
				startTime: fmt.Sprintf("%02d:00:00", h),
				endTime:   endTime,
			}
		}
		return slots
	case domain.SpaceTypeQuincho:
		return []slotTemplate{
			{label: "Mañana", startTime: "10:00:00", endTime: "15:00:00"},
			{label: "Tarde", startTime: "16:00:00", endTime: "20:00:00"},
			// 24:00:00 is valid Postgres TIME and satisfies end_time > start_time
			{label: "Noche", startTime: "21:00:00", endTime: "24:00:00"},
		}
	default:
		return nil
	}
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

	templates := defaultSlotsFor(spaceType)
	fmt.Printf("[DEBUG] spaceType=%q templates=%d\n", spaceType, len(templates))
	for _, t := range templates {
		start, end := t.startTime, t.endTime
		slot := &domain.SpaceSlot{
			SpaceID:   id,
			Label:     t.label,
			StartTime: &start,
			EndTime:   &end,
		}
		if _, slotErr := s.repo.CreateSlot(ctx, slot); slotErr != nil {
			fmt.Printf("[DEBUG] CreateSlot error: %v\n", slotErr)
			_ = s.repo.HardDelete(ctx, id)
			return 0, fmt.Errorf("SpaceService.Create: create default slots: %w", slotErr)
		}
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

// GetSlots devuelve los turnos de un espacio.
// Si se provee date (YYYY-MM-DD), filtra los slots ya reservados para esa fecha.
func (s *SpaceService) GetSlots(ctx context.Context, spaceID int64, date string) ([]domain.SpaceSlot, error) {
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

	if date != "" {
		slots, err := s.repo.GetAvailableSlotsForDate(ctx, spaceID, date)
		if err != nil {
			return nil, fmt.Errorf("SpaceService.GetSlots: %w", err)
		}
		return slots, nil
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

func (s *SpaceService) HardDelete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidSpaceID
	}

	space, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("SpaceService.HardDelete: %w", err)
	}
	if space == nil {
		return ErrNotFound
	}

	hasBookings, err := s.repo.HasBookings(ctx, id)
	if err != nil {
		return fmt.Errorf("SpaceService.HardDelete: %w", err)
	}
	if hasBookings {
		return ErrSpaceHasBookings
	}

	if err := s.repo.HardDelete(ctx, id); err != nil {
		return fmt.Errorf("SpaceService.HardDelete: %w", err)
	}

	return nil
}

func isValidSpaceType(spaceType string) bool {
	switch spaceType {
	case domain.SpaceTypePadel, domain.SpaceTypeFutbol, domain.SpaceTypePadbol, domain.SpaceTypeBeachVoley, domain.SpaceTypeQuincho:
		return true
	default:
		return false
	}
}
