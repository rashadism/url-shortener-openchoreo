package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(dbURL string) error {
	var err error
	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connected successfully")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func GetOrCreateUser(username string) (int, error) {
	if DB == nil {
		return 0, sql.ErrConnDone
	}

	var userID int

	// Try to get existing user
	err := DB.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err == nil {
		return userID, nil
	}

	// User doesn't exist, create new user
	err = DB.QueryRow(
		"INSERT INTO users (username, api_key) VALUES ($1, $2) RETURNING id",
		username, "", // Empty api_key since we're not using it
	).Scan(&userID)

	return userID, err
}

func InsertURL(shortCode, longURL string, userID int) (int, error) {
	if DB == nil {
		return 0, sql.ErrConnDone
	}

	var urlID int
	err := DB.QueryRow(
		"INSERT INTO urls (short_code, long_url, user_id) VALUES ($1, $2, $3) RETURNING id",
		shortCode, longURL, userID,
	).Scan(&urlID)
	return urlID, err
}

func GetURLByShortCode(shortCode string) (int, string, error) {
	if DB == nil {
		return 0, "", sql.ErrConnDone
	}

	var urlID int
	var longURL string
	err := DB.QueryRow(
		"SELECT id, long_url FROM urls WHERE short_code = $1 AND is_active = true",
		shortCode,
	).Scan(&urlID, &longURL)
	return urlID, longURL, err
}

func GetURLIDByShortCode(shortCode string) (int, error) {
	if DB == nil {
		return 0, sql.ErrConnDone
	}

	var urlID int
	err := DB.QueryRow("SELECT id FROM urls WHERE short_code = $1", shortCode).Scan(&urlID)
	return urlID, err
}

func RecordClick(urlID int, ip, userAgent, referer string) error {
	if DB == nil {
		return sql.ErrConnDone
	}

	_, err := DB.Exec(
		"INSERT INTO clicks (url_id, ip_address, user_agent, referer) VALUES ($1, $2, $3, $4)",
		urlID, ip, userAgent, referer,
	)
	return err
}

func UpdateMetadata(urlID int, title, description string) error {
	if DB == nil {
		return sql.ErrConnDone
	}

	_, err := DB.Exec(
		"UPDATE urls SET title = $1, description = $2 WHERE id = $3",
		title, description, urlID,
	)
	return err
}

func CheckHealth() bool {
	if DB == nil {
		return false
	}
	return DB.Ping() == nil
}
