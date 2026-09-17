package repository

import (
	"sync"

	"memora-backend/internal/models"
)

type ContentRepository struct {
	mu       sync.RWMutex
	contents map[string]models.Content
	nextID   int
}

func NewContentRepository() *ContentRepository {
	return &ContentRepository{
		contents: make(map[string]models.Content),
		nextID:   1,
	}
}

func (r *ContentRepository) Create(content models.Content) models.Content {
	r.mu.Lock()
	defer r.mu.Unlock()

	content.ID = nextStringID(&r.nextID)
	r.contents[content.ID] = content

	return content
}

func (r *ContentRepository) List() []models.Content {
	r.mu.RLock()
	defer r.mu.RUnlock()

	contents := make([]models.Content, 0, len(r.contents))
	for _, content := range r.contents {
		contents = append(contents, content)
	}

	return contents
}

func (r *ContentRepository) GetByID(id string) (models.Content, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	content, ok := r.contents[id]
	if !ok {
		return models.Content{}, ErrNotFound
	}

	return content, nil
}

func (r *ContentRepository) Update(id string, content models.Content) (models.Content, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.contents[id]
	if !ok {
		return models.Content{}, ErrNotFound
	}

	content.ID = id
	content.CreatedAt = existing.CreatedAt
	r.contents[id] = content

	return content, nil
}

func (r *ContentRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.contents[id]; !ok {
		return ErrNotFound
	}

	delete(r.contents, id)
	return nil
}
