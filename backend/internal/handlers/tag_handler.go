package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	service           *services.TagService
	contentTagService *services.ContentTagService
}

type tagRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func NewTagHandler(service *services.TagService, contentTagService *services.ContentTagService) *TagHandler {
	return &TagHandler{
		service:           service,
		contentTagService: contentTagService,
	}
}

func (h *TagHandler) Create(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tag, err := h.service.Create(services.CreateTagInput{
		UserID: req.UserID,
		Name:   req.Name,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Tag created", tag)
}

func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.service.List()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Tags retrieved", tags)
}

func (h *TagHandler) GetByID(c *gin.Context) {
	tag, err := h.service.GetByID(c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Tag retrieved", tag)
}

func (h *TagHandler) Update(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tag, err := h.service.Update(c.Param("id"), services.UpdateTagInput{
		UserID: req.UserID,
		Name:   req.Name,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Tag updated", tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Tag deleted", nil)
}

func (h *TagHandler) ListContent(c *gin.Context) {
	content, err := h.contentTagService.ListContentByTag(c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Tagged content retrieved", content)
}
