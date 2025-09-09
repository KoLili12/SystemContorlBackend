package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig загружает переменные окружения из .env файла
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

// GetEnv получает переменную окружения или возвращает значение по умолчанию
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
