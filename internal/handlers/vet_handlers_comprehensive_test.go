package handlers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ТЕСТЫ ДЛЯ КОНСТРУКТОРА И БАЗОВЫХ МЕТОДОВ
// ============================================================================

func TestNewVetHandlers(t *testing.T) {
	// Arrange
	mockBot := NewMockBot()
	mockDB := NewMockDatabase()
	stateManager := NewTestStateManager()

	// Act
	handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

	// Assert
	assert.NotNil(t, handlers)
	assert.Equal(t, mockBot, handlers.bot)
	assert.Equal(t, mockDB, handlers.db)
}

func TestGetDayName(t *testing.T) {
	tests := []struct {
		name     string
		day      int
		expected string
	}{
		{"Monday", 1, "понедельник"},
		{"Tuesday", 2, "вторник"},
		{"Wednesday", 3, "среду"},
		{"Thursday", 4, "четверг"},
		{"Friday", 5, "пятницу"},
		{"Saturday", 6, "субботу"},
		{"Sunday", 7, "воскресенье"},
		{"Any day", 0, "любой день"},
		{"Invalid day", 8, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDayName(tt.day)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleStart
// ============================================================================

func TestVetHandleStart(t *testing.T) {
	t.Run("Successful start with new user", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithMessage("/start", 12345, 67890).
			Build()

		// Act
		handlers.HandleStart(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.NotNil(t, message)
		assert.Contains(t, message.Text, "Добро пожаловать в VetBot")
		assert.Equal(t, int64(12345), message.ChatID)
		assert.NotNil(t, message.ReplyMarkup)
	})

	t.Run("Start with database error", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		mockDB.UserError = fmt.Errorf("database connection failed")
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithMessage("/start", 12345, 67890).
			Build()

		// Act
		handlers.HandleStart(update)

		// Assert
		// Должен отправить сообщение даже при ошибке БД
		assert.Len(t, mockBot.SentMessages, 1)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleSpecializations
// ============================================================================

func TestVetHandleSpecializations(t *testing.T) {
	t.Run("Successful specializations list", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Добавляем тестовые специализации
		mockDB.Specializations[1] = &models.Specialization{ID: 1, Name: "Хирургия"}
		mockDB.Specializations[2] = &models.Specialization{ID: 2, Name: "Терапия"}
		mockDB.Specializations[3] = &models.Specialization{ID: 3, Name: "Дерматология"}

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_specializations", 12345, 1).
			Build()

		// Act
		handlers.HandleSpecializations(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Выберите специализацию врача")
		assert.NotNil(t, message.ReplyMarkup)
	})

	t.Run("Specializations with database error", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		mockDB.SpecializationsError = fmt.Errorf("database error")
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_specializations", 12345, 1).
			Build()

		// Act
		handlers.HandleSpecializations(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Ошибка при получении списка специализаций")
	})

	t.Run("Specializations empty list", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase() // Пустая база
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_specializations", 12345, 1).
			Build()

		// Act
		handlers.HandleSpecializations(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Специализации не найдены")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleSearchBySpecialization
// ============================================================================

func TestVetHandleSearchBySpecialization(t *testing.T) {
	t.Run("Successful search with results", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Создаем специализацию
		spec := &models.Specialization{ID: 1, Name: "Хирургия"}
		mockDB.Specializations[1] = spec

		// Создаем ветеринара
		vet := &models.Veterinarian{
			ID:              sql.NullInt64{Int64: 1, Valid: true},
			FirstName:       "Иван",
			LastName:        "Петров",
			Phone:           "+79123456789",
			Email:           sql.NullString{String: "ivan@vet.ru", Valid: true},
			ExperienceYears: sql.NullInt64{Int64: 5, Valid: true},
			Specializations: []*models.Specialization{spec},
		}
		mockDB.Veterinarians[1] = vet

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchBySpecialization(update, 1)

		// Assert
		// Проверяем, что сообщение было отправлено или отредактировано
		if len(mockBot.SentMessages) > 0 {
			message := mockBot.GetLastMessage()
			assert.Contains(t, message.Text, "Иван Петров")
			assert.Contains(t, message.Text, "+79123456789")
			assert.Contains(t, message.Text, "Иван Петров")
		} else if len(mockBot.EditedMessages) > 0 {
			edited := mockBot.GetLastEditedMessage()
			assert.Contains(t, edited.Text, "Врачи по специализации")
			assert.Contains(t, edited.Text, "Хирургия")
			assert.Contains(t, edited.Text, "Иван Петров")
		} else {
			t.Error("No messages were sent or edited")
		}
	})

	t.Run("Search with no results", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Специализация есть, но врачей нет
		spec := &models.Specialization{ID: 1, Name: "Хирургия"}
		mockDB.Specializations[1] = spec

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchBySpecialization(update, 1)

		// Assert
		if len(mockBot.SentMessages) > 0 {
			message := mockBot.GetLastMessage()
			assert.Contains(t, message.Text, "не найдены")
		} else if len(mockBot.EditedMessages) > 0 {
			edited := mockBot.GetLastEditedMessage()
			assert.Contains(t, edited.Text, "не найдены")
		}
	})

	t.Run("Search with database error", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		mockDB.VeterinariansError = fmt.Errorf("database error")
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchBySpecialization(update, 1)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Ошибка при поиске врачей")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleClinics
// ============================================================================

func TestVetHandleClinics(t *testing.T) {
	t.Run("Successful clinics list", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Добавляем тестовые клиники
		mockDB.Clinics[1] = &models.Clinic{ID: 1, Name: "ВетКлиника №1"}
		mockDB.Clinics[2] = &models.Clinic{ID: 2, Name: "ВетКлиника №2"}

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_clinics", 12345, 1).
			Build()

		// Act
		handlers.HandleClinics(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Выберите клинику")
		assert.NotNil(t, message.ReplyMarkup)
	})

	t.Run("Clinics with database error", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		mockDB.ClinicsError = fmt.Errorf("database error")
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_clinics", 12345, 1).
			Build()

		// Act
		handlers.HandleClinics(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Ошибка при получении списка клиник")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleSearchByClinic
// ============================================================================

func TestVetHandleSearchByClinic(t *testing.T) {
	t.Run("Successful clinic search", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Создаем клинику
		clinic := &models.Clinic{ID: 1, Name: "ВетКлиника Центр"}
		mockDB.Clinics[1] = clinic

		// Создаем ветеринара
		vet := &models.Veterinarian{
			ID:              sql.NullInt64{Int64: 1, Valid: true},
			FirstName:       "Мария",
			LastName:        "Иванова",
			Phone:           "+79123456780",
			Email:           sql.NullString{String: "maria@vet.ru", Valid: true},
			ExperienceYears: sql.NullInt64{Int64: 7, Valid: true},
		}
		mockDB.Veterinarians[1] = vet

		// Создаем расписание
		schedule := &models.Schedule{
			ID:        1,
			VetID:     1,
			ClinicID:  1,
			DayOfWeek: 1,
			StartTime: "09:00",
			EndTime:   "18:00",
			Clinic:    clinic,
		}
		mockDB.Schedules[1] = schedule

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_clinic_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchByClinic(update, 1)

		// Assert
		// Проверяем отправленные или отредактированные сообщения
		var messageText string
		if len(mockBot.SentMessages) > 0 {
			messageText = mockBot.GetLastMessage().Text
		} else if len(mockBot.EditedMessages) > 0 {
			messageText = mockBot.GetLastEditedMessage().Text
		}

		assert.Contains(t, messageText, "Мария Иванова")
		assert.Contains(t, messageText, "+79123456780")
		assert.Contains(t, messageText, "Мария Иванова")
	})

	t.Run("Clinic search with no results", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Клиника есть, но врачей нет
		mockDB.Clinics[1] = &models.Clinic{ID: 1, Name: "ВетКлиника Центр"}

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_clinic_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchByClinic(update, 1)

		// Assert
		var messageText string
		if len(mockBot.SentMessages) > 0 {
			messageText = mockBot.GetLastMessage().Text
		} else if len(mockBot.EditedMessages) > 0 {
			messageText = mockBot.GetLastEditedMessage().Text
		}

		assert.Contains(t, messageText, "не найдены")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleHelp
// ============================================================================

func TestVetHandleHelp(t *testing.T) {
	t.Run("Help message", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_help", 12345, 1).
			Build()

		// Act
		handlers.HandleHelp(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "VetBot - Помощь")
		assert.Contains(t, message.Text, "Основные функции")
		assert.NotNil(t, message.ReplyMarkup)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleCallback
// ============================================================================

func TestVetHandleCallback(t *testing.T) {
	t.Run("Main menu callback", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_menu", 12345, 1).
			Build()

		// Act
		handlers.HandleCallback(update)

		// Assert
		// Должно отредактировать сообщение для показа главного меню
		assert.True(t, len(mockBot.EditedMessages) > 0 || len(mockBot.SentMessages) > 0)
	})

	t.Run("Search spec callback", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Добавляем тестовые данные
		spec := &models.Specialization{ID: 1, Name: "Хирургия"}
		mockDB.Specializations[1] = spec

		vet := &models.Veterinarian{
			ID:              sql.NullInt64{Int64: 1, Valid: true},
			FirstName:       "Иван",
			LastName:        "Петров",
			Phone:           "+79123456789",
			Specializations: []*models.Specialization{spec},
		}
		mockDB.Veterinarians[1] = vet

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 1).
			Build()

		// Act
		handlers.HandleCallback(update)

		// Assert
		// Должен вызвать поиск по специализации
		var messageText string
		if len(mockBot.SentMessages) > 0 {
			messageText = mockBot.GetLastMessage().Text
		} else if len(mockBot.EditedMessages) > 0 {
			messageText = mockBot.GetLastEditedMessage().Text
		}

		assert.Contains(t, messageText, "Иван Петров")
	})

	t.Run("Unknown callback", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("unknown_callback", 12345, 1).
			Build()

		// Act
		handlers.HandleCallback(update)

		// Assert
		// Должен обработать неизвестный callback без паники
		assert.NotNil(t, mockBot) // Просто проверяем, что нет паники
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ HandleTest
// ============================================================================

func TestVetHandleTest(t *testing.T) {
	t.Run("Test command", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithMessage("/test", 12345, 67890).
			Build()

		// Act
		handlers.HandleTest(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Тестовое сообщение")
		assert.Equal(t, int64(12345), message.ChatID)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ КРАЙНИХ СЛУЧАЕВ И ОШИБОК
// ============================================================================

func TestVetEdgeCases(t *testing.T) {
	t.Run("Nil update handling", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		// Act & Assert - не должно быть паники
		assert.NotPanics(t, func() {
			handlers.HandleSpecializations(tgbotapi.Update{})
		})
	})

	t.Run("Update with both message and callback", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "/start",
				Chat: &tgbotapi.Chat{ID: 12345},
				From: &tgbotapi.User{ID: 67890},
			},
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID:   "test",
				Data: "test_data",
			},
		}

		// Act & Assert - не должно быть паники, должен обработать callback
		assert.NotPanics(t, func() {
			handlers.HandleCallback(update)
		})
	})

	t.Run("Invalid specialization ID", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_invalid", 12345, 1).
			Build()

		// Act
		handlers.HandleCallback(update)

		// Assert - не должно быть паники
		assert.NotNil(t, mockBot)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ ФОРМАТИРОВАНИЯ СООБЩЕНИЙ
// ============================================================================

func TestVetMessageFormatting(t *testing.T) {
	t.Run("Veterinarian info formatting", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Создаем ветеринара с полной информацией
		vet := &models.Veterinarian{
			ID:              sql.NullInt64{Int64: 1, Valid: true},
			FirstName:       "Дмитрий",
			LastName:        "Сидоров",
			Phone:           "+79123456789",
			Email:           sql.NullString{String: "dmitry@vet.ru", Valid: true},
			ExperienceYears: sql.NullInt64{Int64: 10, Valid: true},
			Specializations: []*models.Specialization{
				{ID: 1, Name: "Хирургия"},
				{ID: 2, Name: "Терапия"},
			},
		}
		mockDB.Veterinarians[1] = vet

		// Добавляем расписание
		schedule := &models.Schedule{
			VetID:     1,
			DayOfWeek: 1,
			StartTime: "09:00",
			EndTime:   "18:00",
			Clinic:    &models.Clinic{Name: "ВетКлиника Центр"},
		}
		mockDB.Schedules[1] = schedule

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchBySpecialization(update, 1)

		// Assert
		var messageText string
		if len(mockBot.SentMessages) > 0 {
			messageText = mockBot.GetLastMessage().Text
		} else if len(mockBot.EditedMessages) > 0 {
			messageText = mockBot.GetLastEditedMessage().Text
		}

		// Проверяем форматирование
		assert.Contains(t, messageText, "Дмитрий Сидоров")
		assert.Contains(t, messageText, "+79123456789")
		assert.Contains(t, messageText, "dmitry@vet.ru")
		assert.Contains(t, messageText, "10 лет")
		assert.Contains(t, messageText, "понедельник: 09:00-18:00")
	})
}

// ============================================================================
// ИНТЕГРАЦИОННЫЕ ТЕСТЫ
// ============================================================================

func TestVetIntegrationScenarios(t *testing.T) {
	t.Run("Complete user flow", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Настраиваем тестовые данные
		spec1 := &models.Specialization{ID: 1, Name: "Хирургия"}
		spec2 := &models.Specialization{ID: 2, Name: "Терапия"}
		mockDB.Specializations[1] = spec1
		mockDB.Specializations[2] = spec2

		clinic := &models.Clinic{ID: 1, Name: "ВетКлиника Центр"}
		mockDB.Clinics[1] = clinic

		vet := &models.Veterinarian{
			ID:              sql.NullInt64{Int64: 1, Valid: true},
			FirstName:       "Анна",
			LastName:        "Смирнова",
			Phone:           "+79123456789",
			Specializations: []*models.Specialization{spec1, spec2},
		}
		mockDB.Veterinarians[1] = vet

		schedule := &models.Schedule{
			VetID:     1,
			ClinicID:  1,
			DayOfWeek: 1,
			StartTime: "09:00",
			EndTime:   "18:00",
			Clinic:    clinic,
		}
		mockDB.Schedules[1] = schedule

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		// Act & Assert - последовательность действий пользователя

		// 1. Пользователь начинает
		startUpdate := NewTestUpdate().
			WithMessage("/start", 12345, 67890).
			Build()
		handlers.HandleStart(startUpdate)

		// 2. Пользователь выбирает специализации
		specUpdate := NewTestUpdate().
			WithCallback("main_specializations", 12345, 1).
			Build()
		handlers.HandleSpecializations(specUpdate)

		// 3. Пользователь выбирает конкретную специализацию
		searchUpdate := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 2).
			Build()
		handlers.HandleCallback(searchUpdate)

		// Assert - проверяем, что поток работает корректно
		assert.True(t, len(mockBot.SentMessages) >= 2 || len(mockBot.EditedMessages) >= 1)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ ДНЕЙ НЕДЕЛИ
// ============================================================================

func TestVetHandleDaySelection(t *testing.T) {
	t.Run("Search by day with results", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Создаем ветеринара
		vet := &models.Veterinarian{
			ID:        sql.NullInt64{Int64: 1, Valid: true},
			FirstName: "Сергей",
			LastName:  "Кузнецов",
			Phone:     "+79123456789",
		}
		mockDB.Veterinarians[1] = vet

		// Создаем расписание на понедельник
		schedule := &models.Schedule{
			VetID:     1,
			DayOfWeek: 1,
			StartTime: "09:00",
			EndTime:   "18:00",
		}
		mockDB.Schedules[1] = schedule

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_day_1", 12345, 1).
			Build()

		// Act
		handlers.handleDaySelection(update.CallbackQuery)

		// Assert
		var messageText string
		if len(mockBot.SentMessages) > 0 {
			messageText = mockBot.GetLastMessage().Text
		} else if len(mockBot.EditedMessages) > 0 {
			messageText = mockBot.GetLastEditedMessage().Text
		}

		assert.Contains(t, messageText, "Сергей Кузнецов")
		assert.Contains(t, messageText, "+79123456789")
		assert.Contains(t, messageText, "понедельник: 09:00-18:00")
	})

	t.Run("Search by day with no results", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// Врач есть, но не работает в выбранный день
		vet := &models.Veterinarian{
			ID:        sql.NullInt64{Int64: 1, Valid: true},
			FirstName: "Сергей",
			LastName:  "Кузнецов",
			Phone:     "+79123456789",
		}
		mockDB.Veterinarians[1] = vet

		// Расписание только на вторник
		schedule := &models.Schedule{
			VetID:     1,
			DayOfWeek: 2, // Вторник
			StartTime: "09:00",
			EndTime:   "18:00",
		}
		mockDB.Schedules[1] = schedule

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_day_1", 12345, 1). // Понедельник
			Build()

		// Act
		handlers.handleDaySelection(update.CallbackQuery)

		// Assert
		var messageText string
		if len(mockBot.SentMessages) > 0 {
			messageText = mockBot.GetLastMessage().Text
		} else if len(mockBot.EditedMessages) > 0 {
			messageText = mockBot.GetLastEditedMessage().Text
		}

		assert.Contains(t, messageText, "не найдены")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ ОШИБОК БАЗЫ ДАННЫХ
// ============================================================================

func TestVetDatabaseErrorHandling(t *testing.T) {
	t.Run("Database error in search", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		mockDB.VeterinariansError = fmt.Errorf("database connection failed")
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("search_spec_1", 12345, 1).
			Build()

		// Act
		handlers.HandleSearchBySpecialization(update, 1)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Ошибка при поиске врачей")
	})

	t.Run("Database error in clinics", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()
		mockDB.ClinicsError = fmt.Errorf("database connection failed")
		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		update := NewTestUpdate().
			WithCallback("main_clinics", 12345, 1).
			Build()

		// Act
		handlers.HandleClinics(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Ошибка при получении списка клиник")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ ПАРСИНГА CALLBACK ДАННЫХ
// ============================================================================

func TestVetCallbackParsing(t *testing.T) {
	t.Run("Parse valid specialization ID", func(t *testing.T) {
		// Arrange
		callback := &tgbotapi.CallbackQuery{
			Data: "search_spec_42",
			// ID и Message могут понадобиться для будущих тестов
			ID: "test_callback",
			Message: &tgbotapi.Message{
				MessageID: 1,
				Chat:      &tgbotapi.Chat{ID: 12345},
			},
		}
		_ = callback.ID      // Используем чтобы избежать warning
		_ = callback.Message // Используем чтобы избежать warning

		// Act
		specIDStr := strings.TrimPrefix(callback.Data, "search_spec_")
		specID, err := strconv.Atoi(specIDStr)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 42, specID)
	})

	t.Run("Parse invalid specialization ID", func(t *testing.T) {
		// Arrange
		callback := &tgbotapi.CallbackQuery{
			Data: "search_spec_invalid",
			ID:   "test_callback",
			Message: &tgbotapi.Message{
				MessageID: 1,
				Chat:      &tgbotapi.Chat{ID: 12345},
			},
		}
		_ = callback.ID
		_ = callback.Message

		// Act
		specIDStr := strings.TrimPrefix(callback.Data, "search_spec_")
		_, err := strconv.Atoi(specIDStr)

		// Assert
		assert.Error(t, err)
	})

	t.Run("Parse valid clinic ID", func(t *testing.T) {
		// Arrange
		callback := &tgbotapi.CallbackQuery{
			Data: "search_clinic_15",
			ID:   "test_callback",
			Message: &tgbotapi.Message{
				MessageID: 1,
				Chat:      &tgbotapi.Chat{ID: 12345},
			},
		}
		_ = callback.ID
		_ = callback.Message

		// Act
		clinicIDStr := strings.TrimPrefix(callback.Data, "search_clinic_")
		clinicID, err := strconv.Atoi(clinicIDStr)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 15, clinicID)
	})
}

func TestVetReviewFunctionality(t *testing.T) {
	t.Run("Add review callback", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// НАСТРОЙТЕ МОКИ ДЛЯ ОТЗЫВОВ ЧЕРЕЗ ФУНКЦИОНАЛЬНЫЕ ПОЛЯ
		mockDB.HasUserReviewForVetFunc = func(userID int, vetID int) (bool, error) {
			return false, nil // Пользователь еще не оставлял отзыв
		}

		mockDB.GetUserByTelegramIDFunc = func(telegramID int64) (*models.User, error) {
			return &models.User{
				ID:         1,
				TelegramID: telegramID,
				FirstName:  "Test",
				LastName:   "User",
			}, nil
		}

		mockDB.CreateReviewFunc = func(review *models.Review) error {
			return nil // Успешное создание отзыва
		}

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		// СОЗДАЙТЕ ПОЛНЫЙ CALLBACK
		callback := &tgbotapi.CallbackQuery{
			ID:   "test_callback",
			Data: "add_review_1",
			Message: &tgbotapi.Message{
				MessageID: 1,
				Chat: &tgbotapi.Chat{
					ID: 12345,
				},
				From: &tgbotapi.User{
					ID: 67890,
				},
			},
			From: &tgbotapi.User{
				ID: 67890,
			},
		}

		// Act
		handlers.handleAddReviewCallback(callback)

		// Assert
		// Проверяем, что начался процесс добавления отзыва
		assert.NotEmpty(t, mockBot.SentMessages)
	})

	t.Run("Show reviews callback", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		mockDB := NewMockDatabase()

		// НАСТРОЙТЕ МОКИ ДЛЯ ПОКАЗА ОТЗЫВОВ
		mockDB.GetApprovedReviewsByVetFunc = func(vetID int) ([]*models.Review, error) {
			return []*models.Review{
				{
					ID:             1,
					VeterinarianID: vetID,
					UserID:         1,
					Rating:         5,
					Comment:        "Отличный врач!",
					Status:         "approved",
					CreatedAt:      time.Now(),
					User: &models.User{
						FirstName: "Анна",
					},
					Veterinarian: &models.Veterinarian{
						FirstName: "Иван",
						LastName:  "Петров",
					},
				},
			}, nil
		}

		mockDB.GetReviewStatsFunc = func(vetID int) (*models.ReviewStats, error) {
			return &models.ReviewStats{
				VeterinarianID:  vetID,
				AverageRating:   4.5,
				TotalReviews:    1,
				ApprovedReviews: 1,
			}, nil
		}

		stateManager := NewTestStateManager()
		handlers := NewVetHandlers(mockBot, mockDB, []int64{12345}, stateManager)

		callback := &tgbotapi.CallbackQuery{
			ID:   "test_callback",
			Data: "show_reviews_1",
			Message: &tgbotapi.Message{
				MessageID: 1,
				Chat: &tgbotapi.Chat{
					ID: 12345,
				},
				From: &tgbotapi.User{
					ID: 67890,
				},
			},
			From: &tgbotapi.User{
				ID: 67890,
			},
		}

		// Act
		handlers.handleShowReviewsCallback(callback)

		// Assert
		// Проверяем, что был вызван функционал отзывов
		assert.NotEmpty(t, mockBot.SentMessages)
	})
}
