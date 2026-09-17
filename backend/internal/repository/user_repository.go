package repository

import (
	"sync"

	"memora-backend/internal/models"
)

type UserRepository struct {
	mu     sync.RWMutex
	users  map[string]models.User
	nextID int
}

func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		users:  make(map[string]models.User),
		nextID: 1,
	}

	return repo
}

func (r *UserRepository) Create(user models.User) models.User {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = nextStringID(&r.nextID)
	r.users[user.ID] = user

	return user
}

func (r *UserRepository) List() []models.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users
}

func (r *UserRepository) GetByID(id string) (models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}

	return user, nil
}

func (r *UserRepository) Update(id string, user models.User) (models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}

	user.ID = id
	user.CreatedAt = existing.CreatedAt
	r.users[id] = user

	return user, nil
}

func (r *UserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrNotFound
	}

	delete(r.users, id)
	return nil
}
