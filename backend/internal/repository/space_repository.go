package repository

import (
	"sync"

	"memora-backend/internal/models"
)

type SpaceRepository struct {
	mu     sync.RWMutex
	spaces map[string]models.Space
	nextID int
}

func NewSpaceRepository() *SpaceRepository {
	return &SpaceRepository{
		spaces: make(map[string]models.Space),
		nextID: 1,
	}
}

func (r *SpaceRepository) Create(space models.Space) models.Space {
	r.mu.Lock()
	defer r.mu.Unlock()

	space.ID = nextStringID(&r.nextID)
	r.spaces[space.ID] = space

	return space
}

func (r *SpaceRepository) List() []models.Space {
	r.mu.RLock()
	defer r.mu.RUnlock()

	spaces := make([]models.Space, 0, len(r.spaces))
	for _, space := range r.spaces {
		spaces = append(spaces, space)
	}

	return spaces
}

func (r *SpaceRepository) GetByID(id string) (models.Space, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	space, ok := r.spaces[id]
	if !ok {
		return models.Space{}, ErrNotFound
	}

	return space, nil
}

func (r *SpaceRepository) Update(id string, space models.Space) (models.Space, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.spaces[id]
	if !ok {
		return models.Space{}, ErrNotFound
	}

	space.ID = id
	space.CreatedAt = existing.CreatedAt
	r.spaces[id] = space

	return space, nil
}

func (r *SpaceRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.spaces[id]; !ok {
		return ErrNotFound
	}

	delete(r.spaces, id)
	return nil
}
