package services

import (
	"strings"
	"time"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type ContentService struct {
	repo *repository.ContentRepository
}

type CreateContentInput struct {
	UserID       string
	SpaceID      string
	Title        string
	Description  string
	Type         models.ContentType
	SourceURL    *string
	ThumbnailURL *string
}

type UpdateContentInput = CreateContentInput

func NewContentService(repo *repository.ContentRepository) *ContentService {
	return &ContentService{repo: repo}
}

func (s *ContentService) Create(input CreateContentInput) (models.Content, error) {
	content, err := buildContentFromInput(input)
	if err != nil {
		return models.Content{}, err
	}

	now := time.Now().UTC()
	content.CreatedAt = now
	content.UpdatedAt = now

	return s.repo.Create(content)
}

func (s *ContentService) List() ([]models.Content, error) {
	return s.repo.List()
}

func (s *ContentService) GetByID(id string) (models.Content, error) {
	return s.repo.GetByID(id)
}

func (s *ContentService) Update(id string, input UpdateContentInput) (models.Content, error) {
	content, err := buildContentFromInput(input)
	if err != nil {
		return models.Content{}, err
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Content{}, err
	}

	content.CreatedAt = existing.CreatedAt
	content.UpdatedAt = time.Now().UTC()

	return s.repo.Update(id, content)
}

func (s *ContentService) Delete(id string) error {
	return s.repo.Delete(id)
}

func buildContentFromInput(input CreateContentInput) (models.Content, error) {
	userID := strings.TrimSpace(input.UserID)
	spaceID := strings.TrimSpace(input.SpaceID)
	title := strings.TrimSpace(input.Title)
	if userID == "" || spaceID == "" || title == "" || !isValidContentType(input.Type) {
		return models.Content{}, ErrValidation
	}

	return models.Content{
		UserID:       userID,
		SpaceID:      spaceID,
		Title:        title,
		Description:  strings.TrimSpace(input.Description),
		Type:         input.Type,
		SourceURL:    trimOptionalString(input.SourceURL),
		ThumbnailURL: trimOptionalString(input.ThumbnailURL),
	}, nil
}

func isValidContentType(contentType models.ContentType) bool {
	switch contentType {
	case models.ContentTypeVideo,
		models.ContentTypeDocument,
		models.ContentTypeArticle,
		models.ContentTypeImage:
		return true
	default:
		return false
	}
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
