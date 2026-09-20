package handlers

import (
	"net/http"

	"memora-backend/internal/response"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "Server is healthy", gin.H{
		"status": "ok",
	})
}
