package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"api-service/cache"
	"api-service/database"
	"api-service/models"
	"api-service/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var cacheTTL time.Duration

// Init initializes the handlers package with configuration
func Init(ttlSeconds int) {
	cacheTTL = time.Duration(ttlSeconds) * time.Second
}

func CreateURL(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get or create user by username
	userID, err := database.GetOrCreateUser(req.Username)
	if err != nil {
		log.Printf("Database unavailable for user creation: %v", err)
		userID = 0 // Continue without user ID
	}

	// Generate short code
	shortCode := req.CustomCode
	if shortCode == "" {
		shortCode = services.GenerateShortCode()
	}

	// Check if short code already exists in cache
	existingURL, err := cache.GetURL(shortCode)
	if err == nil && existingURL != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "Short code already exists"})
		return
	}

	// Cache the URL (required for redirects to work)
	err = cache.SetURL(shortCode, req.LongURL, cacheTTL)
	if err != nil {
		log.Printf("Failed to cache URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
		return
	}

	// Try to insert into database (best effort)
	urlID := 0
	urlID, err = database.InsertURL(shortCode, req.LongURL, userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			cache.DeleteURL(shortCode)
			c.JSON(http.StatusConflict, gin.H{"error": "Short code already exists"})
			return
		}
		log.Printf("Database unavailable, URL cached for %v: %v", cacheTTL, err)
		// Continue - URL is cached and will work for redirects
	}

	// Fetch metadata asynchronously (only if DB insert succeeded)
	if urlID > 0 {
		go services.FetchAndUpdateMetadata(urlID, req.LongURL)
	}

	response := models.URLResponse{
		ID:        urlID,
		ShortCode: shortCode,
		LongURL:   req.LongURL,
		ShortURL:  fmt.Sprintf("/%s", shortCode),
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, response)
}

func ListURLs(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username parameter is required"})
		return
	}

	userID, err := database.GetOrCreateUser(username)
	if err != nil {
		log.Printf("Database unavailable for listing URLs: %v", err)
		// Return empty list when DB is down
		c.JSON(http.StatusOK, []models.URLResponse{})
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, short_code, long_url, title, created_at FROM urls WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100",
		userID,
	)
	if err != nil {
		log.Printf("Database unavailable for query: %v", err)
		// Return empty list when DB is down
		c.JSON(http.StatusOK, []models.URLResponse{})
		return
	}
	defer rows.Close()

	var urls []models.URLResponse
	for rows.Next() {
		var url models.URLResponse
		var title sql.NullString
		err := rows.Scan(&url.ID, &url.ShortCode, &url.LongURL, &title, &url.CreatedAt)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		if title.Valid {
			url.Title = title.String
		}
		url.ShortURL = fmt.Sprintf("/%s", url.ShortCode)
		urls = append(urls, url)
	}

	c.JSON(http.StatusOK, urls)
}

func Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	// Try Redis cache first
	longURL, err := cache.GetURL(shortCode)

	var urlID int

	if err == redis.Nil {
		// Cache miss - query database
		log.Printf("Cache miss for %s, querying database", shortCode)
		urlID, longURL, err = database.GetURLByShortCode(shortCode)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		if err != nil {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URL"})
			return
		}

		// Update cache
		cache.SetURL(shortCode, longURL, cacheTTL)
	} else if err != nil {
		log.Printf("Redis error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cache error"})
		return
	} else {
		// Cache hit - still need URL ID for click tracking
		urlID, err = database.GetURLIDByShortCode(shortCode)
		if err != nil {
			log.Printf("Failed to get URL ID for click tracking: %v", err)
		}
	}

	// Extract request data before spawning goroutine
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Record click asynchronously (only if we have a valid URL ID)
	if urlID > 0 {
		go recordClickAsync(urlID, ip, userAgent, referer)
	}

	// Increment Redis counter
	cache.IncrementClickCounter(shortCode)

	c.Redirect(http.StatusFound, longURL)
}

func recordClickAsync(urlID int, ip string, userAgent string, referer string) {
	err := database.RecordClick(urlID, ip, userAgent, referer)
	if err != nil {
		log.Printf("Failed to record click for URL ID %d: %v", urlID, err)
	} else {
		log.Printf("Click recorded successfully for URL ID %d", urlID)
	}
}
