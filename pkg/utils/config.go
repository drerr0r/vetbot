package utils

import (
	"log"
	"os"
	"strconv"
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
		log.Fatal("TELEGRAM_TOKEN is required")
	}

	// URL базы данных (обязательный параметр)
	config.DatabaseURL = getEnv("DATABASE_URL", "")
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Режим отладки (опционально)
	debugStr := getEnv("DEBUG", "false")
	config.Debug, _ = strconv.ParseBool(debugStr)

	// ID администраторов (опционально)
	adminIDsStr := getEnv("ADMIN_IDS", "")
	if adminIDsStr != "" {
		// Парсим список ID администраторов (формат: "123,456,789")
		// В реальном приложении нужно добавить парсинг
		config.AdminIDs = []int64{123456789} // Заглушка
	}

	return config, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
