// internal/handlers/bot_adapter.go
package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBotAdapter адаптирует tgbotapi.BotAPI к нашему интерфейсу BotAPI
type TelegramBotAdapter struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramBotAdapter создает новый адаптер
func NewTelegramBotAdapter(bot *tgbotapi.BotAPI) *TelegramBotAdapter {
	return &TelegramBotAdapter{bot: bot}
}

// Send отправляет сообщение
func (a *TelegramBotAdapter) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return a.bot.Send(c)
}

// GetFile получает файл
func (a *TelegramBotAdapter) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return a.bot.GetFile(config)
}

// Request выполняет запрос к API
func (a *TelegramBotAdapter) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return a.bot.Request(c)
}

// GetToken возвращает токен бота
func (a *TelegramBotAdapter) GetToken() string {
	return a.bot.Token
}
