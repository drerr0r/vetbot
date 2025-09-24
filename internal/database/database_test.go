package database

import (
	"os"
	"testing"

	"github.com/drerr0r/vetbot/internal/models"
	_ "github.com/lib/pq"
)

// TestDatabaseConnection тестирует подключение к БД
func TestDatabaseConnection(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	// Проверяем, что подключение работает
	err = db.GetDB().Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

// TestCreateUser тестирует создание пользователя
func TestCreateUser(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	// Создаем тестового пользователя
	user := &models.User{
		TelegramID: 123456789,
		Username:   "test_user",
		FirstName:  "Test",
		LastName:   "User",
		Phone:      "+79161234567",
	}

	err = db.CreateUser(user)
	if err != nil {
		t.Errorf("CreateUser failed: %v", err)
	}

	// Проверяем, что ID и CreatedAt установлены
	if user.ID == 0 {
		t.Error("User ID should be set after creation")
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after creation")
	}

	// Тестируем upsert (обновление существующего пользователя)
	user.Username = "updated_user"
	err = db.CreateUser(user)
	if err != nil {
		t.Errorf("CreateUser upsert failed: %v", err)
	}
}

// TestGetAllSpecializations тестирует получение специализаций
func TestGetAllSpecializations(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	specializations, err := db.GetAllSpecializations()
	if err != nil {
		t.Errorf("GetAllSpecializations failed: %v", err)
	}

	// Проверяем, что получили хотя бы пустой слайс
	if specializations == nil {
		t.Error("GetAllSpecializations should return empty slice, not nil")
	}

	// Если есть данные, проверяем структуру
	if len(specializations) > 0 {
		spec := specializations[0]
		if spec.Name == "" {
			t.Error("Specialization name should not be empty")
		}
	}
}

// TestSpecializationExists тестирует проверку существования специализации
func TestSpecializationExists(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	// Тестируем несуществующую специализацию
	exists, err := db.SpecializationExists(99999)
	if err != nil {
		t.Errorf("SpecializationExists failed: %v", err)
	}
	if exists {
		t.Error("Specialization 99999 should not exist")
	}

	// Тестируем существующую специализацию (если есть данные)
	specializations, _ := db.GetAllSpecializations()
	if len(specializations) > 0 {
		exists, err = db.SpecializationExists(specializations[0].ID)
		if err != nil {
			t.Errorf("SpecializationExists failed for existing ID: %v", err)
		}
		if !exists {
			t.Error("Existing specialization should return true")
		}
	}
}

// TestGetAllClinics тестирует получение клиник
func TestGetAllClinics(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	clinics, err := db.GetAllClinics()
	if err != nil {
		t.Errorf("GetAllClinics failed: %v", err)
	}

	if clinics == nil {
		t.Error("GetAllClinics should return empty slice, not nil")
	}

	if len(clinics) > 0 {
		clinic := clinics[0]
		if clinic.Name == "" {
			t.Error("Clinic name should not be empty")
		}
	}
}

// TestFindAvailableVets тестирует поиск врачей
func TestFindAvailableVets(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	// Тест с пустыми критериями
	criteria := &models.SearchCriteria{}
	vets, err := db.FindAvailableVets(criteria)
	if err != nil {
		t.Errorf("FindAvailableVets failed: %v", err)
	}

	if vets == nil {
		t.Error("FindAvailableVets should return empty slice, not nil")
	}

	// Тест с критериями (если есть данные в БД)
	if len(vets) > 0 {
		criteria := &models.SearchCriteria{
			SpecializationID: 1, // предполагая, что ID 1 существует
		}
		_, err := db.FindAvailableVets(criteria)
		if err != nil {
			t.Errorf("FindAvailableVets with criteria failed: %v", err)
		}
	}
}

// TestAddMissingColumns тестирует добавление недостающих колонок
func TestAddMissingColumns(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vetbot_user:vetbot_password@localhost:5432/vetbot_test?sslmode=disable"
	}

	db, err := New(dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	err = db.AddMissingColumns()
	if err != nil {
		t.Errorf("AddMissingColumns failed: %v", err)
	}

	// Проверяем, что колонки существуют
	var columnExists bool
	err = db.GetDB().QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'clinics' AND column_name = 'is_active'
		)
	`).Scan(&columnExists)
	if err != nil {
		t.Errorf("Check column exists failed: %v", err)
	}
	if !columnExists {
		t.Error("is_active column should exist in clinics table")
	}
}
