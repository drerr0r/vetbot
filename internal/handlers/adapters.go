package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBotAdapter адаптер для реального tgbotapi.BotAPI
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

// Request выполняет запрос к API
func (a *TelegramBotAdapter) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return a.bot.Request(c)
}
