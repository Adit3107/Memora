package routes

import (
	"database/sql"

	"memora-backend/internal/config"
	"memora-backend/internal/handlers"
	"memora-backend/internal/repository"
	"memora-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *sql.DB, cfg config.Config) {
	userRepo := repository.NewPostgresUserRepository(db)
	spaceRepo := repository.NewPostgresSpaceRepository(db)
	contentRepo := repository.NewPostgresContentRepository(db)
	tagRepo := repository.NewPostgresTagRepository(db)
	contentTagRepo := repository.NewPostgresContentTagRepository(db)
	ingestionRepo := repository.NewPostgresIngestionRepository(db, cfg.EmbeddingDimension)
	searchRepo := repository.NewPostgresSearchRepository(db)

	userService := services.NewUserService(userRepo)
	spaceService := services.NewSpaceService(spaceRepo)
	userHandler := handlers.NewUserHandler(userService, spaceService)
	spaceHandler := handlers.NewSpaceHandler(spaceService, userService)
	contentTagService := services.NewContentTagService(contentRepo, tagRepo, contentTagRepo)
	summaryService := services.NewSummaryService(db, cfg.GeminiAPIKey, cfg.GeminiModel)
	contentHandler := handlers.NewContentHandler(services.NewContentService(contentRepo), contentTagService, summaryService)
	tagHandler := handlers.NewTagHandler(services.NewTagService(tagRepo, contentTagRepo), contentTagService)
	contentTagHandler := handlers.NewContentTagHandler(contentTagService)
	ingestionHandler := handlers.NewIngestionHandler(services.NewIngestionService(contentRepo, spaceRepo, ingestionRepo, cfg.AIServiceURL, cfg.EmbeddingDimension, cfg.EmbeddingMaxConcurrency), userService)
	searchService := services.NewSearchService(searchRepo, cfg.AIServiceURL, cfg.EmbeddingDimension)
	searchHandler := handlers.NewSearchHandler(searchService)
	// === PAUSED FEATURE: RAG Service (Put on hold for upcoming release) ===
	// ragHandler := handlers.NewRAGHandler(services.NewRAGService(
	// 	searchService,
	// 	services.NewGeminiLLMService(cfg.GeminiAPIKey, cfg.GeminiModel),
	// 	services.RAGConfig{
	// 		TopK:               cfg.RAG.TopK,
	// 		HistoryMessages:    cfg.RAG.HistoryMessages,
	// 		ContextMaxChars:    cfg.RAG.ContextMaxChars,
	// 		ChunkMaxChars:      cfg.RAG.ChunkMaxChars,
	// 		AllowLocalFallback: cfg.RAG.AllowLocalFallback,
	// 	},
	// ))

	api := router.Group("/api")

	api.GET("/health", handlers.HealthCheck)

	users := api.Group("/users")
	users.POST("", userHandler.Create)
	users.POST("/sync", userHandler.Sync)
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
	content.PUT("/:id/tags", contentTagHandler.SetTags)
	content.GET("/:id/tags", contentTagHandler.ListTags)
	content.DELETE("/:id/tags/:tagID", contentTagHandler.RemoveTag)
	content.GET("/:id/summary", contentHandler.GetSummary)

	ingestionRoutes := api.Group("/ingestion")
	ingestionRoutes.POST("/url", ingestionHandler.IngestURL)
	ingestionRoutes.POST("/file", ingestionHandler.IngestFile)
	ingestionRoutes.GET("/:id", ingestionHandler.GetByID)

	searchRoutes := api.Group("/search")
	searchRoutes.POST("", searchHandler.Search)
	searchRoutes.POST("/semantic", searchHandler.Semantic)

	// === PAUSED FEATURE: RAG Endpoint ===
	// api.POST("/rag", ragHandler.Ask)

	tags := api.Group("/tags")
	tags.POST("", tagHandler.Create)
	tags.GET("", tagHandler.List)
	tags.GET("/:id", tagHandler.GetByID)
	tags.PUT("/:id", tagHandler.Update)
	tags.DELETE("/:id", tagHandler.Delete)
	tags.GET("/:id/content", tagHandler.ListContent)
}
