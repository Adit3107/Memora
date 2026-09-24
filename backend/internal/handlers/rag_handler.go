package handlers

import (
	"net/http"

	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type RAGHandler struct {
	service *services.RAGService
}

type ragRequest struct {
	UserID   string                `json:"user_id"`
	Question string                `json:"question"`
	Scope    services.RAGScope     `json:"scope"`
	History  []services.RAGMessage `json:"history"`
	TopK     int                   `json:"top_k"`
}

func NewRAGHandler(service *services.RAGService) *RAGHandler {
	return &RAGHandler{service: service}
}

func (h *RAGHandler) Ask(c *gin.Context) {
	var req ragRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.Ask(c.Request.Context(), services.RAGInput{
		UserID:   req.UserID,
		Question: req.Question,
		Scope:    req.Scope,
		History:  req.History,
		TopK:     req.TopK,
	})
	if err != nil {
		if services.IsRAGClientError(err) {
			response.Error(c, http.StatusBadRequest, "Invalid RAG request", err.Error())
			return
		}
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "RAG answer generated", result)
}

// Why this file exists:
// The RAG API stays in Go so the frontend never calls an LLM provider directly.
