package services

import (
	"strings"
	"time"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type TagService struct {
	repo           *repository.TagRepository
	contentTagRepo *repository.ContentTagRepository
}

type CreateTagInput struct {
	UserID string
	Name   string
}

type UpdateTagInput struct {
	UserID string
	Name   string
}

func NewTagService(repo *repository.TagRepository, contentTagRepo *repository.ContentTagRepository) *TagService {
	return &TagService{
		repo:           repo,
		contentTagRepo: contentTagRepo,
	}
}

func (s *TagService) Create(input CreateTagInput) (models.Tag, error) {
	userID := strings.TrimSpace(input.UserID)
	name := normalizeTagName(input.Name)
	if userID == "" || name == "" {
		return models.Tag{}, ErrValidation
	}

	return s.repo.Create(models.Tag{
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	})
}

func (s *TagService) List() ([]models.Tag, error) {
	return s.repo.List()
}

func (s *TagService) ListByUserID(userID string) ([]models.Tag, error) {
	return s.repo.ListByUserID(strings.TrimSpace(userID))
}

func (s *TagService) GetByID(id string) (models.Tag, error) {
	return s.repo.GetByID(id)
}

func (s *TagService) Update(id string, input UpdateTagInput) (models.Tag, error) {
	userID := strings.TrimSpace(input.UserID)
	name := normalizeTagName(input.Name)
	if userID == "" || name == "" {
		return models.Tag{}, ErrValidation
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Tag{}, err
	}

	existing.UserID = userID
	existing.Name = name

	return s.repo.Update(id, existing)
}

func (s *TagService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	return s.contentTagRepo.RemoveTagEverywhere(id)
}

func normalizeTagName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "#")
	return strings.TrimSpace(name)
}
