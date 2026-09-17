package services

import (
	"errors"
	"strings"
	"time"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

var ErrValidation = errors.New("validation failed")

type UserService struct {
	repo *repository.UserRepository
}

type CreateUserInput struct {
	Name  string
	Email string
}

type UpdateUserInput struct {
	Name  string
	Email string
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(input CreateUserInput) (models.User, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	if name == "" || email == "" {
		return models.User{}, ErrValidation
	}

	now := time.Now().UTC()
	return s.repo.Create(models.User{
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}), nil
}

func (s *UserService) List() []models.User {
	return s.repo.List()
}

func (s *UserService) GetByID(id string) (models.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) Update(id string, input UpdateUserInput) (models.User, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	if name == "" || email == "" {
		return models.User{}, ErrValidation
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.User{}, err
	}

	existing.Name = name
	existing.Email = email
	existing.UpdatedAt = time.Now().UTC()

	return s.repo.Update(id, existing)
}

func (s *UserService) Delete(id string) error {
	return s.repo.Delete(id)
}
