package routes

import (
	"memora-backend/internal/handlers"
	"memora-backend/internal/repository"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	userRepo := repository.NewUserRepository()
	spaceRepo := repository.NewSpaceRepository()
	contentRepo := repository.NewContentRepository()

	userHandler := handlers.NewUserHandler(services.NewUserService(userRepo))
	spaceHandler := handlers.NewSpaceHandler(services.NewSpaceService(spaceRepo))
	contentHandler := handlers.NewContentHandler(services.NewContentService(contentRepo))

	api := router.Group("/api")

	api.GET("/health", handlers.HealthCheck)

	users := api.Group("/users")
	users.POST("", userHandler.Create)
	users.GET("", userHandler.List)
	users.GET("/:id", userHandler.GetByID)
	users.PUT("/:id", userHandler.Update)
	users.DELETE("/:id", userHandler.Delete)

	spaces := api.Group("/spaces")
	spaces.POST("", spaceHandler.Create)
	spaces.GET("", spaceHandler.List)
	spaces.GET("/:id", spaceHandler.GetByID)
	spaces.PUT("/:id", spaceHandler.Update)
	spaces.DELETE("/:id", spaceHandler.Delete)

	content := api.Group("/content")
	content.POST("", contentHandler.Create)
	content.GET("", contentHandler.List)
	content.GET("/:id", contentHandler.GetByID)
	content.PUT("/:id", contentHandler.Update)
	content.DELETE("/:id", contentHandler.Delete)
}
