package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SpaceHandler struct {
	service *services.SpaceService
}

type spaceRequest struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewSpaceHandler(service *services.SpaceService) *SpaceHandler {
	return &SpaceHandler{service: service}
}

func (h *SpaceHandler) Create(c *gin.Context) {
	var req spaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	space, err := h.service.Create(services.CreateSpaceInput{
		UserID:      req.UserID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Space created", space)
}

func (h *SpaceHandler) List(c *gin.Context) {
	spaces, err := h.service.List()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Spaces retrieved", spaces)
}

func (h *SpaceHandler) GetByID(c *gin.Context) {
	space, err := h.service.GetByID(c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Space retrieved", space)
}

func (h *SpaceHandler) Update(c *gin.Context) {
	var req spaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	space, err := h.service.Update(c.Param("id"), services.UpdateSpaceInput{
		UserID:      req.UserID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Space updated", space)
}

func (h *SpaceHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Space deleted", nil)
}
