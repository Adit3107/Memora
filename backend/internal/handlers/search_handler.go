package handlers

import (
	"net/http"
	"time"

	"memora-backend/internal/models"
	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service *services.SearchService
}

type searchRequest struct {
	UserID      string              `json:"user_id"`
	Query       string              `json:"query"`
	Mode        services.SearchMode `json:"mode"`
	SpaceID     *string             `json:"space_id"`
	ContentIDs  []string            `json:"content_ids"`
	ContentType *models.ContentType `json:"content_type"`
	SourceType  *string             `json:"source_type"`
	TagIDs      []string            `json:"tag_ids"`
	CreatedFrom *string             `json:"created_from"`
	CreatedTo   *string             `json:"created_to"`
	Limit       int                 `json:"limit"`
	Offset      int                 `json:"offset"`
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

func (h *SearchHandler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	input, err := toSearchInput(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid search request", err.Error())
		return
	}

	results, err := h.service.Search(c.Request.Context(), input)
	if err != nil {
		if services.IsSearchClientError(err) {
			response.Error(c, http.StatusBadRequest, "Search failed", err.Error())
			return
		}
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Search completed", results)
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
		if services.IsSearchClientError(err) {
			response.Error(c, http.StatusBadRequest, "Semantic search failed", err.Error())
			return
		}
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Semantic search completed", results)
}

func toSearchInput(req searchRequest) (services.SearchInput, error) {
	createdFrom, err := parseOptionalTime(req.CreatedFrom)
	if err != nil {
		return services.SearchInput{}, err
	}
	createdTo, err := parseOptionalTime(req.CreatedTo)
	if err != nil {
		return services.SearchInput{}, err
	}

	return services.SearchInput{
		UserID:      req.UserID,
		Query:       req.Query,
		Mode:        req.Mode,
		SpaceID:     req.SpaceID,
		ContentIDs:  req.ContentIDs,
		ContentType: req.ContentType,
		SourceType:  req.SourceType,
		TagIDs:      req.TagIDs,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}, nil
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// Why this file exists:
// The frontend asks Go for search. Go may use Python internally for embeddings,
// but that implementation detail stays behind this handler.
