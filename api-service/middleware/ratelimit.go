package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"api-service/cache"
	"api-service/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func RateLimit(requests, window int) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Read body and restore it for later handlers
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore the body for the next handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to extract API key from JSON
				var req models.CreateURLRequest
				if json.Unmarshal(bodyBytes, &req) == nil {
					apiKey = req.Username
				}
			}
		}

		if apiKey == "" {
			apiKey = c.ClientIP()
		}

		// Get current count
		val, err := cache.GetRateLimit(apiKey)
		if err != nil && err != redis.Nil {
			log.Printf("Redis error in rate limiting: %v", err)
		}

		if val >= requests {
			log.Printf("[Redis] Rate limit exceeded for key: %s (%d/%d requests)", apiKey, val, requests)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Try again later.",
			})
			c.Abort()
			return
		}

		// Increment counter
		err = cache.IncrementRateLimit(apiKey, window)
		if err != nil {
			log.Printf("Failed to update rate limit: %v", err)
		}

		c.Next()
	}
}
