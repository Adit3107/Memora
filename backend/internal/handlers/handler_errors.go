package handlers

import (
	"errors"
	"net/http"

	"memora-backend/internal/repository"
	"memora-backend/internal/response"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrValidation):
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Resource not found", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
	}
}
