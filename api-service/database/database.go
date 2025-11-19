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

	// Initialize schema
	if err = initSchema(); err != nil {
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

func initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			api_key VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			long_url TEXT NOT NULL,
			title VARCHAR(500),
			description TEXT,
			user_id INTEGER REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE
		);

		CREATE TABLE IF NOT EXISTS clicks (
			id SERIAL PRIMARY KEY,
			url_id INTEGER REFERENCES urls(id) ON DELETE CASCADE,
			ip_address VARCHAR(45),
			user_agent TEXT,
			referer TEXT,
			country VARCHAR(100),
			city VARCHAR(100),
			clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
		CREATE INDEX IF NOT EXISTS idx_clicks_url_id ON clicks(url_id);
		CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks(clicked_at);
		CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

		INSERT INTO users (username, api_key)
		VALUES ('testuser', 'test-api-key-12345')
		ON CONFLICT (username) DO NOTHING;
	`

	_, err := DB.Exec(schema)
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func ValidateAPIKey(apiKey string) (int, error) {
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE api_key = $1", apiKey).Scan(&userID)
	return userID, err
}

func InsertURL(shortCode, longURL string, userID int) (int, error) {
	var urlID int
	err := DB.QueryRow(
		"INSERT INTO urls (short_code, long_url, user_id) VALUES ($1, $2, $3) RETURNING id",
		shortCode, longURL, userID,
	).Scan(&urlID)
	return urlID, err
}

func GetURLByShortCode(shortCode string) (int, string, error) {
	var urlID int
	var longURL string
	err := DB.QueryRow(
		"SELECT id, long_url FROM urls WHERE short_code = $1 AND is_active = true",
		shortCode,
	).Scan(&urlID, &longURL)
	return urlID, longURL, err
}

func GetURLIDByShortCode(shortCode string) (int, error) {
	var urlID int
	err := DB.QueryRow("SELECT id FROM urls WHERE short_code = $1", shortCode).Scan(&urlID)
	return urlID, err
}

func RecordClick(urlID int, ip, userAgent, referer string) error {
	_, err := DB.Exec(
		"INSERT INTO clicks (url_id, ip_address, user_agent, referer) VALUES ($1, $2, $3, $4)",
		urlID, ip, userAgent, referer,
	)
	return err
}

func UpdateMetadata(urlID int, title, description string) error {
	_, err := DB.Exec(
		"UPDATE urls SET title = $1, description = $2 WHERE id = $3",
		title, description, urlID,
	)
	return err
}

func CheckHealth() bool {
	return DB.Ping() == nil
}
