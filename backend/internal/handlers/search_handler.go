package handlers

import (
	"net/http"

	"memora-backend/internal/ingestion"
	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service *services.SearchService
}

type semanticSearchRequest struct {
	UserID  string  `json:"user_id"`
	SpaceID *string `json:"space_id"`
	Query   string  `json:"query"`
	Limit   int     `json:"limit"`
}

func NewSearchHandler(service *services.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) Semantic(c *gin.Context) {
	var req semanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	results, err := h.service.SemanticSearch(c.Request.Context(), services.SemanticSearchInput{
		UserID:  req.UserID,
		SpaceID: req.SpaceID,
		Query:   req.Query,
		Limit:   req.Limit,
	})
	if err != nil {
		if services.IsIngestionClientError(err) || err == ingestion.ErrEmbeddingFailed {
			response.Error(c, http.StatusBadRequest, "Semantic search failed", err.Error())
			return
		}
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Semantic search completed", results)
}

// Why this file exists:
// The frontend asks Go for search. Go may use Python internally for embeddings,
// but that implementation detail stays behind this handler.
