package main // This is the main package for the backend server application. purpose - It serves as the entry point for the application, initializing and starting the server. (code builds into exe prog)

import (
	"fmt"
	"memora-backend/internal/config"
	"memora-backend/internal/routes"

	"github.com/gin-gonic/gin"
) // Importing the fmt package for formatted I/O operations.
// Importing the config package from the internal directory to load configuration settings.

func main() {
	cfg := config.Load()

	router := gin.Default()
	routes.RegisterRoutes(router)
	router.Run(":" + cfg.Port) // Start the server on the specified port from the configuration.
	fmt.Println("Memora backend Starting...")
	fmt.Printf("Server is running on port %s\n", cfg.Port)
}
