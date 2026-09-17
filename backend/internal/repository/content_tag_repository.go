package repository

import "sync"

type ContentTagRepository struct {
	mu            sync.RWMutex
	contentToTags map[string]map[string]bool
}

func NewContentTagRepository() *ContentTagRepository {
	return &ContentTagRepository{
		contentToTags: make(map[string]map[string]bool),
	}
}

func (r *ContentTagRepository) SetTags(contentID string, tagIDs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tags := make(map[string]bool, len(tagIDs))
	for _, tagID := range tagIDs {
		tags[tagID] = true
	}

	r.contentToTags[contentID] = tags
}

func (r *ContentTagRepository) ListTagIDs(contentID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tagSet := r.contentToTags[contentID]
	tagIDs := make([]string, 0, len(tagSet))
	for tagID := range tagSet {
		tagIDs = append(tagIDs, tagID)
	}

	return tagIDs
}

func (r *ContentTagRepository) RemoveTag(contentID string, tagID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tagSet := r.contentToTags[contentID]
	if tagSet == nil {
		return
	}

	delete(tagSet, tagID)
	if len(tagSet) == 0 {
		delete(r.contentToTags, contentID)
	}
}

func (r *ContentTagRepository) ListContentIDsByTag(tagID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	contentIDs := make([]string, 0)
	for contentID, tagSet := range r.contentToTags {
		if tagSet[tagID] {
			contentIDs = append(contentIDs, contentID)
		}
	}

	return contentIDs
}

func (r *ContentTagRepository) RemoveTagEverywhere(tagID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for contentID, tagSet := range r.contentToTags {
		delete(tagSet, tagID)
		if len(tagSet) == 0 {
			delete(r.contentToTags, contentID)
		}
	}
}

func (r *ContentTagRepository) RemoveContent(contentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.contentToTags, contentID)
}
