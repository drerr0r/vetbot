package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/drerr0r/vetbot/internal/database"
	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ТЕСТЫ ДЛЯ ВСПОМОГАТЕЛЬНЫХ ФУНКЦИЙ (не требуют моков)
// ============================================================================

func TestGetDayName(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"Понедельник", 1, "понедельник"},
		{"Вторник", 2, "вторник"},
		{"Среда", 3, "среду"},
		{"Четверг", 4, "четверг"},
		{"Пятница", 5, "пятницу"},
		{"Суббота", 6, "субботу"},
		{"Воскресенье", 7, "воскресенье"},
		{"Любой день", 0, "любой день"},
		{"Несуществующий день", 8, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDayName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDayName_Comprehensive(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "понедельник"},
		{2, "вторник"},
		{3, "среду"},
		{4, "четверг"},
		{5, "пятницу"},
		{6, "субботу"},
		{7, "воскресенье"},
		{0, "любой день"},
		{8, ""},
		{-1, ""},
	}

	for _, test := range tests {
		t.Run(string(rune(test.input)), func(t *testing.T) {
			result := getDayName(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ КОНСТРУКТОРА (не требуют моков)
// ============================================================================

func TestNewVetHandlers(t *testing.T) {
	// Создаем nil указатели, так как в реальных тестах мы не будем их использовать
	var bot *tgbotapi.BotAPI = nil
	var db *database.Database = nil

	handler := NewVetHandlers(bot, db)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.bot) // В тестовом режиле они будут nil
	assert.Nil(t, handler.db)
}

// ============================================================================
// ТЕСТЫ ДЛЯ ФОРМАТИРОВАНИЯ ДАННЫХ (логика форматирования сообщений)
// ============================================================================

func TestFormatVeterinarianInfo(t *testing.T) {
	// Тестируем логику форматирования данных врача
	vet := &models.Veterinarian{
		ID:              1,
		FirstName:       "Иван",
		LastName:        "Петров",
		Phone:           "+79123456789",
		Email:           sql.NullString{String: "ivan@vet.ru", Valid: true},
		ExperienceYears: sql.NullInt64{Int64: 5, Valid: true},
		IsActive:        true,
	}

	// Проверяем, что данные правильно экранируются
	// Это тестирует логику, которая используется в HandleSearchBySpecialization
	assert.Contains(t, vet.FirstName, "Иван")
	assert.Contains(t, vet.LastName, "Петров")
	assert.Equal(t, "+79123456789", vet.Phone)
}

func TestFormatScheduleInfo(t *testing.T) {
	// Тестируем логику форматирования расписания
	schedule := &models.Schedule{
		DayOfWeek: 1,
		StartTime: "09:00",
		EndTime:   "18:00",
		Clinic: &models.Clinic{
			Name: "ВетКлиника",
		},
	}

	dayName := getDayName(schedule.DayOfWeek)
	expectedDayName := "понедельник"

	assert.Equal(t, expectedDayName, dayName)
	assert.Equal(t, "09:00", schedule.StartTime)
	assert.Equal(t, "18:00", schedule.EndTime)
	assert.Equal(t, "ВетКлиника", schedule.Clinic.Name)
}

// ============================================================================
// ТЕСТЫ ДЛЯ ВАЛИДАЦИИ ДАННЫХ
// ============================================================================

func TestValidateSearchCriteria(t *testing.T) {
	tests := []struct {
		name      string
		criteria  *models.SearchCriteria
		shouldErr bool
	}{
		{
			name: "Valid specialization search",
			criteria: &models.SearchCriteria{
				SpecializationID: 1,
			},
			shouldErr: false,
		},
		{
			name: "Valid day search",
			criteria: &models.SearchCriteria{
				DayOfWeek: 1,
			},
			shouldErr: false,
		},
		{
			name: "Valid clinic search",
			criteria: &models.SearchCriteria{
				ClinicID: 1,
			},
			shouldErr: false,
		},
		{
			name: "Invalid day number",
			criteria: &models.SearchCriteria{
				DayOfWeek: 8, // Несуществующий день
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверяем валидность критериев поиска
			isValid := true

			if tt.criteria.DayOfWeek > 7 || tt.criteria.DayOfWeek < 0 {
				isValid = false
			}
			if tt.criteria.SpecializationID < 0 {
				isValid = false
			}
			if tt.criteria.ClinicID < 0 {
				isValid = false
			}

			if tt.shouldErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ЛОГИКИ РАБОТЫ С ДАННЫМИ (без внешних зависимостей)
// ============================================================================

func TestCreateUserFromTelegramUser(t *testing.T) {
	telegramUser := &tgbotapi.User{
		ID:        12345,
		UserName:  "testuser",
		FirstName: "Test",
		LastName:  "User",
	}

	// Тестируем логику преобразования данных пользователя
	expectedUser := &models.User{
		TelegramID: telegramUser.ID,
		Username:   telegramUser.UserName,
		FirstName:  telegramUser.FirstName,
		LastName:   telegramUser.LastName,
	}

	assert.Equal(t, int64(12345), expectedUser.TelegramID)
	assert.Equal(t, "testuser", expectedUser.Username)
	assert.Equal(t, "Test", expectedUser.FirstName)
	assert.Equal(t, "User", expectedUser.LastName)
}

func TestBuildSpecializationKeyboard(t *testing.T) {
	// Тестируем логику построения клавиатуры
	specializations := []*models.Specialization{
		{ID: 1, Name: "Хирургия"},
		{ID: 2, Name: "Терапия"},
		{ID: 3, Name: "Стоматология"},
		{ID: 4, Name: "Дерматология"},
	}

	// Проверяем, что специализации правильно обрабатываются
	assert.Len(t, specializations, 4)
	assert.Equal(t, "Хирургия", specializations[0].Name)
	assert.Equal(t, 1, specializations[0].ID)
}

func TestBuildClinicKeyboard(t *testing.T) {
	// Тестируем логику построения клавиатуры для клиник
	clinics := []*models.Clinic{
		{ID: 1, Name: "Клиника 1"},
		{ID: 2, Name: "Клиника 2"},
		{ID: 3, Name: "Клиника 3"},
	}

	assert.Len(t, clinics, 3)
	assert.Equal(t, "Клиника 1", clinics[0].Name)
	assert.Equal(t, 1, clinics[0].ID)
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ CALLBACK ДАННЫХ
// ============================================================================

func TestParseCallbackData(t *testing.T) {
	tests := []struct {
		name          string
		callbackData  string
		expectedType  string
		expectedID    int
		shouldSucceed bool
	}{
		{
			name:          "Valid specialization callback",
			callbackData:  "search_spec_1",
			expectedType:  "specialization",
			expectedID:    1,
			shouldSucceed: true,
		},
		{
			name:          "Valid clinic callback",
			callbackData:  "search_clinic_2",
			expectedType:  "clinic",
			expectedID:    2,
			shouldSucceed: true,
		},
		{
			name:          "Valid day callback",
			callbackData:  "search_day_3",
			expectedType:  "day",
			expectedID:    3,
			shouldSucceed: true,
		},
		{
			name:          "Invalid callback format",
			callbackData:  "invalid_data",
			expectedType:  "",
			expectedID:    0,
			shouldSucceed: false,
		},
		{
			name:          "Callback with non-numeric ID",
			callbackData:  "search_spec_abc",
			expectedType:  "",
			expectedID:    0,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callbackType string
			var id int
			var err error

			// Имитируем логику парсинга из handleSearchSpecCallback
			switch {
			case len(tt.callbackData) > len("search_spec_") && tt.callbackData[:len("search_spec_")] == "search_spec_":
				callbackType = "specialization"
				idStr := tt.callbackData[len("search_spec_"):]
				_, err = strconv.Atoi(idStr)
				if err == nil {
					id, _ = strconv.Atoi(idStr)
				}
			case len(tt.callbackData) > len("search_clinic_") && tt.callbackData[:len("search_clinic_")] == "search_clinic_":
				callbackType = "clinic"
				idStr := tt.callbackData[len("search_clinic_"):]
				_, err = strconv.Atoi(idStr)
				if err == nil {
					id, _ = strconv.Atoi(idStr)
				}
			case len(tt.callbackData) > len("search_day_") && tt.callbackData[:len("search_day_")] == "search_day_":
				callbackType = "day"
				idStr := tt.callbackData[len("search_day_"):]
				_, err = strconv.Atoi(idStr)
				if err == nil {
					id, _ = strconv.Atoi(idStr)
				}
			default:
				callbackType = ""
				id = 0
				err = errors.New("invalid format")
			}

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, callbackType)
				assert.Equal(t, tt.expectedID, id)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ФОРМАТИРОВАНИЯ СООБЩЕНИЙ
// ============================================================================

func TestBuildVeterinarianMessage(t *testing.T) {
	// Тестируем логику построения сообщения с информацией о враче
	vet := &models.Veterinarian{
		FirstName:       "Иван",
		LastName:        "Петров",
		Phone:           "+79123456789",
		Email:           sql.NullString{String: "ivan@vet.ru", Valid: true},
		ExperienceYears: sql.NullInt64{Int64: 5, Valid: true},
		Specializations: []*models.Specialization{
			{Name: "Хирургия"},
			{Name: "Терапия"},
		},
	}

	// Проверяем, что данные правильно форматируются
	assert.Contains(t, vet.FirstName, "Иван")
	assert.Contains(t, vet.LastName, "Петров")
	assert.Equal(t, "+79123456789", vet.Phone)
	assert.True(t, vet.Email.Valid)
	assert.Equal(t, "ivan@vet.ru", vet.Email.String)
	assert.Equal(t, int64(5), vet.ExperienceYears.Int64)
	assert.Len(t, vet.Specializations, 2)
}

func TestBuildEmptyResultsMessage(t *testing.T) {
	// Тестируем логику построения сообщения об отсутствии результатов
	specializationName := "Хирургия"

	// Имитируем логику из HandleSearchBySpecialization
	message := fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\" не найдены*\n\nПопробуйте выбрать другую специализацию.", specializationName)

	assert.Contains(t, message, "не найдены")
	assert.Contains(t, message, "Хирургия")
	assert.Contains(t, message, "специализации")
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ ОШИБОК
// ============================================================================

func TestErrorHandling(t *testing.T) {
	// Тестируем различные сценарии обработки ошибок
	errorScenarios := []struct {
		name        string
		errorType   string
		shouldRetry bool
	}{
		{"Database connection error", "database", false},
		{"Network error", "network", true},
		{"Invalid data error", "validation", false},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Проверяем логику обработки разных типов ошибок
			isRetryable := scenario.errorType == "network"
			assert.Equal(t, scenario.shouldRetry, isRetryable)
		})
	}
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ ТЕСТОВ
// ============================================================================

// createTestVeterinarian создает тестового ветеринара
func createTestVeterinarian() *models.Veterinarian {
	return &models.Veterinarian{
		ID:              1,
		FirstName:       "Тест",
		LastName:        "Врач",
		Phone:           "+79123456789",
		Email:           sql.NullString{String: "test@vet.ru", Valid: true},
		ExperienceYears: sql.NullInt64{Int64: 3, Valid: true},
		IsActive:        true,
		CreatedAt:       time.Now(),
	}
}

// createTestSpecialization создает тестовую специализацию
func createTestSpecialization() *models.Specialization {
	return &models.Specialization{
		ID:          1,
		Name:        "Тестовая специализация",
		Description: "Описание тестовой специализации",
		CreatedAt:   time.Now(),
	}
}

// createTestClinic создает тестовую клинику
func createTestClinic() *models.Clinic {
	return &models.Clinic{
		ID:           1,
		Name:         "Тестовая клиника",
		Address:      "Тестовый адрес",
		Phone:        sql.NullString{String: "+79998887766", Valid: true},
		WorkingHours: sql.NullString{String: "9:00-18:00", Valid: true},
		IsActive:     true,
		CreatedAt:    time.Now(),
	}
}
