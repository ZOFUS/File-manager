package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"secure-fm/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(cfg *config.Config) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	var err error
	// Retry logic for waiting DB to start
	for i := 0; i < 10; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err == nil {
			err = DB.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to DB, retrying in 2s... (%v)", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	log.Println("Connected to database successfully")
	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			size BIGINT,
			location TEXT,
			owner_id INT REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS operations (
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			operation_type VARCHAR(50),
			file_id INT REFERENCES files(id),
			user_id INT REFERENCES users(id)
		);`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
	log.Println("Tables created successfully")
}
