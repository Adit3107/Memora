package routes

import (
	"memora-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")

	api.GET("/health", handlers.HealthCheck)
}
