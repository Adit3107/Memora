package main

import (
	"fmt"

	"memora-backend/internal/config"
	"memora-backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	router := gin.Default()
	routes.RegisterRoutes(router)

	fmt.Printf("Server is running on port %s\n", cfg.Port)
	router.Run(":" + cfg.Port)
}
