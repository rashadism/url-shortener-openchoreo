package handlers

import (
	"net/http"

	"api-service/cache"
	"api-service/database"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	dbHealthy := database.CheckHealth()
	redisHealthy := cache.CheckHealth()

	status := "healthy"
	statusCode := http.StatusOK

	if !dbHealthy || !redisHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":   status,
		"database": dbHealthy,
		"redis":    redisHealthy,
	})
}
