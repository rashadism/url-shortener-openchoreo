package services

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"api-service/database"
	"api-service/models"
)

func FetchAndUpdateMetadata(urlID int, longURL string) {
	metadata := fetchLinkMetadata(longURL)

	if metadata.Title != "" {
		err := database.UpdateMetadata(urlID, metadata.Title, metadata.Description)
		if err != nil {
			log.Printf("Failed to update metadata: %v", err)
		}
	}
}

func fetchLinkMetadata(url string) models.LinkMetadata {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to fetch URL metadata: %v", err)
		return models.LinkMetadata{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.LinkMetadata{}
	}

	html := string(body)
	title := extractTitle(html)

	return models.LinkMetadata{
		Title:       title,
		Description: "",
	}
}

func extractTitle(html string) string {
	start := strings.Index(strings.ToLower(html), "<title>")
	if start == -1 {
		return "Untitled"
	}
	start += 7
	end := strings.Index(strings.ToLower(html[start:]), "</title>")
	if end == -1 {
		return "Untitled"
	}
	return html[start : start+end]
}
