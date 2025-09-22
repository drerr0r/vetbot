package handlers

import (
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler обрабатывает все входящие сообщения
type MainHandler struct {
	bot           *tgbotapi.BotAPI
	db            *database.Database
	vetHandlers   *VetHandlers
	adminHandlers *AdminHandlers
}

// NewMainHandler создает новый экземпляр MainHandler
func NewMainHandler(bot *tgbotapi.BotAPI, db *database.Database) *MainHandler {
	return &MainHandler{
		bot:           bot,
		db:            db,
		vetHandlers:   NewVetHandlers(bot, db),
		adminHandlers: NewAdminHandlers(bot, db),
	}
}

// HandleUpdate обрабатывает входящее обновление от Telegram
func (h *MainHandler) HandleUpdate(update tgbotapi.Update) {
	// Обрабатываем callback queries (нажатия на inline кнопки)
	if update.CallbackQuery != nil {
		h.vetHandlers.HandleCallback(update)
		return
	}

	// Игнорируем любые не-text сообщения
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	// Проверяем, является ли пользователь администратором
	isAdmin := h.isAdmin(update.Message.From.ID)

	// Обрабатываем команды
	switch {
	case update.Message.IsCommand():
		h.handleCommand(update, isAdmin)
	case strings.HasPrefix(update.Message.Text, "/search_"):
		h.handleSearchCommand(update)
	default:
		h.handleTextMessage(update, isAdmin)
	}
}

// handleCommand обрабатывает текстовые команды
func (h *MainHandler) handleCommand(update tgbotapi.Update, isAdmin bool) {
	command := update.Message.Command()

	switch command {
	case "start":
		h.vetHandlers.HandleStart(update)
	case "specializations":
		h.vetHandlers.HandleSpecializations(update)
	case "search":
		h.vetHandlers.HandleSearch(update)
	case "clinics":
		h.vetHandlers.HandleClinics(update)
	case "help":
		h.vetHandlers.HandleHelp(update)
	case "admin":
		if isAdmin {
			h.adminHandlers.HandleAdmin(update)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав администратора")
			h.bot.Send(msg)
		}
	case "stats":
		if isAdmin {
			h.adminHandlers.HandleStats(update)
		}
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Неизвестная команда. Используйте /help для списка команд")
		h.bot.Send(msg)
	}
}

// handleSearchCommand обрабатывает команды поиска по специализации (/search_1, /search_2 и т.д.)
func (h *MainHandler) handleSearchCommand(update tgbotapi.Update) {
	text := update.Message.Text
	if strings.HasPrefix(text, "/search_") {
		specIDStr := strings.TrimPrefix(text, "/search_")
		specID, err := strconv.Atoi(specIDStr)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат команды поиска")
			h.bot.Send(msg)
			return
		}
		h.vetHandlers.HandleSearchBySpecialization(update, specID)
	}
}

// handleTextMessage обрабатывает обычные текстовые сообщения
func (h *MainHandler) handleTextMessage(update tgbotapi.Update, isAdmin bool) {
	// Если пользователь администратор, передаем сообщение админским хендлерам
	if isAdmin {
		h.adminHandlers.HandleAdminMessage(update)
		return
	}

	// Для обычных пользователей показываем справку
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Я понимаю только команды. Используйте /help для списка доступных команд.")
	h.bot.Send(msg)
}

// isAdmin проверяет, является ли пользователь администратором
func (h *MainHandler) isAdmin(userID int64) bool {
	// Здесь можно добавить логику проверки администратора
	// Например, проверка по списку ID из конфигурации
	adminIDs := []int64{123456789} // Замените на реальные ID администраторов

	for _, adminID := range adminIDs {
		if userID == adminID {
			return true
		}
	}
	return false
}
