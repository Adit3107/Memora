package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"memora-backend/internal/config"
	"memora-backend/internal/database"
	"memora-backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database config error: %v", err)
	}
	defer db.Close()

	if err := database.Ping(ctx, db); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.RunMigrations(ctx, db); err != nil {
		log.Fatalf("database migrations failed: %v", err)
	}

	router := gin.Default()
	router.Use(corsMiddleware())
	routes.RegisterRoutes(router, db)

	fmt.Printf("Server is running on port %s\n", cfg.Port)
	router.Run(":" + cfg.Port)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
