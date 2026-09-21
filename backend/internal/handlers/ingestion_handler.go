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

func (h *IngestionHandler) IngestFile(c *gin.Context) {
	userID := c.PostForm("user_id")
	spaceID := c.PostForm("space_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid file upload", err.Error())
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	result, err := h.service.IngestFile(c.Request.Context(), services.IngestFileInput{
		UserID:      userID,
		SpaceID:     spaceID,
		FileName:    header.Filename,
		ContentType: contentType,
		Body:        file,
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

// Why this file exists:
// Handlers translate HTTP JSON into service input and service output into API responses.
// This keeps request parsing and status codes out of the ingestion business logic.
