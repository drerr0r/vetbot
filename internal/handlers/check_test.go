package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBasicCompilation проверяет базовую компиляцию
func TestBasicCompilation(t *testing.T) {
	// Проверяем что можем создать моки
	mockBot := NewMockBot()
	assert.NotNil(t, mockBot)

	mockDB := NewMockDatabase()
	assert.NotNil(t, mockDB)

	// Проверяем создание обработчиков
	vetHandlers := NewVetHandlers(mockBot, mockDB, []int64{12345})
	assert.NotNil(t, vetHandlers)

	config := CreateTestConfig()
	adminHandlers := NewAdminHandlers(mockBot, mockDB, config)
	assert.NotNil(t, adminHandlers)

	mainHandler := NewMainHandler(mockBot, mockDB, config)
	assert.NotNil(t, mainHandler)

	// Проверяем TestUpdateBuilder
	builder := NewTestUpdate()
	assert.NotNil(t, builder)

	update := builder.WithMessage("/start", 12345, 67890).Build()
	assert.NotNil(t, update)
	assert.NotNil(t, update.Message)

	// Проверяем методы моков
	vetHandlers.HandleTest(update)
	assert.Len(t, mockBot.SentMessages, 1)

	message := mockBot.GetLastMessage()
	assert.Contains(t, message.Text, "Тестовое сообщение")
}
