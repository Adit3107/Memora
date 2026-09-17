package services

import (
	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type ContentTagService struct {
	contentRepo    *repository.ContentRepository
	tagRepo        *repository.TagRepository
	contentTagRepo *repository.ContentTagRepository
}

type SetContentTagsInput struct {
	TagIDs []string
}

func NewContentTagService(
	contentRepo *repository.ContentRepository,
	tagRepo *repository.TagRepository,
	contentTagRepo *repository.ContentTagRepository,
) *ContentTagService {
	return &ContentTagService{
		contentRepo:    contentRepo,
		tagRepo:        tagRepo,
		contentTagRepo: contentTagRepo,
	}
}

func (s *ContentTagService) SetContentTags(contentID string, input SetContentTagsInput) (models.ContentWithTags, error) {
	content, err := s.contentRepo.GetByID(contentID)
	if err != nil {
		return models.ContentWithTags{}, err
	}

	tags, err := s.resolveTagsForContent(content.UserID, input.TagIDs)
	if err != nil {
		return models.ContentWithTags{}, err
	}

	tagIDs := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagIDs = append(tagIDs, tag.ID)
	}

	s.contentTagRepo.SetTags(contentID, tagIDs)

	return models.ContentWithTags{
		Content: content,
		Tags:    tags,
	}, nil
}

func (s *ContentTagService) ListTagsForContent(contentID string) ([]models.Tag, error) {
	content, err := s.contentRepo.GetByID(contentID)
	if err != nil {
		return nil, err
	}

	tagIDs := s.contentTagRepo.ListTagIDs(contentID)
	return s.resolveTagsForContent(content.UserID, tagIDs)
}

func (s *ContentTagService) RemoveTagFromContent(contentID string, tagID string) error {
	content, err := s.contentRepo.GetByID(contentID)
	if err != nil {
		return err
	}

	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return err
	}

	if tag.UserID != content.UserID {
		return ErrValidation
	}

	s.contentTagRepo.RemoveTag(contentID, tagID)
	return nil
}

func (s *ContentTagService) ListContentByTag(tagID string) ([]models.ContentWithTags, error) {
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return nil, err
	}

	contentIDs := s.contentTagRepo.ListContentIDsByTag(tagID)
	results := make([]models.ContentWithTags, 0, len(contentIDs))

	for _, contentID := range contentIDs {
		content, err := s.contentRepo.GetByID(contentID)
		if err != nil {
			continue
		}

		if content.UserID != tag.UserID {
			continue
		}

		tags, err := s.ListTagsForContent(contentID)
		if err != nil {
			continue
		}

		results = append(results, models.ContentWithTags{
			Content: content,
			Tags:    tags,
		})
	}

	return results, nil
}

func (s *ContentTagService) RemoveContent(contentID string) {
	s.contentTagRepo.RemoveContent(contentID)
}

func (s *ContentTagService) resolveTagsForContent(userID string, tagIDs []string) ([]models.Tag, error) {
	if len(tagIDs) == 0 {
		return []models.Tag{}, nil
	}

	seen := make(map[string]bool, len(tagIDs))
	tags := make([]models.Tag, 0, len(tagIDs))

	for _, tagID := range tagIDs {
		if tagID == "" || seen[tagID] {
			continue
		}

		tag, err := s.tagRepo.GetByID(tagID)
		if err != nil {
			return nil, err
		}

		if tag.UserID != userID {
			return nil, ErrValidation
		}

		seen[tagID] = true
		tags = append(tags, tag)
	}

	return tags, nil
}
