package main

import (
	"log"

	"api-service/cache"
	"api-service/config"
	"api-service/database"
	"api-service/handlers"
	"api-service/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.Init(cfg.DatabaseURL); err != nil {
		log.Println("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize Redis
	if err := cache.Init(cfg.RedisURL); err != nil {
		log.Println("Failed to connect to Redis:", err)
	}
	defer cache.Close()

	// Initialize handlers with cache TTL
	handlers.Init(cfg.CacheTTL)

	// Setup Gin router
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.CORS())

	// Routes
	r.GET("/health", handlers.HealthCheck)
	r.POST("/api/urls", middleware.RateLimit(cfg.RateLimitRequests, cfg.RateLimitWindow), handlers.CreateURL)
	r.GET("/api/urls", handlers.ListURLs)
	r.GET("/:code", handlers.Redirect)

	log.Printf("API Service starting on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
