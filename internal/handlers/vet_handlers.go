package handlers

import (
	"fmt"
	"html"
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
	log.Printf("HandleStart called")

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

	log.Printf("Sending start message")
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	} else {
		log.Printf("Start message sent successfully")
	}
}

// HandleSpecializations показывает список специализаций (HTML версия)
func (h *VetHandlers) HandleSpecializations(update tgbotapi.Update) {
	log.Printf("HandleSpecializations called")

	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		log.Printf("Error getting specializations: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка специализаций")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d specializations", len(specializations))

	if len(specializations) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Специализации не найдены")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏥 <b>Доступные специализации:</b>\n\n")

	for _, spec := range specializations {
		log.Printf("Specialization: %s (ID: %d)", spec.Name, spec.ID)

		sb.WriteString(fmt.Sprintf("• <b>%s</b>", html.EscapeString(spec.Name)))
		if spec.Description != "" {
			sb.WriteString(fmt.Sprintf(" - %s", html.EscapeString(spec.Description)))
		}
		sb.WriteString(fmt.Sprintf(" (/search_%d)\n", spec.ID))
	}

	sb.WriteString("\nНажмите на команду для поиска врачей этой специализации")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "HTML"

	log.Printf("Sending specializations message with HTML formatting")
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending specializations message with HTML: %v", err)

		// Попробуем отправить без форматирования
		log.Printf("Trying without formatting")
		msg2 := tgbotapi.NewMessage(update.Message.Chat.ID,
			"🏥 Доступные специализации:\n\n"+
				"• Терапевт (/search_1)\n"+
				"• Хирург (/search_2)\n"+
				"• Стоматолог (/search_3)\n"+
				"• Дерматолог (/search_4)\n"+
				"• Офтальмолог (/search_5)\n"+
				"• Кардиолог (/search_6)\n"+
				"• Ортопед (/search_7)\n\n"+
				"Нажмите на команду для поиска врачей этой специализации")
		h.bot.Send(msg2)
	} else {
		log.Printf("Specializations message sent successfully with HTML")
	}
}

// HandleSearch запускает процесс поиска врачей
func (h *VetHandlers) HandleSearch(update tgbotapi.Update) {
	log.Printf("HandleSearch called")

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

	log.Printf("Sending search message")
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending search message: %v", err)
	} else {
		log.Printf("Search message sent successfully")
	}
}

// HandleSearchBySpecialization ищет врачей по специализации
func (h *VetHandlers) HandleSearchBySpecialization(update tgbotapi.Update, specializationID int) {
	log.Printf("HandleSearchBySpecialization called with ID: %d", specializationID)

	vets, err := h.db.GetVeterinariansBySpecialization(specializationID)
	if err != nil {
		log.Printf("Error getting veterinarians: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при поиске врачей")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d veterinarians for specialization ID: %d", len(vets), specializationID)

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
		sb.WriteString(fmt.Sprintf("👨‍⚕️ <b>Врачи по специализации \"%s\":</b>\n\n", html.EscapeString(spec.Name)))
	} else {
		sb.WriteString("👨‍⚕️ <b>Найденные врачи:</b>\n\n")
	}

	for i, vet := range vets {
		sb.WriteString(fmt.Sprintf("<b>%d. %s %s</b>\n", i+1, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
		sb.WriteString(fmt.Sprintf("📞 Телефон: <code>%s</code>\n", html.EscapeString(vet.Phone)))

		// Проверяем email (может быть NULL)
		if vet.Email.Valid && vet.Email.String != "" {
			sb.WriteString(fmt.Sprintf("📧 Email: %s\n", html.EscapeString(vet.Email.String)))
		}

		// Проверяем опыт работы (может быть NULL)
		if vet.ExperienceYears.Valid {
			sb.WriteString(fmt.Sprintf("💼 Опыт: %d лет\n", vet.ExperienceYears.Int64))
		}

		// Проверяем описание (может быть NULL)
		if vet.Description.Valid && vet.Description.String != "" {
			sb.WriteString(fmt.Sprintf("📝 %s\n", html.EscapeString(vet.Description.String)))
		}

		// Показываем специализации врача
		if len(vet.Specializations) > 0 {
			sb.WriteString("🎯 Специализации: ")
			specNames := make([]string, len(vet.Specializations))
			for j, spec := range vet.Specializations {
				specNames[j] = html.EscapeString(spec.Name)
			}
			sb.WriteString(strings.Join(specNames, ", "))
			sb.WriteString("\n")
		}

		// Получаем расписание врача
		schedules, err := h.db.GetSchedulesByVetID(vet.ID)
		if err == nil && len(schedules) > 0 {
			sb.WriteString("🕐 Расписание:\n")
			for _, schedule := range schedules {
				dayName := getDayName(schedule.DayOfWeek)
				// Проверяем что время не пустое
				startTime := schedule.StartTime
				endTime := schedule.EndTime
				if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
					sb.WriteString(fmt.Sprintf("   - %s: %s-%s", dayName, startTime, endTime))
					if schedule.Clinic != nil && schedule.Clinic.Name != "" {
						sb.WriteString(fmt.Sprintf(" (%s)", html.EscapeString(schedule.Clinic.Name)))
					}
					sb.WriteString("\n")
				}
			}
		}

		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "HTML"

	log.Printf("Sending search results message")
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending search results message with HTML: %v", err)

		// Попробуем без HTML
		msg2 := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
		msg2.ParseMode = ""
		h.bot.Send(msg2)
	} else {
		log.Printf("Search results message sent successfully")
	}
}

// HandleClinics показывает список клиник
func (h *VetHandlers) HandleClinics(update tgbotapi.Update) {
	log.Printf("HandleClinics called")

	clinics, err := h.db.GetAllClinics()
	if err != nil {
		log.Printf("Error getting clinics: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка клиник")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d clinics", len(clinics))

	if len(clinics) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Клиники не найдены")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏢 <b>Список клиник:</b>\n\n")

	for i, clinic := range clinics {
		sb.WriteString(fmt.Sprintf("<b>%d. %s</b>\n", i+1, html.EscapeString(clinic.Name)))
		sb.WriteString(fmt.Sprintf("📍 Адрес: %s\n", html.EscapeString(clinic.Address)))

		// Проверяем телефон (может быть NULL)
		if clinic.Phone.Valid && clinic.Phone.String != "" {
			sb.WriteString(fmt.Sprintf("📞 Телефон: <code>%s</code>\n", html.EscapeString(clinic.Phone.String)))
		}

		// Проверяем часы работы (может быть NULL)
		if clinic.WorkingHours.Valid && clinic.WorkingHours.String != "" {
			sb.WriteString(fmt.Sprintf("🕐 Часы работы: %s\n", html.EscapeString(clinic.WorkingHours.String)))
		}

		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "HTML"

	log.Printf("Sending clinics message")
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending clinics message with HTML: %v", err)

		// Попробуем без HTML
		msg2 := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
		msg2.ParseMode = ""
		h.bot.Send(msg2)
	} else {
		log.Printf("Clinics message sent successfully")
	}
}

// HandleHelp показывает справку
func (h *VetHandlers) HandleHelp(update tgbotapi.Update) {
	log.Printf("HandleHelp called")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		`🐾 <b>VetBot - Помощь</b> 🐾

<b>Команды:</b>
/start - Начать работу с ботом
/specializations - Показать все специализации врачей
/search - Поиск врача по расписанию
/clinics - Список всех клиник
/help - Показать эту справку

<b>Как пользоваться:</b>
1. Используйте /specializations чтобы увидеть доступные специализации
2. Нажмите на команду поиска рядом с нужной специализацией
3. Или используйте /search для выбора дня недели
4. Бот покажет список врачей с контактами и расписанием

<b>Примечание:</b> Телефоны врачей и клиник отображаются в формате, удобном для звонка.`)

	msg.ParseMode = "HTML"

	log.Printf("Sending help message")
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending help message with HTML: %v", err)

		// Попробуем без HTML
		msg2 := tgbotapi.NewMessage(update.Message.Chat.ID,
			`🐾 VetBot - Помощь 🐾

Команды:
/start - Начать работу с ботом
/specializations - Показать все специализации врачей
/search - Поиск врача по расписанию
/clinics - Список всех клиник
/help - Показать эту справку

Как пользоваться:
1. Используйте /specializations чтобы увидеть доступные специализации
2. Нажмите на команду поиска рядом с нужной специализацией
3. Или используйте /search для выбора дня недели
4. Бот покажет список врачей с контактами и расписанием

Примечание: Телефоны врачей и клиник отображаются в формате, удобном для звонка.`)
		h.bot.Send(msg2)
	} else {
		log.Printf("Help message sent successfully")
	}
}

// HandleTest для тестирования
func (h *VetHandlers) HandleTest(update tgbotapi.Update) {
	log.Printf("HandleTest called")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Тестовое сообщение: бот работает!")
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending test message: %v", err)
	} else {
		log.Printf("Test message sent successfully")
	}
}

// HandleCallback обрабатывает inline callback запросы
func (h *VetHandlers) HandleCallback(update tgbotapi.Update) {
	log.Printf("HandleCallback called")

	callback := update.CallbackQuery
	data := callback.Data

	log.Printf("Callback data: %s", data)

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
	log.Printf("handleDaySelection called")

	data := callback.Data
	dayStr := strings.TrimPrefix(data, "search_day_")
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		log.Printf("Error parsing day: %v", err)
		return
	}

	log.Printf("Searching for day: %d", day)

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

	log.Printf("Found %d vets for day %d", len(vets), day)

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
	sb.WriteString(fmt.Sprintf("👨‍⚕️ <b>Врачи, работающие в %s:</b>\n\n", dayName))

	for i, vet := range vets {
		sb.WriteString(fmt.Sprintf("<b>%d. %s %s</b>\n", i+1, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
		sb.WriteString(fmt.Sprintf("📞 Телефон: <code>%s</code>\n", html.EscapeString(vet.Phone)))

		// Проверяем email (может быть NULL)
		if vet.Email.Valid && vet.Email.String != "" {
			sb.WriteString(fmt.Sprintf("📧 Email: %s\n", html.EscapeString(vet.Email.String)))
		}

		// Проверяем опыт работы (может быть NULL)
		if vet.ExperienceYears.Valid {
			sb.WriteString(fmt.Sprintf("💼 Опыт: %d лет\n", vet.ExperienceYears.Int64))
		}

		// Показываем расписание для выбранного дня
		schedules, err := h.db.GetSchedulesByVetID(vet.ID)
		if err == nil {
			for _, schedule := range schedules {
				if schedule.DayOfWeek == day || day == 0 {
					scheduleDayName := getDayName(schedule.DayOfWeek)
					// Проверяем что время не пустое
					startTime := schedule.StartTime
					endTime := schedule.EndTime
					if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
						sb.WriteString(fmt.Sprintf("🕐 %s: %s-%s", scheduleDayName, startTime, endTime))
						if schedule.Clinic != nil && schedule.Clinic.Name != "" {
							sb.WriteString(fmt.Sprintf(" (%s)", html.EscapeString(schedule.Clinic.Name)))
						}
						sb.WriteString("\n")
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, sb.String())
	msg.ParseMode = "HTML"

	log.Printf("Sending day search results")
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending day search results with HTML: %v", err)

		// Попробуем без HTML
		msg2 := tgbotapi.NewMessage(callback.Message.Chat.ID, sb.String())
		msg2.ParseMode = ""
		h.bot.Send(msg2)
	}

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
