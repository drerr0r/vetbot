package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler обрабатывает все входящие обновления
type MainHandler struct {
	bot           BotAPI   // Используем интерфейс
	db            Database // Используем интерфейс
	config        *utils.Config
	vetHandlers   *VetHandlers
	adminHandlers *AdminHandlers
}

// NewMainHandler создает новый экземпляр MainHandler
func NewMainHandler(bot BotAPI, db Database, config *utils.Config) *MainHandler {
	return &MainHandler{
		bot:           bot,
		db:            db,
		config:        config,
		vetHandlers:   NewVetHandlers(bot, db),
		adminHandlers: NewAdminHandlers(bot, db, config),
	}
}

// HandleUpdate обрабатывает входящее обновление от Telegram
func (h *MainHandler) HandleUpdate(update tgbotapi.Update) {
	InfoLog.Printf("Received update")

	// Обрабатываем callback queries (нажатия на inline кнопки)
	if update.CallbackQuery != nil {
		InfoLog.Printf("Callback query: %s", update.CallbackQuery.Data)
		h.vetHandlers.HandleCallback(update)
		return
	}

	// Обрабатываем документы (файлы для импорта)
	if update.Message != nil && update.Message.Document != nil {
		InfoLog.Printf("Document received: %s", update.Message.Document.FileName)
		h.handleDocument(update)
		return
	}

	// Игнорируем любые не-text сообщения
	if update.Message == nil {
		InfoLog.Printf("Message is nil")
		return
	}

	if update.Message.Text == "" {
		InfoLog.Printf("Text is empty")
		return
	}

	InfoLog.Printf("Processing message: %s", update.Message.Text)

	// Проверяем, является ли пользователь администратором
	isAdmin := h.isAdmin(update.Message.From.ID)
	InfoLog.Printf("User %d is admin: %t", update.Message.From.ID, isAdmin)

	// Если пользователь администратор и находится в админском режиме, передаем админским хендлерам
	if isAdmin && h.isInAdminMode(update.Message.From.ID) {
		InfoLog.Printf("Redirecting to admin handlers")
		h.adminHandlers.HandleAdminMessage(update)
		return
	}

	// Сначала проверяем команды поиска (/search_1, /search_2 и т.д.)
	if strings.HasPrefix(update.Message.Text, "/search_") {
		InfoLog.Printf("Is search command: %s", update.Message.Text)
		h.handleSearchCommand(update)
		return
	}

	// Затем проверяем обычные команды
	if update.Message.IsCommand() {
		InfoLog.Printf("Is command: %s", update.Message.Command())
		h.handleCommand(update, isAdmin)
		return
	}

	// Обычные текстовые сообщения
	InfoLog.Printf("Is text message: %s", update.Message.Text)
	h.handleTextMessage(update)
}

// handleCommand обрабатывает текстовые команды
func (h *MainHandler) handleCommand(update tgbotapi.Update, isAdmin bool) {
	command := update.Message.Command()
	InfoLog.Printf("Handling command: %s", command)

	switch command {
	case "start":
		InfoLog.Printf("Executing /start")
		h.vetHandlers.HandleStart(update)
	case "specializations":
		InfoLog.Printf("Executing /specializations")
		h.vetHandlers.HandleSpecializations(update)
	case "search":
		InfoLog.Printf("Executing /search")
		h.vetHandlers.HandleSearch(update)
	case "clinics":
		InfoLog.Printf("Executing /clinics")
		h.vetHandlers.HandleClinics(update)
	case "cities":
		InfoLog.Printf("Executing /cities")
		h.vetHandlers.HandleSearchByCity(update)
	case "help":
		InfoLog.Printf("Executing /help")
		h.vetHandlers.HandleHelp(update)
	case "test":
		InfoLog.Printf("Executing /test")
		h.vetHandlers.HandleTest(update)
	case "admin":
		if isAdmin {
			InfoLog.Printf("Executing /admin")
			h.adminHandlers.HandleAdmin(update)
		} else {
			InfoLog.Printf("Admin access denied for user %d", update.Message.From.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав администратора")
			h.bot.Send(msg)
		}
	case "stats":
		if isAdmin {
			InfoLog.Printf("Executing /stats")
			h.adminHandlers.HandleStats(update)
		}
	default:
		InfoLog.Printf("Unknown command: %s", command)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Неизвестная команда. Используйте /help для списка команд")
		h.bot.Send(msg)
	}
}

// handleSearchCommand обрабатывает команды поиска по специализации (/search_1, /search_2 и т.д.)
func (h *MainHandler) handleSearchCommand(update tgbotapi.Update) {
	text := update.Message.Text
	InfoLog.Printf("Handling search command: %s", text)

	if strings.HasPrefix(text, "/search_") {
		specIDStr := strings.TrimPrefix(text, "/search_")
		specID, err := strconv.Atoi(specIDStr)
		if err != nil {
			ErrorLog.Printf("Error parsing specialization ID: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат команды поиска")
			h.bot.Send(msg)
			return
		}
		InfoLog.Printf("Searching for specialization ID: %d", specID)
		h.vetHandlers.HandleSearchBySpecialization(update, specID)
	}
}

// handleTextMessage обрабатывает обычные текстовые сообщения
func (h *MainHandler) handleTextMessage(update tgbotapi.Update) {
	// Для обычных пользователей показываем справку
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Я понимаю только команды. Используйте /help для списка доступных команд.")
	h.bot.Send(msg)
}

// handleDocument обрабатывает загружаемые документы (CSV/Excel для импорта)
func (h *MainHandler) handleDocument(update tgbotapi.Update) {
	fileName := update.Message.Document.FileName

	InfoLog.Printf("Received document: %s", fileName)

	// Проверяем расширение файла
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") &&
		!strings.HasSuffix(strings.ToLower(fileName), ".xlsx") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Поддерживаются только CSV и Excel файлы (.csv, .xlsx)")
		h.bot.Send(msg)
		return
	}

	// Проверяем, является ли пользователь администратором
	if !h.isAdmin(update.Message.From.ID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Импорт данных доступен только администраторам")
		h.bot.Send(msg)
		return
	}

	// Определяем тип импорта по имени файла
	var importType string
	if strings.Contains(strings.ToLower(fileName), "город") {
		importType = "cities"
	} else if strings.Contains(strings.ToLower(fileName), "врач") {
		importType = "veterinarians"
	} else if strings.Contains(strings.ToLower(fileName), "клиник") {
		importType = "clinics"
	} else {
		// Если не удалось определить тип, просим уточнить
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"📥 Файл получен. Укажите тип импорта:\n\n"+
				"• Для городов: файл должен содержать 'город' в названии\n"+
				"• Для врачей: файл должен содержать 'врач' в названии\n"+
				"• Для клиник: файл должен содержать 'клиник' в названии")
		h.bot.Send(msg)
		return
	}

	// Здесь будет логика скачивания и обработки файла
	// Пока просто отправляем сообщение о получении файла
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("📥 Файл '%s' получен для импорта %s.\n\nФункция импорта в разработке.", fileName, importType))
	h.bot.Send(msg)
}

// isAdmin проверяет, является ли пользователь администратором
func (h *MainHandler) isAdmin(userID int64) bool {
	if h.config == nil || len(h.config.AdminIDs) == 0 {
		InfoLog.Printf("Config or AdminIDs is empty")
		return false
	}

	for _, adminID := range h.config.AdminIDs {
		if userID == adminID {
			InfoLog.Printf("User %d found in admin list", userID)
			return true
		}
	}

	InfoLog.Printf("User %d not found in admin list: %v", userID, h.config.AdminIDs)
	return false
}

// isInAdminMode проверяет, находится ли пользователь в админском режиме
func (h *MainHandler) isInAdminMode(userID int64) bool {
	// Защита от nil указателя
	if h.adminHandlers == nil {
		InfoLog.Printf("Admin handlers is nil for user %d", userID)
		return false
	}

	// Проверяем состояние админской сессии
	if state, exists := h.adminHandlers.adminState[userID]; exists {
		return state != ""
	}
	return false
}
