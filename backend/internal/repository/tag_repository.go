package repository

import (
	"sync"

	"memora-backend/internal/models"
)

type TagRepository struct {
	mu     sync.RWMutex
	tags   map[string]models.Tag
	nextID int
}

func NewTagRepository() *TagRepository {
	return &TagRepository{
		tags:   make(map[string]models.Tag),
		nextID: 1,
	}
}

func (r *TagRepository) Create(tag models.Tag) models.Tag {
	r.mu.Lock()
	defer r.mu.Unlock()

	tag.ID = nextStringID(&r.nextID)
	r.tags[tag.ID] = tag

	return tag
}

func (r *TagRepository) List() []models.Tag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]models.Tag, 0, len(r.tags))
	for _, tag := range r.tags {
		tags = append(tags, tag)
	}

	return tags
}

func (r *TagRepository) GetByID(id string) (models.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tag, ok := r.tags[id]
	if !ok {
		return models.Tag{}, ErrNotFound
	}

	return tag, nil
}

func (r *TagRepository) Update(id string, tag models.Tag) (models.Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.tags[id]
	if !ok {
		return models.Tag{}, ErrNotFound
	}

	tag.ID = id
	tag.CreatedAt = existing.CreatedAt
	r.tags[id] = tag

	return tag, nil
}

func (r *TagRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tags[id]; !ok {
		return ErrNotFound
	}

	delete(r.tags, id)
	return nil
}
