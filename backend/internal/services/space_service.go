package services

import (
	"strings"
	"time"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type SpaceService struct {
	repo *repository.SpaceRepository
}

type CreateSpaceInput struct {
	UserID      string
	Name        string
	Description string
}

type UpdateSpaceInput struct {
	UserID      string
	Name        string
	Description string
}

func NewSpaceService(repo *repository.SpaceRepository) *SpaceService {
	return &SpaceService{repo: repo}
}

func (s *SpaceService) Create(input CreateSpaceInput) (models.Space, error) {
	userID := strings.TrimSpace(input.UserID)
	name := strings.TrimSpace(input.Name)
	if userID == "" || name == "" {
		return models.Space{}, ErrValidation
	}

	now := time.Now().UTC()
	return s.repo.Create(models.Space{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *SpaceService) List() ([]models.Space, error) {
	return s.repo.List()
}

func (s *SpaceService) GetByID(id string) (models.Space, error) {
	return s.repo.GetByID(id)
}

func (s *SpaceService) Update(id string, input UpdateSpaceInput) (models.Space, error) {
	userID := strings.TrimSpace(input.UserID)
	name := strings.TrimSpace(input.Name)
	if userID == "" || name == "" {
		return models.Space{}, ErrValidation
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Space{}, err
	}

	existing.UserID = userID
	existing.Name = name
	existing.Description = strings.TrimSpace(input.Description)
	existing.UpdatedAt = time.Now().UTC()

	return s.repo.Update(id, existing)
}

func (s *SpaceService) Delete(id string) error {
	return s.repo.Delete(id)
}
