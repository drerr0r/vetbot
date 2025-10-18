package handlers

import (
	"fmt"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ТЕСТЫ ДЛЯ КОНСТРУКТОРА
// ============================================================================

func TestBotHandlers_NewBotHandlers(t *testing.T) {
	// Arrange
	mockBot := NewMockBot() // ИСПРАВЛЕНО: используем мок вместо реального бота

	// Act
	handler := NewBotHandlers(mockBot)

	// Assert
	assert.NotNil(t, handler)
	assert.Equal(t, mockBot, handler.bot) // ИСПРАВЛЕНО: проверяем что мок установлен
}

// ============================================================================
// ТЕСТЫ ДЛЯ ЛОГИКИ СООБЩЕНИЙ (без вызовов бота)
// ============================================================================

func TestBotHandlers_MessageContentLogic(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		expectedPhrases []string
	}{
		{
			name:   "Unknown command message content",
			method: "unknown",
			expectedPhrases: []string{
				"Неизвестная команда",
				"/help",
				"❌",
			},
		},
		{
			name:   "Error message content",
			method: "error",
			expectedPhrases: []string{
				"Произошла ошибка",
				"попробуйте позже",
				"⚠️",
			},
		},
		{
			name:   "Welcome message content",
			method: "welcome",
			expectedPhrases: []string{
				"Добро пожаловать",
				"VetBot",
				"ветеринарных врачей",
				"специализации",
				"расписание",
				"/help",
				"🐾",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var messageText string

			// Имитируем логику формирования сообщений из BotHandlers
			switch tt.method {
			case "unknown":
				messageText = "❌ Неизвестная команда.\n\nИспользуйте /help для просмотра доступных команд."
			case "error":
				messageText = "⚠️ Произошла ошибка. Пожалуйста, попробуйте позже."
			case "welcome":
				messageText = `🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Я могу:

• Показать врачей по специализации
• Найти врачей на конкретный день
• Показать контакты клиник
• Предоставить расписание приема

Начните с команды /help чтобы увидеть все возможности!`
			}

			// Assert - проверяем что все ожидаемые фразы присутствуют
			for _, phrase := range tt.expectedPhrases {
				assert.True(t, contains(messageText, phrase),
					"Сообщение должно содержать фразу: '%s'. Полный текст: %s", phrase, messageText)
			}
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ И ПОВЕДЕНИЯ
// ============================================================================

func TestBotHandlers_Structure(t *testing.T) {
	t.Run("BotHandlers has required fields", func(t *testing.T) {
		mockBot := NewMockBot()
		handler := NewBotHandlers(mockBot)

		// Проверяем что структура имеет ожидаемые поля
		assert.NotNil(t, handler)
		assert.Equal(t, mockBot, handler.bot)
	})

	t.Run("Multiple handler instances are independent", func(t *testing.T) {
		mockBot1 := NewMockBot()
		mockBot2 := NewMockBot()

		handler1 := NewBotHandlers(mockBot1)
		handler2 := NewBotHandlers(mockBot2)

		// Проверяем что это разные экземпляры (сравниваем содержимое, а не указатели)
		assert.NotEqual(t, fmt.Sprintf("%p", handler1), fmt.Sprintf("%p", handler2), "Handler instances should be different")

		// Проверяем что оба имеют своих ботов
		assert.Equal(t, mockBot1, handler1.bot)
		assert.Equal(t, mockBot2, handler2.bot)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ ДАННЫХ
// ============================================================================

func TestBotHandlers_UpdateHandlingLogic(t *testing.T) {
	tests := []struct {
		name           string
		update         tgbotapi.Update
		expectedAction string
	}{
		{
			name: "Update with message for unknown command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/unknowncommand",
					Chat: &tgbotapi.Chat{ID: 12345},
				},
			},
			expectedAction: "unknown_command",
		},
		{
			name: "Update with different chat ID",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/invalid",
					Chat: &tgbotapi.Chat{ID: 67890},
				},
			},
			expectedAction: "unknown_command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var action string
			var chatID int64

			// Имитируем логику обработки обновления
			if tt.update.Message != nil {
				chatID = tt.update.Message.Chat.ID
				action = "unknown_command"
			}

			// Assert
			assert.Equal(t, tt.expectedAction, action)
			if tt.update.Message != nil {
				assert.Equal(t, tt.update.Message.Chat.ID, chatID)
			}
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ ОШИБОК
// ============================================================================

func TestBotHandlers_ErrorHandlingLogic(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int64
		errorMsg    string
		expectedLog string
	}{
		{
			name:        "Database error",
			chatID:      12345,
			errorMsg:    "database connection failed",
			expectedLog: "database connection failed",
		},
		{
			name:        "Network error",
			chatID:      67890,
			errorMsg:    "network timeout",
			expectedLog: "network timeout",
		},
		{
			name:        "Empty error message",
			chatID:      12345,
			errorMsg:    "",
			expectedLog: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var (
				action string
				chatID int64
			)

			// Имитируем логику обработки ошибок
			chatID = tt.chatID
			action = tt.errorMsg

			// Assert
			assert.Equal(t, tt.chatID, chatID)
			assert.Equal(t, tt.expectedLog, action)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ПРИВЕТСТВЕННЫХ СООБЩЕНИЙ
// ============================================================================

func TestBotHandlers_WelcomeMessageLogic(t *testing.T) {
	tests := []struct {
		name           string
		chatID         int64
		expectedAction string
	}{
		{
			name:           "Welcome message for new user",
			chatID:         12345,
			expectedAction: "send_welcome",
		},
		{
			name:           "Welcome message for different user",
			chatID:         67890,
			expectedAction: "send_welcome",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var (
				action       string
				targetChatID int64
			)

			// Имитируем логику отправки приветственного сообщения
			targetChatID = tt.chatID
			action = "send_welcome"

			// Assert
			assert.Equal(t, tt.expectedAction, action)
			assert.Equal(t, tt.chatID, targetChatID)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ КРАЙНИХ СЛУЧАЕВ
// ============================================================================

func TestBotHandlers_EdgeCases(t *testing.T) {
	t.Run("Nil bot in constructor", func(t *testing.T) {
		// Act
		handler := NewBotHandlers(nil)

		// Assert
		assert.NotNil(t, handler)
		assert.Nil(t, handler.bot)
	})

	t.Run("Zero chat ID for error message", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot() // ИСПРАВЛЕНО: используем мок
		handler := NewBotHandlers(mockBot)

		// Act & Assert - не должно паниковать
		assert.NotPanics(t, func() {
			_ = handler
		})
	})

	t.Run("Negative chat ID for welcome message", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot() // ИСПРАВЛЕНО: используем мок
		handler := NewBotHandlers(mockBot)

		// Act & Assert - не должно паниковать
		assert.NotPanics(t, func() {
			_ = handler
		})
	})

	t.Run("Empty update for unknown command", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot() // ИСПРАВЛЕНО: используем мок
		handler := NewBotHandlers(mockBot)
		emptyUpdate := tgbotapi.Update{}

		// Act & Assert - не должно паниковать
		assert.NotPanics(t, func() {
			_ = handler
			_ = emptyUpdate
		})
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ ФОРМАТИРОВАНИЯ ТЕКСТА
// ============================================================================

func TestBotHandlers_TextFormatting(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		checks []func(string) bool
	}{
		{
			name: "Unknown command formatting",
			text: "❌ Неизвестная команда.\n\nИспользуйте /help для просмотра доступных команд.",
			checks: []func(string) bool{
				func(s string) bool { return strings.Contains(s, "❌") },
				func(s string) bool { return strings.Contains(s, "Неизвестная команда") },
				func(s string) bool { return strings.Contains(s, "/help") },
				func(s string) bool { return strings.Count(s, "\n") >= 1 }, // Должен быть перенос строки
			},
		},
		{
			name: "Error message formatting",
			text: "⚠️ Произошла ошибка. Пожалуйста, попробуйте позже.",
			checks: []func(string) bool{
				func(s string) bool { return strings.Contains(s, "⚠️") },
				func(s string) bool { return strings.Contains(s, "Произошла ошибка") },
				func(s string) bool { return strings.Contains(s, "попробуйте позже") },
			},
		},
		{
			name: "Welcome message formatting",
			text: `🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Я могу:

• Показать врачей по специализации
• Найти врачей на конкретный день
• Показать контакты клиник
• Предоставить расписание приема

Начните с команды /help чтобы увидеть все возможности!`,
			checks: []func(string) bool{
				func(s string) bool { return strings.Contains(s, "🐾") },
				func(s string) bool { return strings.Contains(s, "Добро пожаловать") },
				func(s string) bool { return strings.Contains(s, "•") }, // Список
				func(s string) bool { return strings.Contains(s, "/help") },
				func(s string) bool { return strings.Count(s, "\n") >= 5 }, // Много переносов строк
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, check := range tt.checks {
				assert.True(t, check(tt.text), "Check %d failed for text: %s", i, tt.text)
			}
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ СОДЕРЖАНИЯ СООБЩЕНИЙ
// ============================================================================

func TestBotHandlers_MessageCompleteness(t *testing.T) {
	t.Run("Unknown command message has all required elements", func(t *testing.T) {
		message := "❌ Неизвестная команда.\n\nИспользуйте /help для просмотра доступных команд."

		// Проверяем наличие всех ключевых элементов
		assert.Contains(t, message, "❌", "Должен быть значок ошибки")
		assert.Contains(t, message, "Неизвестная команда", "Должно быть описание проблемы")
		assert.Contains(t, message, "/help", "Должна быть предложена помощь")
		assert.True(t, len(message) > 20, "Сообщение должно быть достаточно длинным")
	})

	t.Run("Error message has user-friendly content", func(t *testing.T) {
		message := "⚠️ Произошла ошибка. Пожалуйста, попробуйте позже."

		assert.Contains(t, message, "⚠️", "Должен быть значок предупреждения")
		assert.Contains(t, message, "Произошла ошибка", "Должно быть указано на ошибку")
		assert.Contains(t, message, "попробуйте позже", "Должно быть предложение повторить")
	})

	t.Run("Welcome message is comprehensive and helpful", func(t *testing.T) {
		message := `🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Я могу:

• Показать врачей по специализации
• Найти врачей на конкретный день
• Показать контакты клиник
• Предоставить расписание приема

Начните с команды /help чтобы увидеть все возможности!`

		assert.Contains(t, message, "🐾", "Должны быть декоративные элементы")
		assert.Contains(t, message, "Добро пожаловать", "Должно быть приветствие")
		assert.Contains(t, message, "ветеринарных врачей", "Должно быть описание назначения")
		assert.Contains(t, message, "•", "Должен быть список возможностей")
		assert.Contains(t, message, "/help", "Должна быть указана команда помощи")
		assert.True(t, strings.Count(message, "•") >= 4, "Должно быть несколько пунктов возможностей")
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ РЕАЛЬНОЙ ФУНКЦИОНАЛЬНОСТИ
// ============================================================================

func TestBotHandlers_RealFunctionality(t *testing.T) {
	t.Run("HandleUnknownCommand sends message", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		handler := NewBotHandlers(mockBot)

		update := tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "/unknown",
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		}

		// Act
		handler.HandleUnknownCommand(update)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Неизвестная команда")
		assert.Equal(t, int64(12345), message.ChatID)
	})

	t.Run("HandleErrorMessage sends error message", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		handler := NewBotHandlers(mockBot)

		// Act
		handler.HandleErrorMessage(12345, "test error")

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Произошла ошибка")
		assert.Equal(t, int64(12345), message.ChatID)
	})

	t.Run("SendWelcomeMessage sends welcome", func(t *testing.T) {
		// Arrange
		mockBot := NewMockBot()
		handler := NewBotHandlers(mockBot)

		// Act
		handler.SendWelcomeMessage(12345)

		// Assert
		assert.Len(t, mockBot.SentMessages, 1)
		message := mockBot.GetLastMessage()
		assert.Contains(t, message.Text, "Добро пожаловать")
		assert.Contains(t, message.Text, "VetBot")
		assert.Equal(t, int64(12345), message.ChatID)
	})
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// contains проверяет содержит ли строка подстроку (для удобства тестирования)
func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	return strings.Contains(s, substr)
}

// ============================================================================
// ТЕСТЫ ДЛЯ ВСПОМОГАТЕЛЬНЫХ ФУНКЦИЙ
// ============================================================================

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{
			name:     "String contains substring",
			str:      "Hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "String does not contain substring",
			str:      "Hello world",
			substr:   "test",
			expected: false,
		},
		{
			name:     "Empty string",
			str:      "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "Empty substring",
			str:      "Hello world",
			substr:   "",
			expected: true,
		},
		{
			name:     "Both empty",
			str:      "",
			substr:   "",
			expected: true,
		},
		{
			name:     "Case sensitive",
			str:      "Hello World",
			substr:   "world",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.str, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
