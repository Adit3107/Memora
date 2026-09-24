package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service      *services.UserService
	spaceService *services.SpaceService
}

type userRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type syncUserRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUserHandler(service *services.UserService, spaceService *services.SpaceService) *UserHandler {
	return &UserHandler{
		service:      service,
		spaceService: spaceService,
	}
}

func (h *UserHandler) Sync(c *gin.Context) {
	var req syncUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.Sync(req.ID, req.Name, req.Email)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Ensure user has at least one space
	spaces, err := h.spaceService.ListByUserID(user.ID)
	if err == nil && len(spaces) == 0 {
		defaultSpace, createErr := h.spaceService.Create(services.CreateSpaceInput{
			UserID:      user.ID,
			Name:        "Personal",
			Description: "Default space for your saved videos and documents",
		})
		if createErr == nil {
			spaces = append(spaces, defaultSpace)
		}
	}

	response.Success(c, http.StatusOK, "User synchronized", gin.H{
		"user":   user,
		"spaces": spaces,
	})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.Create(services.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User created", user)
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.List()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Users retrieved", users)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	user, err := h.service.GetByID(c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User retrieved", user)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.Update(c.Param("id"), services.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User updated", user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User deleted", nil)
}
