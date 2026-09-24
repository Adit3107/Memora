package services

import (
	"net/url"
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
	Name         string
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
	sourceURL := trimOptionalString(input.SourceURL)
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = contentDisplayName(title, input.Type, sourceURL)
	}

	return models.Content{
		UserID:       userID,
		SpaceID:      spaceID,
		Name:         name,
		Title:        title,
		Description:  strings.TrimSpace(input.Description),
		Type:         input.Type,
		SourceURL:    sourceURL,
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

func looksLikeURL(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func contentDisplayName(title string, contentType models.ContentType, sourceURL *string) string {
	title = strings.TrimSpace(title)
	if title != "" && !looksLikeURL(title) {
		return title
	}

	rawURL := title
	if sourceURL != nil && strings.TrimSpace(*sourceURL) != "" {
		rawURL = strings.TrimSpace(*sourceURL)
	}

	switch contentType {
	case models.ContentTypeVideo:
		return youtubeDisplayName(rawURL)
	case models.ContentTypeDocument:
		return fallbackTypedName("Document", title)
	case models.ContentTypeImage:
		return fallbackTypedName("Image", title)
	case models.ContentTypeArticle:
		return fallbackTypedName("Article", title)
	default:
		return fallbackTypedName("Saved content", title)
	}
}

func youtubeDisplayName(rawURL string) string {
	videoID, isShort := youtubeLabelParts(rawURL)
	if isShort && videoID != "" {
		return "YouTube Short " + videoID
	}
	if videoID != "" {
		return "YouTube video " + videoID
	}
	return "YouTube video"
}

func youtubeLabelParts(rawURL string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", false
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if host == "youtu.be" && len(parts) > 0 {
		return cleanLabelID(parts[0]), false
	}
	if !strings.HasSuffix(host, "youtube.com") {
		return "", false
	}
	if len(parts) >= 2 && parts[0] == "shorts" {
		return cleanLabelID(parts[1]), true
	}
	if len(parts) >= 2 && (parts[0] == "embed" || parts[0] == "live") {
		return cleanLabelID(parts[1]), false
	}
	if parsed.Path == "/watch" {
		return cleanLabelID(parsed.Query().Get("v")), false
	}
	return "", false
}

func cleanLabelID(value string) string {
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, "/?#&=") {
		return ""
	}
	return value
}

func fallbackTypedName(prefix string, title string) string {
	if title != "" && !looksLikeURL(title) {
		return title
	}
	return prefix
}
