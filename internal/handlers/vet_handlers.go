package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/internal/database"
	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// VetHandlers содержит обработчики для ветеринарного бота
type VetHandlers struct {
	bot *tgbotapi.BotAPI
	db  *database.Database
}

// NewVetHandlers создает новый экземпляр VetHandlers
func NewVetHandlers(bot *tgbotapi.BotAPI, db *database.Database) *VetHandlers {
	return &VetHandlers{
		bot: bot,
		db:  db,
	}
}

// HandleStart обрабатывает команду /start
func (h *VetHandlers) HandleStart(update tgbotapi.Update) {
	// Создаем или обновляем пользователя
	user := &models.User{
		TelegramID: update.Message.From.ID,
		Username:   update.Message.From.UserName,
		FirstName:  update.Message.From.FirstName,
		LastName:   update.Message.From.LastName,
	}

	err := h.db.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		`🐾 Добро пожаловать в VetBot! 🐾

Я помогу вам найти ветеринарного врача по нужной специализации и расписанию.

Доступные команды:
/start - начать работу
/specializations - показать специализации врачей
/search - поиск врача
/clinics - список клиник
/help - помощь`)

	h.bot.Send(msg)
}

// HandleSpecializations показывает список специализаций
func (h *VetHandlers) HandleSpecializations(update tgbotapi.Update) {
	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		log.Printf("Error getting specializations: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка специализаций")
		h.bot.Send(msg)
		return
	}

	if len(specializations) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Специализации не найдены")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏥 *Доступные специализации:*\n\n")

	for _, spec := range specializations {
		sb.WriteString(fmt.Sprintf("• *%s*", spec.Name))
		if spec.Description != "" {
			sb.WriteString(fmt.Sprintf(" - %s", spec.Description))
		}
		sb.WriteString(fmt.Sprintf(" (/search_%d)\n", spec.ID))
	}

	sb.WriteString("\nНажмите на команду для поиска врачей этой специализации")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleSearch запускает процесс поиска врачей
func (h *VetHandlers) HandleSearch(update tgbotapi.Update) {
	// Создаем клавиатуру с днями недели
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Понедельник", "search_day_1"),
			tgbotapi.NewInlineKeyboardButtonData("Вторник", "search_day_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Среда", "search_day_3"),
			tgbotapi.NewInlineKeyboardButtonData("Четверг", "search_day_4"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пятница", "search_day_5"),
			tgbotapi.NewInlineKeyboardButtonData("Суббота", "search_day_6"),
			tgbotapi.NewInlineKeyboardButtonData("Воскресенье", "search_day_7"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Любой день", "search_day_0"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите день недели для поиска:")
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

// HandleSearchBySpecialization ищет врачей по специализации
func (h *VetHandlers) HandleSearchBySpecialization(update tgbotapi.Update, specializationID int) {
	vets, err := h.db.GetVeterinariansBySpecialization(specializationID)
	if err != nil {
		log.Printf("Error getting veterinarians: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при поиске врачей")
		h.bot.Send(msg)
		return
	}

	if len(vets) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Врачи по выбранной специализации не найдены")
		h.bot.Send(msg)
		return
	}

	spec, err := h.db.GetSpecializationByID(specializationID)
	if err != nil {
		log.Printf("Error getting specialization: %v", err)
	}

	var sb strings.Builder
	if spec != nil {
		sb.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\":*\n\n", spec.Name))
	} else {
		sb.WriteString("👨‍⚕️ *Найденные врачи:*\n\n")
	}

	for i, vet := range vets {
		sb.WriteString(fmt.Sprintf("*%d. %s %s*\n", i+1, vet.FirstName, vet.LastName))
		sb.WriteString(fmt.Sprintf("📞 Телефон: `%s`\n", vet.Phone))
		if vet.Email != "" {
			sb.WriteString(fmt.Sprintf("📧 Email: %s\n", vet.Email))
		}
		if vet.ExperienceYears > 0 {
			sb.WriteString(fmt.Sprintf("💼 Опыт: %d лет\n", vet.ExperienceYears))
		}
		if vet.Description != "" {
			sb.WriteString(fmt.Sprintf("📝 %s\n", vet.Description))
		}

		// Получаем расписание врача
		schedules, err := h.db.GetSchedulesByVetID(vet.ID)
		if err == nil && len(schedules) > 0 {
			sb.WriteString("🕐 Расписание:\n")
			for _, schedule := range schedules {
				dayName := getDayName(schedule.DayOfWeek)
				sb.WriteString(fmt.Sprintf("   - %s: %s-%s\n", dayName, schedule.StartTime, schedule.EndTime))
			}
		}

		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleClinics показывает список клиник
func (h *VetHandlers) HandleClinics(update tgbotapi.Update) {
	clinics, err := h.db.GetAllClinics()
	if err != nil {
		log.Printf("Error getting clinics: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка клиник")
		h.bot.Send(msg)
		return
	}

	if len(clinics) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Клиники не найдены")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏢 *Список клиник:*\n\n")

	for i, clinic := range clinics {
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, clinic.Name))
		sb.WriteString(fmt.Sprintf("📍 Адрес: %s\n", clinic.Address))
		if clinic.Phone != "" {
			sb.WriteString(fmt.Sprintf("📞 Телефон: `%s`\n", clinic.Phone))
		}
		if clinic.WorkingHours != "" {
			sb.WriteString(fmt.Sprintf("🕐 Часы работы: %s\n", clinic.WorkingHours))
		}
		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleHelp показывает справку
func (h *VetHandlers) HandleHelp(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		`🐾 *VetBot - Помощь* 🐾

*Команды:*
/start - Начать работу с ботом
/specializations - Показать все специализации врачей
/search - Поиск врача по расписанию
/clinics - Список всех клиник
/help - Показать эту справку

*Как пользоваться:*
1. Используйте /specializations чтобы увидеть доступные специализации
2. Нажмите на команду поиска рядом с нужной специализацией
3. Или используйте /search для выбора дня недели
4. Бот покажет список врачей с контактами и расписанием

*Примечание:* Телефоны врачей и клиник отображаются в формате, удобном для звонка.`)

	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleCallback обрабатывает inline callback запросы
func (h *VetHandlers) HandleCallback(update tgbotapi.Update) {
	callback := update.CallbackQuery
	data := callback.Data

	if strings.HasPrefix(data, "search_day_") {
		h.handleDaySelection(callback)
		return
	}

	// Отправляем уведомление о том, что действие выполнено
	callbackConfig := tgbotapi.NewCallback(callback.ID, "Обработка...")
	h.bot.Request(callbackConfig)
}

// handleDaySelection обрабатывает выбор дня для поиска
func (h *VetHandlers) handleDaySelection(callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	dayStr := strings.TrimPrefix(data, "search_day_")
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		log.Printf("Error parsing day: %v", err)
		return
	}

	criteria := &models.SearchCriteria{
		DayOfWeek: day,
	}

	vets, err := h.db.FindAvailableVets(criteria)
	if err != nil {
		log.Printf("Error finding vets: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка при поиске врачей")
		h.bot.Request(callbackConfig)
		return
	}

	if len(vets) == 0 {
		dayName := getDayName(day)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
			fmt.Sprintf("Врачи на %s не найдены", dayName))
		h.bot.Send(msg)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
		h.bot.Request(callbackConfig)
		return
	}

	var sb strings.Builder
	dayName := getDayName(day)
	sb.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи, работающие в %s:*\n\n", dayName))

	for i, vet := range vets {
		sb.WriteString(fmt.Sprintf("*%d. %s %s*\n", i+1, vet.FirstName, vet.LastName))
		sb.WriteString(fmt.Sprintf("📞 Телефон: `%s`\n", vet.Phone))

		// Показываем расписание для выбранного дня
		schedules, err := h.db.GetSchedulesByVetID(vet.ID)
		if err == nil {
			for _, schedule := range schedules {
				if schedule.DayOfWeek == day || day == 0 {
					scheduleDayName := getDayName(schedule.DayOfWeek)
					sb.WriteString(fmt.Sprintf("🕐 %s: %s-%s\n", scheduleDayName, schedule.StartTime, schedule.EndTime))
				}
			}
		}
		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
	h.bot.Request(callbackConfig)
}

// getDayName возвращает русское название дня недели
func getDayName(day int) string {
	days := map[int]string{
		1: "понедельник",
		2: "вторник",
		3: "среду",
		4: "четверг",
		5: "пятницу",
		6: "субботу",
		7: "воскресенье",
		0: "любой день",
	}
	return days[day]
}
