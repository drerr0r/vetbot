package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/drerr0r/vetbot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminHandlers содержит обработчики для административных функций
type AdminHandlers struct {
	bot *tgbotapi.BotAPI
	db  *database.Database
}

// NewAdminHandlers создает новый экземпляр AdminHandlers
func NewAdminHandlers(bot *tgbotapi.BotAPI, db *database.Database) *AdminHandlers {
	return &AdminHandlers{
		bot: bot,
		db:  db,
	}
}

// HandleAdmin показывает админскую панель
func (h *AdminHandlers) HandleAdmin(update tgbotapi.Update) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Статистика"),
			tgbotapi.NewKeyboardButton("👥 Пользователи"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏥 Клиники"),
			tgbotapi.NewKeyboardButton("👨‍⚕️ Врачи"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Закрыть админку"),
		),
	)
	keyboard.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"🔧 *Административная панель*\n\nВыберите действие:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// HandleStats показывает статистику бота
func (h *AdminHandlers) HandleStats(update tgbotapi.Update) {
	// Получаем базовую статистику
	userCount, err := h.getUserCount()
	if err != nil {
		log.Printf("Error getting user count: %v", err)
		userCount = 0
	}

	requestCount, err := h.getRequestCount()
	if err != nil {
		log.Printf("Error getting request count: %v", err)
		requestCount = 0
	}

	vetCount, err := h.getVetCount()
	if err != nil {
		log.Printf("Error getting vet count: %v", err)
		vetCount = 0
	}

	statsMsg := fmt.Sprintf(`📊 *Статистика бота*

👥 Пользователей: %d
📞 Запросов: %d
👨‍⚕️ Врачей в базе: %d

*Последние действия:*
- Бот работает стабильно
- Все системы в норме`, userCount, requestCount, vetCount)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, statsMsg)
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleAdminMessage обрабатывает текстовые сообщения в админском режиме
func (h *AdminHandlers) HandleAdminMessage(update tgbotapi.Update) {
	text := update.Message.Text

	switch text {
	case "📊 Статистика":
		h.HandleStats(update)
	case "👥 Пользователи":
		h.handleUsers(update)
	case "🏥 Клиники":
		h.handleClinicsAdmin(update)
	case "👨‍⚕️ Врачи":
		h.handleVetsAdmin(update)
	case "❌ Закрыть админку":
		h.closeAdmin(update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Используйте кнопки админской панели или команду /admin для возврата")
		h.bot.Send(msg)
	}
}

// handleUsers показывает управление пользователями
func (h *AdminHandlers) handleUsers(update tgbotapi.Update) {
	userCount, err := h.getUserCount()
	if err != nil {
		log.Printf("Error getting user count: %v", err)
		userCount = 0
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("👥 *Управление пользователями*\n\nВсего пользователей: %d\n\nДля подробной статистики используйте сторонние инструменты аналитики.", userCount))
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// handleClinicsAdmin показывает управление клиниками
func (h *AdminHandlers) handleClinicsAdmin(update tgbotapi.Update) {
	clinics, err := h.db.GetAllClinics()
	if err != nil {
		log.Printf("Error getting clinics: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка клиник")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏥 *Управление клиниками*\n\n")

	for i, clinic := range clinics {
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, clinic.Name))
		sb.WriteString(fmt.Sprintf("📍 %s\n", clinic.Address))
		sb.WriteString(fmt.Sprintf("📞 %s\n", clinic.Phone))
		sb.WriteString("---\n")
	}

	sb.WriteString("\nДля изменения данных клиник используйте прямые SQL-запросы к базе данных.")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// handleVetsAdmin показывает управление врачами
func (h *AdminHandlers) handleVetsAdmin(update tgbotapi.Update) {
	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		log.Printf("Error getting specializations: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении специализаций")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("👨‍⚕️ *Управление врачами*\n\n")
	sb.WriteString("*Специализации:*\n")

	for _, spec := range specializations {
		vets, err := h.db.GetVeterinariansBySpecialization(spec.ID)
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("• %s: %d врачей\n", spec.Name, len(vets)))
	}

	sb.WriteString("\nДля добавления/изменения данных врачей используйте прямые SQL-запросы к базе данных.")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// closeAdmin закрывает админскую панель
func (h *AdminHandlers) closeAdmin(update tgbotapi.Update) {
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Админская панель закрыта")
	msg.ReplyMarkup = removeKeyboard
	h.bot.Send(msg)
}

// Вспомогательные методы для статистики
func (h *AdminHandlers) getUserCount() (int, error) {
	query := "SELECT COUNT(*) FROM users"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getRequestCount() (int, error) {
	query := "SELECT COUNT(*) FROM user_requests"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getVetCount() (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians WHERE is_active = true"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}
