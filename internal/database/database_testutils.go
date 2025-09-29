package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/drerr0r/vetbot/internal/models"
)

// TestConfig содержит конфигурацию для тестов базы данных
type TestConfig struct {
	DatabaseURL string
	UseTestDB   bool
}

// GetTestConfig возвращает конфигурацию для тестов
func GetTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: getTestDatabaseURL(),
		UseTestDB:   os.Getenv("USE_TEST_DB") == "true",
	}
}

func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
}

// SetupTestDatabase создает тестовую базу данных
func SetupTestDatabase(t testingT, config *TestConfig) *Database {
	if !config.UseTestDB {
		t.Skip("Skipping database test - set USE_TEST_DB=true to enable")
		return nil
	}

	db, err := New(config.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
		return nil
	}

	// Создаем тестовые таблицы если их нет
	if err := createTestTables(db.GetDB()); err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	return db
}

// testingT это минимальный интерфейс для testing.T
type testingT interface {
	Skip(args ...interface{})
	Skipf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// createTestTables создает тестовые таблицы
func createTestTables(db *sql.DB) error {
	queries := []string{
		// Таблица пользователей
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			phone TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Таблица специализаций
		`CREATE TABLE IF NOT EXISTS specializations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Таблица врачей
		`CREATE TABLE IF NOT EXISTS veterinarians (
			id SERIAL PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			phone TEXT NOT NULL,
			email TEXT,
			description TEXT,
			experience_years INTEGER,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Таблица клиник
		`CREATE TABLE IF NOT EXISTS clinics (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			address TEXT NOT NULL,
			phone TEXT,
			working_hours TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Таблица связей врачей и специализаций
		`CREATE TABLE IF NOT EXISTS vet_specializations (
			vet_id INTEGER REFERENCES veterinarians(id),
			specialization_id INTEGER REFERENCES specializations(id),
			PRIMARY KEY (vet_id, specialization_id)
		)`,

		// Таблица расписания
		`CREATE TABLE IF NOT EXISTS schedules (
			id SERIAL PRIMARY KEY,
			vet_id INTEGER REFERENCES veterinarians(id),
			clinic_id INTEGER REFERENCES clinics(id),
			day_of_week INTEGER NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			is_available BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Вставляем тестовые данные
		`INSERT INTO specializations (id, name, description) VALUES 
			(1, 'Хирург', 'Ветеринарный хирург'),
			(2, 'Терапевт', 'Ветеринарный терапевт'),
			(3, 'Дерматолог', 'Ветеринарный дерматолог')
		ON CONFLICT (id) DO NOTHING`,

		`INSERT INTO veterinarians (id, first_name, last_name, phone, is_active) VALUES 
			(1, 'Иван', 'Петров', '+79123456789', true),
			(2, 'Анна', 'Смирнова', '+79123456780', true)
		ON CONFLICT (id) DO NOTHING`,

		`INSERT INTO clinics (id, name, address, is_active) VALUES 
			(1, 'ВетКлиника Центр', 'ул. Центральная, 1', true),
			(2, 'ВетКлиника Север', 'ул. Северная, 2', true)
		ON CONFLICT (id) DO NOTHING`,

		`INSERT INTO vet_specializations (vet_id, specialization_id) VALUES 
			(1, 1), (1, 2), (2, 3)
		ON CONFLICT DO NOTHING`,

		`INSERT INTO schedules (vet_id, clinic_id, day_of_week, start_time, end_time) VALUES 
			(1, 1, 1, '09:00', '18:00'),
			(1, 1, 3, '09:00', '18:00'),
			(2, 2, 2, '10:00', '19:00')
		ON CONFLICT DO NOTHING`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v\nQuery: %s", err, query)
		}
	}

	return nil
}

// CleanupTestDatabase очищает тестовые данные
func CleanupTestDatabase(db *Database) {
	if db == nil {
		return
	}

	queries := []string{
		"DELETE FROM schedules",
		"DELETE FROM vet_specializations",
		"DELETE FROM veterinarians",
		"DELETE FROM clinics",
		"DELETE FROM specializations",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := db.GetDB().Exec(query); err != nil {
			log.Printf("Warning: failed to cleanup: %v", err)
		}
	}
}

// CreateTestUser создает тестового пользователя
func CreateTestUser(db *Database, telegramID int64) (*models.User, error) {
	user := &models.User{
		TelegramID: telegramID,
		Username:   fmt.Sprintf("testuser_%d", telegramID),
		FirstName:  "Test",
		LastName:   "User",
		Phone:      "+79123456789",
	}

	err := db.CreateUser(user)
	return user, err
}
