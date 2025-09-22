package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config структура для хранения всех конфигурационных параметров приложения
type Config struct {
	BotToken      string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	AdminUsername string
	AdminChatID   int64
	LogLevel      string
	AppPort       int
	AppEnv        string
}

// LoadConfig загружает конфигурацию из переменных окружения и .env файла
func LoadConfig() *Config {
	// Загружаем .env файл если он существует (для разработки)
	// В продакшене используем только переменные окружения
	godotenv.Load(".env")

	return &Config{
		BotToken:      getEnv("TELEGRAM_BOT_TOKEN", "7447450525:AAE4kX8KsYXG3ieGJCbrk6cMgypuoTSlybs"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnvAsInt("DB_PORT", 5432),
		DBUser:        getEnv("DB_USER", "vetbot_user"),
		DBPassword:    getEnv("DB_PASSWORD", "secure_password_123"),
		DBName:        getEnv("DB_NAME", "vetbot_db"),
		AdminUsername: getEnv("ADMIN_USERNAME", "drerr0r"),
		AdminChatID:   getEnvAsInt64("ADMIN_CHAT_ID", 0),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		AppPort:       getEnvAsInt("APP_PORT", 8080),
		AppEnv:        getEnv("APP_ENV", "development"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчание
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как integer или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Environment variable %s is not a valid integer, using default: %d", key, defaultValue)
	}
	return defaultValue
}

// getEnvAsInt64 получает значение переменной окружения как int64 или возвращает значение по умолчанию
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: Environment variable %s is not a valid int64, using default: %d", key, defaultValue)
	}
	return defaultValue
}

// GetDBConnectionString возвращает строку подключения к PostgreSQL
func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// Validate проверяет обязательные параметры конфигурации
func (c *Config) Validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.DBUser == "" || c.DBPassword == "" || c.DBName == "" {
		return fmt.Errorf("database configuration is incomplete")
	}
	return nil
}

// IsProduction проверяет, работает ли приложение в production режиме
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsDevelopment проверяет, работает ли приложение в development режиме
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}
