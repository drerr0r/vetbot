package handlers

import (
	"log"
	"strings"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler главный обработчик, который распределяет обновления между специализированными обработчиками
type MainHandler struct {
	botHandlers   *BotHandlers
	adminHandlers *AdminHandlers
}

// NewMainHandler создает новый экземпляр главного обработчика
func NewMainHandler(botHandlers *BotHandlers, adminHandlers *AdminHandlers) *MainHandler {
	return &MainHandler{
		botHandlers:   botHandlers,
		adminHandlers: adminHandlers,
	}
}

// HandleUpdate обрабатывает входящее обновление и распределяет его между обработчиками
func (h *MainHandler) HandleUpdate(update telegram.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Паника при обработке обновления: %v", r)
		}
	}()

	if update.Message == nil {
		return
	}

	log.Printf("📨 Получено сообщение от %s: %s", update.Message.From.UserName, update.Message.Text)

	// Регистрируем пользователя
	h.botHandlers.RegisterUser(update.Message.From.UserName, update.Message.Chat.ID)

	// Определяем тип команды и передаем соответствующему обработчику
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "admin":
			h.adminHandlers.HandleAdminCommand(update)
		default:
			h.botHandlers.HandleCommand(update)
		}
	} else if strings.HasPrefix(update.Message.Text, "/find") {
		h.botHandlers.HandleFindCommand(update)
	} else {
		h.botHandlers.HandleDefaultMessage(update)
	}
}
