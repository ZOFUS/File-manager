package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"secure-fm/config"

	_ "github.com/lib/pq"
)

// DB — глобальное подключение к базе данных
var DB *sql.DB

// InitDB инициализирует подключение к PostgreSQL
func InitDB(cfg *config.Config) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	var err error
	// Повторные попытки подключения (ожидание запуска БД)
	for i := 0; i < 10; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err == nil {
			err = DB.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Не удалось подключиться к БД, повтор через 2с... (%v)", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	log.Println("Успешное подключение к базе данных")
	createTables()
}

// createTables создаёт необходимые таблицы, если они не существуют
func createTables() {
	queries := []string{
		// Таблица пользователей
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		);`,
		// Таблица метаданных файлов
		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			size BIGINT,
			location TEXT,
			owner_id INT REFERENCES users(id)
		);`,
		// Таблица логов операций
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
			log.Fatalf("Ошибка создания таблицы: %v", err)
		}
	}
	log.Println("Таблицы успешно созданы")
}
