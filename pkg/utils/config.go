package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config содержит все конфигурационные параметры приложения
type Config struct {
	TelegramToken string
	DatabaseURL   string
	Debug         bool
	AdminIDs      []int64
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Токен Telegram бота (обязательный параметр)
	config.TelegramToken = getEnv("TELEGRAM_TOKEN", "")
	if config.TelegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	// Безопасное логирование токена
	tokenPreview := config.TelegramToken
	if len(tokenPreview) > 10 {
		tokenPreview = tokenPreview[:10] + "..."
	}
	log.Printf("Telegram token loaded: %s", tokenPreview)

	// URL базы данных (обязательный параметр)
	config.DatabaseURL = getEnv("DATABASE_URL", "")
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	log.Printf("Database URL loaded: %s", config.DatabaseURL)

	// Режим отладки (опционально)
	debugStr := getEnv("DEBUG", "false")
	config.Debug, _ = strconv.ParseBool(debugStr)
	log.Printf("Debug mode: %t", config.Debug)

	// ID администраторов (опционально)
	adminIDsStr := getEnv("ADMIN_IDS", "")
	log.Printf("Raw ADMIN_IDS from env: '%s'", adminIDsStr)

	if adminIDsStr != "" {
		config.AdminIDs = parseAdminIDs(adminIDsStr)
		log.Printf("Loaded admin IDs: %v", config.AdminIDs)
	} else {
		log.Printf("No admin IDs configured")
	}

	return config, nil
}

// parseAdminIDs парсит строку с ID администраторов
func parseAdminIDs(adminIDsStr string) []int64 {
	var adminIDs []int64
	ids := strings.Split(adminIDsStr, ",")

	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing admin ID %s: %v", idStr, err)
			continue
		}
		adminIDs = append(adminIDs, id)
	}

	return adminIDs
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
