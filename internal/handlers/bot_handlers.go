package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotHandlers содержит базовые обработчики бота
type BotHandlers struct {
	bot BotAPI // Используем интерфейс вместо конкретной реализации
}

// NewBotHandlers создает новый экземпляр BotHandlers
func NewBotHandlers(bot BotAPI) *BotHandlers {
	return &BotHandlers{
		bot: bot,
	}
}

// HandleUnknownCommand обрабатывает неизвестные команды
func (h *BotHandlers) HandleUnknownCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"❌ Неизвестная команда.\n\n"+
			"Используйте /help для просмотра доступных команд.")
	h.bot.Send(msg)
}

// HandleErrorMessage обрабатывает ошибки
func (h *BotHandlers) HandleErrorMessage(chatID int64, errorMsg string) {
	msg := tgbotapi.NewMessage(chatID,
		"⚠️ Произошла ошибка. Пожалуйста, попробуйте позже.")
	h.bot.Send(msg)

	// Логируем ошибку для администратора
	ErrorLog.Printf("Error for user %d: %s", chatID, errorMsg)
}

// SendWelcomeMessage отправляет приветственное сообщение
func (h *BotHandlers) SendWelcomeMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID,
		`🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Я могу:

• Показать врачей по специализации
• Найти врачей на конкретный день
• Показать контакты клиник
• Предоставить расписание приема

Начните с команды /help чтобы увидеть все возможности!`)
	h.bot.Send(msg)
}

// GetToken возвращает токен бота
func (b *BotHandlers) GetToken() string {
	return b.bot.GetToken()
}
