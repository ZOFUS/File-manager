package config

import (
	"os"
)

// Config содержит настройки приложения
type Config struct {
	DBHost      string // Хост базы данных
	DBPort      string // Порт базы данных
	DBUser      string // Имя пользователя БД
	DBPassword  string // Пароль БД
	DBName      string // Имя базы данных
	SandboxPath string // Путь к изолированной папке sandbox
}

// LoadConfig загружает конфигурацию из переменных окружения
// Если переменная не задана, используется значение по умолчанию
func LoadConfig() *Config {
	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "secret"),
		DBName:      getEnv("DB_NAME", "securefm"),
		SandboxPath: getEnv("SANDBOX_PATH", "./sandbox"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
