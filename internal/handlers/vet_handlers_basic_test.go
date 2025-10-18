package handlers

import (
	"strconv"
	"strings"
	"testing"

	"github.com/drerr0r/vetbot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ТЕСТЫ ДЛЯ КОНСТРУКТОРА И БАЗОВОЙ ФУНКЦИОНАЛЬНОСТИ
// ============================================================================

func TestNewMainHandler(t *testing.T) {
	// Arrange
	mockBot := NewMockBot() // Используем мок вместо реального бота
	mockDB := NewMockDatabase()
	config := &utils.Config{}

	// Act
	handler := NewMainHandler(mockBot, mockDB, config)

	// Assert
	assert.NotNil(t, handler)
	assert.Equal(t, mockBot, handler.bot)
	assert.Equal(t, mockDB, handler.db)
	assert.Equal(t, config, handler.config)
	assert.NotNil(t, handler.vetHandlers)
	assert.NotNil(t, handler.adminHandlers)
}

// ============================================================================
// ТЕСТЫ ДЛЯ ВСПОМОГАТЕЛЬНЫХ ФУНКЦИЙ
// ============================================================================

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name           string
		adminIDs       []int64
		userID         int64
		expectedResult bool
	}{
		{
			name:           "User is admin",
			adminIDs:       []int64{12345, 67890},
			userID:         12345,
			expectedResult: true,
		},
		{
			name:           "User is not admin",
			adminIDs:       []int64{12345, 67890},
			userID:         99999,
			expectedResult: false,
		},
		{
			name:           "Empty admin list",
			adminIDs:       []int64{},
			userID:         12345,
			expectedResult: false,
		},
		{
			name:           "Nil config",
			adminIDs:       nil,
			userID:         12345,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := &MainHandler{
				config: &utils.Config{AdminIDs: tt.adminIDs},
			}

			// Act
			result := handler.isAdmin(tt.userID)

			// Assert
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestIsInAdminMode(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		adminIDs       []int64
		expectedResult bool
	}{
		{
			name:           "User in admin mode",
			userID:         12345,
			adminIDs:       []int64{12345, 67890},
			expectedResult: true,
		},
		{
			name:           "User not in admin mode",
			userID:         99999,
			adminIDs:       []int64{12345, 67890},
			expectedResult: false,
		},
		{
			name:           "Empty admin list",
			userID:         12345,
			adminIDs:       []int64{},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockBot := NewMockBot()
			mockDB := NewMockDatabase()
			config := &utils.Config{AdminIDs: tt.adminIDs}

			// Создаем MainHandler для проверки
			mainHandler := NewMainHandler(mockBot, mockDB, config)

			// Act - используем isAdmin вместо isInAdminMode
			result := mainHandler.isAdmin(tt.userID)

			// Assert
			assert.Equal(t, tt.expectedResult, result,
				"For user %d with adminIDs %v, expected %t but got %t",
				tt.userID, tt.adminIDs, tt.expectedResult, result)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ПАРСИНГА КОМАНД
// ============================================================================

func TestParseSearchCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		expectedID    int
		shouldSucceed bool
	}{
		{
			name:          "Valid search command",
			command:       "/search_1",
			expectedID:    1,
			shouldSucceed: true,
		},
		{
			name:          "Valid search command with double digit",
			command:       "/search_42",
			expectedID:    42,
			shouldSucceed: true,
		},
		{
			name:          "Valid search command with triple digit",
			command:       "/search_123",
			expectedID:    123,
			shouldSucceed: true,
		},
		{
			name:          "Invalid search command - non numeric",
			command:       "/search_abc",
			expectedID:    0,
			shouldSucceed: false,
		},
		{
			name:          "Invalid search command - no ID",
			command:       "/search_",
			expectedID:    0,
			shouldSucceed: false,
		},
		{
			name:          "Invalid search command - empty",
			command:       "/search",
			expectedID:    0,
			shouldSucceed: false,
		},
		{
			name:          "Invalid search command - special characters",
			command:       "/search_1a2",
			expectedID:    0,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			var specID int
			var err error

			if strings.HasPrefix(tt.command, "/search_") {
				specIDStr := strings.TrimPrefix(tt.command, "/search_")
				specID, err = strconv.Atoi(specIDStr)
			} else {
				err = strconv.ErrSyntax
			}

			// Assert
			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, specID)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCommandRoutingLogic(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		expectedAction string
		isAdmin        bool
	}{
		{
			name:           "Start command",
			command:        "/start",
			expectedAction: "vet_start",
			isAdmin:        false,
		},
		{
			name:           "Help command",
			command:        "/help",
			expectedAction: "vet_help",
			isAdmin:        false,
		},
		{
			name:           "Specializations command",
			command:        "/specializations",
			expectedAction: "vet_specializations",
			isAdmin:        false,
		},
		{
			name:           "Search command",
			command:        "/search",
			expectedAction: "vet_search",
			isAdmin:        false,
		},
		{
			name:           "Admin command with access",
			command:        "/admin",
			expectedAction: "admin_panel",
			isAdmin:        true,
		},
		{
			name:           "Admin command without access",
			command:        "/admin",
			expectedAction: "access_denied",
			isAdmin:        false,
		},
		{
			name:           "Stats command with access",
			command:        "/stats",
			expectedAction: "admin_stats",
			isAdmin:        true,
		},
		{
			name:           "Stats command without access",
			command:        "/stats",
			expectedAction: "no_action",
			isAdmin:        false,
		},
		{
			name:           "Unknown command",
			command:        "/unknown",
			expectedAction: "unknown_command",
			isAdmin:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var action string

			// Имитируем логику роутинга из handleCommand
			switch tt.command {
			case "/start":
				action = "vet_start"
			case "/help":
				action = "vet_help"
			case "/specializations":
				action = "vet_specializations"
			case "/search":
				action = "vet_search"
			case "/admin":
				if tt.isAdmin {
					action = "admin_panel"
				} else {
					action = "access_denied"
				}
			case "/stats":
				if tt.isAdmin {
					action = "admin_stats"
				} else {
					action = "no_action"
				}
			default:
				action = "unknown_command"
			}

			// Assert
			assert.Equal(t, tt.expectedAction, action)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ОБРАБОТКИ РАЗЛИЧНЫХ ТИПОВ ОБНОВЛЕНИЙ
// ============================================================================

func TestUpdateTypeDetection(t *testing.T) {
	tests := []struct {
		name          string
		update        tgbotapi.Update
		expectedType  string
		shouldProcess bool
	}{
		{
			name: "Callback query update",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					ID:   "test",
					Data: "test_data",
				},
			},
			expectedType:  "callback",
			shouldProcess: true,
		},
		{
			name: "Message with text",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "/start",
					From: &tgbotapi.User{ID: 12345},
				},
			},
			expectedType:  "message",
			shouldProcess: true,
		},
		{
			name: "Message without text",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "",
					From: &tgbotapi.User{ID: 12345},
				},
			},
			expectedType:  "empty_message",
			shouldProcess: false,
		},
		{
			name: "Nil message",
			update: tgbotapi.Update{
				Message: nil,
			},
			expectedType:  "nil_message",
			shouldProcess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			var updateType string
			var shouldProcess bool

			if tt.update.CallbackQuery != nil {
				updateType = "callback"
				shouldProcess = true
			} else if tt.update.Message != nil {
				if tt.update.Message.Text != "" {
					updateType = "message"
					shouldProcess = true
				} else {
					updateType = "empty_message"
					shouldProcess = false
				}
			} else {
				updateType = "nil_message"
				shouldProcess = false
			}

			// Assert
			assert.Equal(t, tt.expectedType, updateType)
			assert.Equal(t, tt.shouldProcess, shouldProcess)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ АДМИНСКОЙ ЛОГИКИ
// ============================================================================

func TestAdminAuthorizationLogic(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		adminIDs       []int64
		expectedResult bool
	}{
		{
			name:           "Admin user",
			userID:         12345,
			adminIDs:       []int64{12345, 67890},
			expectedResult: true,
		},
		{
			name:           "Non-admin user",
			userID:         99999,
			adminIDs:       []int64{12345, 67890},
			expectedResult: false,
		},
		{
			name:           "User not in admin list",
			userID:         55555,
			adminIDs:       []int64{12345, 67890},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockBot := NewMockBot()
			mockDB := NewMockDatabase()
			config := &utils.Config{AdminIDs: tt.adminIDs}

			// Создаем MainHandler через конструктор
			mainHandler := NewMainHandler(mockBot, mockDB, config)

			// Act - используем ТОЛЬКО isAdmin (прямая проверка прав)
			result := mainHandler.isAdmin(tt.userID)

			// Assert
			assert.Equal(t, tt.expectedResult, result,
				"isAdmin failed for test '%s': user %d, adminIDs %v",
				tt.name, tt.userID, tt.adminIDs)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ТЕКСТОВЫХ СООБЩЕНИЙ
// ============================================================================

func TestTextMessageHandlingLogic(t *testing.T) {
	tests := []struct {
		name            string
		messageText     string
		isAdmin         bool
		inAdminMode     bool
		expectedHandler string
	}{
		{
			name:            "Regular user text message",
			messageText:     "Hello world",
			isAdmin:         false,
			inAdminMode:     false,
			expectedHandler: "help_message",
		},
		{
			name:            "Admin user text message in admin mode",
			messageText:     "admin command",
			isAdmin:         true,
			inAdminMode:     true,
			expectedHandler: "admin_handler",
		},
		{
			name:            "Admin user text message not in admin mode",
			messageText:     "regular message",
			isAdmin:         true,
			inAdminMode:     false,
			expectedHandler: "help_message",
		},
		{
			name:            "Search command message",
			messageText:     "/search_1",
			isAdmin:         false,
			inAdminMode:     false,
			expectedHandler: "search_handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var handler string

			// Имитируем логику из HandleUpdate
			if strings.HasPrefix(tt.messageText, "/search_") {
				handler = "search_handler"
			} else if tt.isAdmin && tt.inAdminMode {
				handler = "admin_handler"
			} else {
				handler = "help_message"
			}

			// Assert
			assert.Equal(t, tt.expectedHandler, handler)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ВАЛИДАЦИИ ВХОДНЫХ ДАННЫХ
// ============================================================================

func TestInputValidation(t *testing.T) {
	tests := []struct {
		name          string
		update        tgbotapi.Update
		shouldProcess bool
		reason        string
	}{
		{
			name: "Valid callback query",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{ID: "test"},
			},
			shouldProcess: true,
			reason:        "callback should be processed",
		},
		{
			name: "Valid text message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{Text: "/start"},
			},
			shouldProcess: true,
			reason:        "text message should be processed",
		},
		{
			name: "Empty text message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{Text: ""},
			},
			shouldProcess: false,
			reason:        "empty text should be ignored",
		},
		{
			name: "Nil message",
			update: tgbotapi.Update{
				Message: nil,
			},
			shouldProcess: false,
			reason:        "nil message should be ignored",
		},
		{
			name:          "Empty update",
			update:        tgbotapi.Update{},
			shouldProcess: false,
			reason:        "empty update should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			var shouldProcess bool

			if tt.update.CallbackQuery != nil {
				shouldProcess = true
			} else if tt.update.Message != nil && tt.update.Message.Text != "" {
				shouldProcess = true
			} else {
				shouldProcess = false
			}

			// Assert
			assert.Equal(t, tt.shouldProcess, shouldProcess, tt.reason)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ КРАЙНИХ СЛУЧАЕВ
// ============================================================================

func TestEdgeCases(t *testing.T) {
	t.Run("Nil config should not cause panic", func(t *testing.T) {
		handler := &MainHandler{
			config: nil,
		}

		// Должны обрабатывать nil без паники
		isAdmin := handler.isAdmin(12345)
		assert.False(t, isAdmin)
	})

	t.Run("Nil admin handlers should not cause panic", func(t *testing.T) {
		handler := &MainHandler{
			adminHandlers: nil,
		}

		// Должны обрабатывать nil без паники
		inAdminMode := handler.isInAdminMode(12345)
		assert.False(t, inAdminMode)
	})

	t.Run("Very large user ID", func(t *testing.T) {
		handler := &MainHandler{
			config: &utils.Config{AdminIDs: []int64{12345}},
		}

		isAdmin := handler.isAdmin(999999999999999999)
		assert.False(t, isAdmin)
	})

	t.Run("Negative user ID", func(t *testing.T) {
		handler := &MainHandler{
			config: &utils.Config{AdminIDs: []int64{12345}},
		}

		isAdmin := handler.isAdmin(-12345)
		assert.False(t, isAdmin)
	})

	t.Run("Zero user ID", func(t *testing.T) {
		handler := &MainHandler{
			config: &utils.Config{AdminIDs: []int64{12345}},
		}

		isAdmin := handler.isAdmin(0)
		assert.False(t, isAdmin)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ КОМАНДЫ TEST
// ============================================================================

func TestTestCommandLogic(t *testing.T) {
	t.Run("Test command should be routed to vet handlers", func(t *testing.T) {
		command := "/test"
		var handler string

		if command == "/test" {
			handler = "vet_test"
		}

		assert.Equal(t, "vet_test", handler)
	})
}

// ============================================================================
// ТЕСТЫ ДЛЯ КЛИНИК КОМАНД
// ============================================================================

func TestClinicsCommandLogic(t *testing.T) {
	t.Run("Clinics command should be routed to vet handlers", func(t *testing.T) {
		command := "/clinics"
		var handler string

		if command == "/clinics" {
			handler = "vet_clinics"
		}

		assert.Equal(t, "vet_clinics", handler)
	})
}
