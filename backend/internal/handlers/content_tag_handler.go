package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ContentTagHandler struct {
	service *services.ContentTagService
}

type setContentTagsRequest struct {
	TagIDs []string `json:"tag_ids"`
}

func NewContentTagHandler(service *services.ContentTagService) *ContentTagHandler {
	return &ContentTagHandler{service: service}
}

func (h *ContentTagHandler) SetTags(c *gin.Context) {
	var req setContentTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	content, err := h.service.SetContentTags(c.Param("id"), services.SetContentTagsInput{
		TagIDs: req.TagIDs,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Content tags updated", content)
}

func (h *ContentTagHandler) ListTags(c *gin.Context) {
	tags, err := h.service.ListTagsForContent(c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Content tags retrieved", tags)
}

func (h *ContentTagHandler) RemoveTag(c *gin.Context) {
	if err := h.service.RemoveTagFromContent(c.Param("id"), c.Param("tagID")); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Content tag removed", nil)
}
