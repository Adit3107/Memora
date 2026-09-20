package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type IngestionHandler struct {
	service *services.IngestionService
}

type ingestURLRequest struct {
	UserID  string `json:"user_id"`
	SpaceID string `json:"space_id"`
	URL     string `json:"url"`
}

func NewIngestionHandler(service *services.IngestionService) *IngestionHandler {
	return &IngestionHandler{service: service}
}

func (h *IngestionHandler) IngestURL(c *gin.Context) {
	var req ingestURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.IngestURL(c.Request.Context(), services.IngestURLInput{
		UserID:  req.UserID,
		SpaceID: req.SpaceID,
		URL:     req.URL,
	})
	if err != nil {
		if services.IsIngestionClientError(err) {
			response.Error(c, http.StatusBadRequest, "Ingestion failed", err.Error())
			return
		}
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Content ingested", result)
}
