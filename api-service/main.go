package main

import (
	"context"
	"log"

	"api-service/cache"
	"api-service/config"
	"api-service/database"
	"api-service/handlers"
	"api-service/middleware"
	"api-service/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize OpenTelemetry tracing
	log.Printf("Setting up OpenTelemetry tracing to: %s", cfg.OTELExporterURL)
	shutdown, err := tracing.Init("api-service", cfg.OTELExporterURL)
	if err != nil {
		log.Printf("Failed to initialize tracing: %v", err)
	} else {
		log.Printf("OpenTelemetry tracing successfully initialized")
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				log.Printf("Failed to shutdown tracer: %v", err)
			}
		}()
	}

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
	r.Use(otelgin.Middleware("api-service")) // OpenTelemetry auto-instrumentation

	// Routes
	r.GET("/health", handlers.HealthCheck)
	r.POST("/api/urls", middleware.RateLimit(cfg.RateLimitRequests, cfg.RateLimitWindow), handlers.CreateURL)
	r.GET("/api/urls", handlers.ListURLs)
	r.GET("/:code", handlers.Redirect)

	log.Printf("API Service starting on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
