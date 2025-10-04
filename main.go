package main

import (
	"log"
	"os"

	"utara_backend/config"
	"utara_backend/handlers"
	"utara_backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	handlers.StartAutomaticCleanup()
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to MongoDB
	config.ConnectDB()

	// Initialize Gin
	r := gin.Default()

	// ✅ Serve static files from ./static folder (place login.html & index.html here)
	r.Static("/static", "./static")

	// ✅ Map clean routes for your pages
	r.GET("/", func(c *gin.Context) {
		c.File("./static/login.html")
	})
	r.GET("/login", func(c *gin.Context) {
		c.File("./static/login.html")
	})
	r.GET("/upload", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Setup routes
	routes.SetupRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
