package handlers

import (
	"net/http"

	"memora-backend/internal/models"
	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ContentHandler struct {
	service           *services.ContentService
	contentTagService *services.ContentTagService
	summaryService    *services.SummaryService
}

type contentRequest struct {
	UserID       string             `json:"user_id"`
	SpaceID      string             `json:"space_id"`
	Name         string             `json:"name"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Type         models.ContentType `json:"type"`
	SourceURL    *string            `json:"source_url"`
	ThumbnailURL *string            `json:"thumbnail_url"`
}

func NewContentHandler(service *services.ContentService, contentTagService *services.ContentTagService, summaryService *services.SummaryService) *ContentHandler {
	return &ContentHandler{
		service:           service,
		contentTagService: contentTagService,
		summaryService:    summaryService,
	}
}

func (h *ContentHandler) Create(c *gin.Context) {
	var req contentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	content, err := h.service.Create(toCreateContentInput(req))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Content created", content)
}

func (h *ContentHandler) List(c *gin.Context) {
	userID := c.Query("user_id")
	var content []models.Content
	var err error
	if userID != "" {
		content, err = h.service.ListByUserID(userID)
	} else {
		content, err = h.service.List()
	}
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Content retrieved", content)
}

func (h *ContentHandler) GetByID(c *gin.Context) {
	content, err := h.service.GetByID(c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Content retrieved", content)
}

func (h *ContentHandler) Update(c *gin.Context) {
	var req contentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	content, err := h.service.Update(c.Param("id"), toCreateContentInput(req))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Content updated", content)
}

func (h *ContentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		handleServiceError(c, err)
		return
	}

	if err := h.contentTagService.RemoveContent(id); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Content deleted", nil)
}

func (h *ContentHandler) GetSummary(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "Content ID is required", "")
		return
	}

	if h.summaryService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Summary service not configured", "")
		return
	}

	summary, err := h.summaryService.GenerateContentSummary(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Summary generated successfully", summary)
}

func toCreateContentInput(req contentRequest) services.CreateContentInput {
	return services.CreateContentInput{
		UserID:       req.UserID,
		SpaceID:      req.SpaceID,
		Name:         req.Name,
		Title:        req.Title,
		Description:  req.Description,
		Type:         req.Type,
		SourceURL:    req.SourceURL,
		ThumbnailURL: req.ThumbnailURL,
	}
}
