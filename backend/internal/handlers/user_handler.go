package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *services.UserService
}

type userRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
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
	response.Success(c, http.StatusOK, "Users retrieved", h.service.List())
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
