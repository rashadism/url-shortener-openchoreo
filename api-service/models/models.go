package models

import "time"

type CreateURLRequest struct {
	LongURL    string `json:"long_url" binding:"required"`
	CustomCode string `json:"custom_code"`
	APIKey     string `json:"api_key" binding:"required"`
}

type URLResponse struct {
	ID        int       `json:"id"`
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	ShortURL  string    `json:"short_url"`
	Title     string    `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type LinkMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
