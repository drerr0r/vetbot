package handlers

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// VetHandlers содержит обработчики для ветеринарного бота
type VetHandlers struct {
	bot            BotAPI
	db             Database
	stateManager   *StateManager
	reviewHandlers *ReviewHandlers
}

// NewVetHandlers создает новый экземпляр VetHandlers
func NewVetHandlers(bot BotAPI, db Database, adminIDs []int64, stateManager *StateManager) *VetHandlers {
	return &VetHandlers{
		bot:            bot,
		db:             db,
		stateManager:   stateManager,
		reviewHandlers: NewReviewHandlers(bot, db, adminIDs, stateManager),
	}
}

// HandleStart обрабатывает команду /start
func (h *VetHandlers) HandleStart(update tgbotapi.Update) {
	InfoLog.Printf("HandleStart called")

	// Создаем или обновляем пользователя
	user := &models.User{
		TelegramID: update.Message.From.ID,
		Username:   update.Message.From.UserName,
		FirstName:  update.Message.From.FirstName,
		LastName:   update.Message.From.LastName,
	}

	err := h.db.CreateUser(user)
	if err != nil {
		ErrorLog.Printf("Error creating user: %v", err)
	}

	// Очищаем историю навигации
	h.stateManager.ClearHistory(update.Message.From.ID)

	// Создаем главное меню с inline-кнопками
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Поиск по специализациям", "main_specializations"),
			tgbotapi.NewInlineKeyboardButtonData("🕐 Поиск по времени", "main_time"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏥 Поиск по клиникам", "main_clinics"),
			tgbotapi.NewInlineKeyboardButtonData("🏙️ Поиск по городу", "main_city"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		`🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Выберите способ поиска:

*Используйте кнопки ниже для быстрой навигации:*`)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"

	// Добавляем постоянную клавиатуру
	persistentKeyboard := h.createPersistentKeyboard()
	msg.ReplyMarkup = persistentKeyboard

	InfoLog.Printf("Sending start message with persistent keyboard")
	_, err = h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending start message: %v", err)
	} else {
		InfoLog.Printf("Start message sent successfully")
	}
}

// HandleSpecializations показывает список специализаций с улучшенным интерфейсом
func (h *VetHandlers) HandleSpecializations(update tgbotapi.Update) {
	InfoLog.Printf("HandleSpecializations called")

	// Сохраняем текущее состояние в историю
	if update.Message != nil {
		h.stateManager.PushState(update.Message.From.ID, "main_menu")
	} else if update.CallbackQuery != nil {
		h.stateManager.PushState(update.CallbackQuery.From.ID, "main_menu")
	}

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		// Отвечаем на callback query чтобы убрать "часики" у кнопки
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		ErrorLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		ErrorLog.Printf("Error getting specializations: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении списка специализаций")
		h.bot.Send(msg)
		return
	}

	InfoLog.Printf("Found %d specializations", len(specializations))

	if len(specializations) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Специализации не найдены")
		h.bot.Send(msg)
		return
	}

	// Создаем кнопки для специализаций (максимум 3 в ряду)
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i, spec := range specializations {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			spec.Name,
			fmt.Sprintf("search_spec_%d", spec.ID),
		)
		currentRow = append(currentRow, btn)

		// Создаем новый ряд после каждых 3 кнопок или в конце
		if (i+1)%3 == 0 || i == len(specializations)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Добавляем кнопку "Назад"
	backRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "main_menu"),
	)
	keyboardRows = append(keyboardRows, backRow)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	msg := tgbotapi.NewMessage(chatID,
		"🏥 *Выберите специализацию врача:*\n\nНажмите на кнопку с нужной специализацией для поиска врачей.")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	InfoLog.Printf("Sending specializations menu to chat %d", chatID)
	_, err = h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending specializations menu: %v", err)
	}
}

// HandleSearch показывает меню поиска по времени
func (h *VetHandlers) HandleSearch(update tgbotapi.Update) {
	InfoLog.Printf("HandleSearch called")

	// Сохраняем текущее состояние в историю
	if update.Message != nil {
		h.stateManager.PushState(update.Message.From.ID, "main_menu")
	} else if update.CallbackQuery != nil {
		h.stateManager.PushState(update.CallbackQuery.From.ID, "main_menu")
	}

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		ErrorLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	// Создаем клавиатуру с днями недели и кнопкой "Назад"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Понедельник", "search_day_1"),
			tgbotapi.NewInlineKeyboardButtonData("Вторник", "search_day_2"),
			tgbotapi.NewInlineKeyboardButtonData("Среда", "search_day_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Четверг", "search_day_4"),
			tgbotapi.NewInlineKeyboardButtonData("Пятница", "search_day_5"),
			tgbotapi.NewInlineKeyboardButtonData("Суббота", "search_day_6"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Воскресенье", "search_day_7"),
			tgbotapi.NewInlineKeyboardButtonData("Любой день", "search_day_0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID,
		"🕐 *Выберите день недели для поиска:*\n\nЯ покажу врачей, работающих в выбранный день.")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	InfoLog.Printf("Sending search by time menu to chat %d", chatID)
	_, err := h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending search menu: %v", err)
	}
}

// HandleClinics показывает меню клиник
func (h *VetHandlers) HandleClinics(update tgbotapi.Update) {
	InfoLog.Printf("HandleClinics called")

	// Сохраняем текущее состояние в историю
	if update.Message != nil {
		h.stateManager.PushState(update.Message.From.ID, "main_menu")
	} else if update.CallbackQuery != nil {
		h.stateManager.PushState(update.CallbackQuery.From.ID, "main_menu")
	}

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		ErrorLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	clinics, err := h.db.GetAllClinics()
	if err != nil {
		ErrorLog.Printf("Error getting clinics: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении списка клиник")
		h.bot.Send(msg)
		return
	}

	InfoLog.Printf("Found %d clinics", len(clinics))

	if len(clinics) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Клиники не найдены")
		h.bot.Send(msg)
		return
	}

	// Создаем кнопки для клиник
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i, clinic := range clinics {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			clinic.Name,
			fmt.Sprintf("search_clinic_%d", clinic.ID),
		)
		currentRow = append(currentRow, btn)

		// Создаем новый ряд после каждых 2 кнопок или в конце
		if (i+1)%2 == 0 || i == len(clinics)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Добавляем кнопку "Назад"
	backRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "main_menu"),
	)
	keyboardRows = append(keyboardRows, backRow)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	msg := tgbotapi.NewMessage(chatID,
		"🏥 *Выберите клинику:*\n\nЯ покажу врачей, работающих в выбранной клинике.")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	InfoLog.Printf("Sending clinics menu to chat %d", chatID)
	_, err = h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending clinics menu: %v", err)
	}
}

// HandleSearchByCity показывает меню поиска по городам
func (h *VetHandlers) HandleSearchByCity(update tgbotapi.Update) {
	InfoLog.Printf("HandleSearchByCity called")

	// Сохраняем текущее состояние в историю
	if update.Message != nil {
		h.stateManager.PushState(update.Message.From.ID, "main_menu")
	} else if update.CallbackQuery != nil {
		h.stateManager.PushState(update.CallbackQuery.From.ID, "main_menu")
	}

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		ErrorLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	// Получаем список городов
	cities, err := h.db.GetAllCities()
	if err != nil {
		ErrorLog.Printf("Error getting cities: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении списка городов")
		h.bot.Send(msg)
		return
	}

	if len(cities) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Городы не найдены в базе данных")
		h.bot.Send(msg)
		return
	}

	// Создаем кнопки для городов
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i, city := range cities {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s (%s)", city.Name, city.Region),
			fmt.Sprintf("search_city_%d", city.ID),
		)
		currentRow = append(currentRow, btn)

		// Создаем новый ряд после каждых 2 кнопок или в конце
		if (i+1)%2 == 0 || i == len(cities)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Добавляем кнопку "Назад"
	backRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "main_menu"),
	)
	keyboardRows = append(keyboardRows, backRow)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	msg := tgbotapi.NewMessage(chatID,
		"🏙️ *Выберите город для поиска врачей:*\n\nЯ покажу врачей, работающих в выбранном городе.")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	InfoLog.Printf("Sending cities menu to chat %d", chatID)
	_, err = h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending cities menu: %v", err)
	}
}

// HandleHelp показывает справку с кнопкой "Назад"
func (h *VetHandlers) HandleHelp(update tgbotapi.Update) {
	InfoLog.Printf("HandleHelp called")

	// Сохраняем текущее состояние в историю
	if update.Message != nil {
		h.stateManager.PushState(update.Message.From.ID, "main_menu")
	} else if update.CallbackQuery != nil {
		h.stateManager.PushState(update.CallbackQuery.From.ID, "main_menu")
	}

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		InfoLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	// keyboard := tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "main_menu"),
	// 	),
	// )

	helpText := `🐾 *VetBot - Помощь* 🐾

*Основные функции:*
• 🔍 *Поиск по специализациям* - найти врача по направлению
• 🕐 *Поиск по времени* - найти врача по дню недели
• 🏥 *Поиск по клиникам* - найти врачей в конкретной клинике
• 🏙️ *Поиск по городу* - найти врачей в определенном городе

*Как пользоваться:*
1. Выберите способ поиска из главного меню
2. Нажмите на нужную кнопку (специализация, день, клиника или город)
3. Бот покажет список врачей с контактами и расписанием

*Команды:*
/start - Главное меню
/cities - Поиск по городам
/help - Эта справка`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	// НОВОЕ: Добавляем постоянную клавиатуру вместо inline-кнопок
	persistentKeyboard := h.createPersistentKeyboard()
	msg.ReplyMarkup = persistentKeyboard

	InfoLog.Printf("Sending help message to chat %d", chatID)
	_, err := h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending help message: %v", err)
	}
}

// createPersistentKeyboard создает постоянную Reply-клавиатуру
func (h *VetHandlers) createPersistentKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏠 Главное меню"),
			tgbotapi.NewKeyboardButton("ℹ️ Помощь"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)
}

// HandleSearchBySpecialization ищет врачей по специализации с кнопками отзывов
func (h *VetHandlers) HandleSearchBySpecialization(update tgbotapi.Update, specializationID int) {
	InfoLog.Printf("HandleSearchBySpecialization called with ID: %d", specializationID)

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		// Отвечаем на callback query чтобы убрать "часики" у кнопки
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		ErrorLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	vets, err := h.db.GetVeterinariansBySpecialization(specializationID)
	if err != nil {
		ErrorLog.Printf("Error getting veterinarians: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при поиске врачей")
		h.bot.Send(msg)
		return
	}

	InfoLog.Printf("Found %d veterinarians for specialization ID: %d", len(vets), specializationID)

	spec, err := h.db.GetSpecializationByID(specializationID)
	if err != nil {
		ErrorLog.Printf("Error getting specialization: %v", err)
	}

	// Создаем клавиатуру с кнопкой "Назад"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 К специализациям", "main_specializations"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if len(vets) == 0 {
		var specName string
		if spec != nil {
			specName = spec.Name
		} else {
			specName = "выбранной специализации"
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\" не найдены*\n\nПопробуйте выбрать другую специализацию.", specName))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
		return
	}

	// Отправляем заголовок
	var specName string
	if spec != nil {
		specName = spec.Name
	} else {
		specName = "специализации"
	}

	// Если врачей больше 5, используем компактный формат с разбивкой по сообщениям
	if len(vets) > 5 {
		headerMsg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\":*\n\nНайдено врачей: %d\n\n*Показаны в компактном формате. Нажмите \"Подробнее\" для полной информации:*", specName, len(vets)))
		headerMsg.ParseMode = "Markdown"
		h.bot.Send(headerMsg)

		// Используем компактный формат с кнопкой "Подробнее"
		for i, vet := range vets {
			err := h.sendVetWithDetailsButton(chatID, vet, i+1)
			if err != nil {
				ErrorLog.Printf("Error sending vet info: %v", err)
			}
		}
	} else {
		// Если врачей мало, показываем детальную информацию с кнопками отзывов
		headerMsg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\":*\n\nНайдено врачей: %d\n\nВыберите врача для просмотра отзывов:", specName, len(vets)))
		headerMsg.ParseMode = "Markdown"
		h.bot.Send(headerMsg)

		// Отправляем каждого врача с детальной информацией и кнопками отзывов
		for i, vet := range vets {
			err := h.sendVetWithSpecializationDetailsAndReviews(chatID, vet, i+1)
			if err != nil {
				ErrorLog.Printf("Error sending vet info: %v", err)
			}
		}
	}
}

// sendVetWithSpecializationDetailsAndReviews отправляет врача с детальной информацией и кнопками отзывов для специализаций
func (h *VetHandlers) sendVetWithSpecializationDetailsAndReviews(chatID int64, vet *models.Veterinarian, index int) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", index, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
	sb.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", html.EscapeString(vet.Phone)))

	if vet.Email.Valid && vet.Email.String != "" {
		sb.WriteString(fmt.Sprintf("📧 *Email:* %s\n", html.EscapeString(vet.Email.String)))
	}

	if vet.ExperienceYears.Valid {
		sb.WriteString(fmt.Sprintf("💼 *Опыт:* %d лет\n", vet.ExperienceYears.Int64))
	}

	// Информация о городе
	if vet.City != nil {
		sb.WriteString(fmt.Sprintf("🏙️ *Город:* %s\n", vet.City.Name))
	}

	// Расписание врача (первые 3 дня)
	schedules, err := h.db.GetSchedulesByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil && len(schedules) > 0 {
		sb.WriteString("🕐 *Ближайшее расписание:*\n")
		displayed := 0
		for _, schedule := range schedules {
			if displayed >= 3 { // Показываем только 3 ближайших записи
				break
			}
			scheduleDayName := getDayName(schedule.DayOfWeek)
			startTime := schedule.StartTime
			endTime := schedule.EndTime
			if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
				sb.WriteString(fmt.Sprintf("• %s: %s-%s", scheduleDayName, startTime, endTime))
				if schedule.Clinic != nil && schedule.Clinic.Name != "" {
					sb.WriteString(fmt.Sprintf(" (%s)", html.EscapeString(schedule.Clinic.Name)))
				}
				sb.WriteString("\n")
				displayed++
			}
		}
	}

	// Создаем клавиатуру с кнопками отзывов
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", models.GetVetIDAsIntOrZero(vet))),
		),
	)

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err = h.bot.Send(msg)
	return err
}

// HandleSearchByClinic ищет врачей по клинике
func (h *VetHandlers) HandleSearchByClinic(update tgbotapi.Update, clinicID int) {
	InfoLog.Printf("HandleSearchByClinic called with ID: %d", clinicID)

	var chatID int64

	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		ErrorLog.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	// Получаем информацию о клинике
	clinic, err := h.db.GetClinicByID(clinicID)
	if err != nil {
		ErrorLog.Printf("Error getting clinic: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении информации о клинике")
		h.bot.Send(msg)
		return
	}

	// Получаем врачей клиники через расписание
	criteria := &models.SearchCriteria{
		ClinicID: clinicID,
	}
	vets, err := h.db.FindAvailableVets(criteria)
	if err != nil {
		ErrorLog.Printf("Error finding vets by clinic: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при поиске врачей")
		h.bot.Send(msg)
		return
	}

	InfoLog.Printf("Found %d veterinarians for clinic ID: %d", len(vets), clinicID)

	// Клавиатура с кнопками навигации
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 К клиникам", "main_clinics"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if len(vets) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("🏥 *Врачи в клинике \"%s\" не найдены*\n\nПопробуйте выбрать другую клинику.", clinic.Name))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
		return
	}

	// Сначала отправляем информацию о клинике с кнопкой отзыва
	err = h.sendClinicInfoWithReviewButton(chatID, clinic, len(vets))
	if err != nil {
		ErrorLog.Printf("Error sending clinic info: %v", err)
	}

	// Затем отправляем компактный список врачей
	headerMsg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("👨‍⚕️ *Врачи клиники \"%s\":*\n\nНайдено врачей: %d\n\n", html.EscapeString(clinic.Name), len(vets)))
	headerMsg.ParseMode = "Markdown"
	h.bot.Send(headerMsg)

	// Отправляем каждого врача в компактном формате с кнопкой "Подробнее"
	for i, vet := range vets {
		err := h.sendCompactVetInfo(chatID, vet, i+1, clinicID)
		if err != nil {
			ErrorLog.Printf("Error sending vet info: %v", err)
		}
	}
}

// sendClinicInfoWithReviewButton отправляет информацию о клинике с кнопкой отзыва
func (h *VetHandlers) sendClinicInfoWithReviewButton(chatID int64, clinic *models.Clinic, vetCount int) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🏥 *%s*\n\n", html.EscapeString(clinic.Name)))
	sb.WriteString(fmt.Sprintf("📍 *Адрес:* %s\n", html.EscapeString(clinic.Address)))

	if clinic.Phone.Valid && clinic.Phone.String != "" {
		sb.WriteString(fmt.Sprintf("📞 *Телефон:* %s\n", html.EscapeString(clinic.Phone.String)))
	}

	if clinic.WorkingHours.Valid && clinic.WorkingHours.String != "" {
		sb.WriteString(fmt.Sprintf("🕐 *Часы работы:* %s\n", html.EscapeString(clinic.WorkingHours.String)))
	}

	if clinic.MetroStation.Valid && clinic.MetroStation.String != "" {
		sb.WriteString(fmt.Sprintf("🚇 *Метро:* %s\n", html.EscapeString(clinic.MetroStation.String)))
	}

	if clinic.District.Valid && clinic.District.String != "" {
		sb.WriteString(fmt.Sprintf("🏘️ *Район:* %s\n", html.EscapeString(clinic.District.String)))
	}

	sb.WriteString(fmt.Sprintf("👨‍⚕️ *Количество врачей:* %d\n", vetCount))

	// Клавиатура с кнопкой отзыва о клинике
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Оставить отзыв о клинике", fmt.Sprintf("add_clinic_review_%d", clinic.ID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

// sendCompactVetInfo отправляет компактную информацию о враче с кнопкой "Подробнее"
func (h *VetHandlers) sendCompactVetInfo(chatID int64, vet *models.Veterinarian, index int, clinicID int) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", index, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
	sb.WriteString(fmt.Sprintf("📞 `%s`", html.EscapeString(vet.Phone)))

	// Специализации (первые 2)
	specs, err := h.db.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil && len(specs) > 0 {
		sb.WriteString(" 🎯 ")
		specNames := make([]string, 0)
		for j, spec := range specs {
			if j >= 2 { // Ограничиваем 2 специализациями
				break
			}
			specNames = append(specNames, html.EscapeString(spec.Name))
		}
		sb.WriteString(strings.Join(specNames, ", "))
		if len(specs) > 2 {
			sb.WriteString(fmt.Sprintf(" (+%d)", len(specs)-2))
		}
	}

	// Рейтинг (если есть)
	stats, err := h.db.GetReviewStats(models.GetVetIDAsIntOrZero(vet))
	if err == nil && stats.ApprovedReviews > 0 {
		sb.WriteString(fmt.Sprintf(" ⭐ %.1f/5", stats.AverageRating))
	}

	// Используем clinicID для показа расписания в этой клинике
	schedules, err := h.db.GetSchedulesByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil {
		// Ищем ближайший рабочий день в этой клинике
		for _, schedule := range schedules {
			if schedule.ClinicID == clinicID && schedule.IsAvailable {
				scheduleDayName := getDayName(schedule.DayOfWeek)
				startTime := schedule.StartTime
				endTime := schedule.EndTime
				if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
					sb.WriteString(fmt.Sprintf(" 🕐 %s %s-%s", scheduleDayName, startTime, endTime))
					break // Показываем только первый найденный день
				}
			}
		}
	}

	sb.WriteString("\n")

	// Клавиатура с кнопками (используем clinicID в callback)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Подробнее", fmt.Sprintf("vet_details_clinic_%d_%d", models.GetVetIDAsIntOrZero(vet), clinicID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
		),
	)

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err = h.bot.Send(msg)
	return err
}

// HandleVetDetailsFromClinic показывает детальную информацию о враче в контексте клиники
func (h *VetHandlers) HandleVetDetailsFromClinic(chatID int64, vetID int, clinicID int, messageID int) error {
	InfoLog.Printf("HandleVetDetailsFromClinic called for vet ID: %d, clinic ID: %d", vetID, clinicID)

	// Получаем полную информацию о враче
	vet, err := h.db.GetVeterinarianWithDetails(vetID)
	if err != nil {
		ErrorLog.Printf("Error getting vet details: %v", err)
		return fmt.Errorf("failed to get vet details: %v", err)
	}

	// Получаем информацию о клинике
	clinic, err := h.db.GetClinicByID(clinicID)
	if err != nil {
		ErrorLog.Printf("Error getting clinic: %v", err)
		clinic = &models.Clinic{Name: "Неизвестная клиника"}
	}

	// Формируем сообщение с полной информацией
	message := h.formatVeterinarianDetailsForClinic(vet, clinic)

	// Создаем клавиатуру с закрепленными кнопками
	replyMarkup := h.createVetDetailsKeyboardForClinic(vetID, clinicID)

	// Если есть предыдущее сообщение, редактируем его
	if messageID != 0 {
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, message)
		editMsg.ParseMode = "Markdown"
		editMsg.ReplyMarkup = &replyMarkup
		_, err = h.bot.Send(editMsg)
	} else {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = replyMarkup
		_, err = h.bot.Send(msg)
	}

	return err
}

// formatVeterinarianDetailsForClinic форматирует детальную информацию о враче для конкретной клиники
func (h *VetHandlers) formatVeterinarianDetailsForClinic(vet *models.Veterinarian, clinic *models.Clinic) string {
	var message strings.Builder

	message.WriteString("🐾 *Детальная информация о враче*\n\n")

	// Основная информация
	message.WriteString(fmt.Sprintf("👨‍⚕️ *%s %s*\n", vet.FirstName, vet.LastName))

	// Получаем статистику отзывов
	stats, err := h.db.GetReviewStats(models.GetVetIDAsIntOrZero(vet))
	if err == nil {
		if stats.ApprovedReviews > 0 {
			message.WriteString(fmt.Sprintf("⭐ *Рейтинг:* %.1f/5 (%d отзывов)\n", stats.AverageRating, stats.ApprovedReviews))
		} else {
			message.WriteString("⭐ *Рейтинг:* пока нет отзывов\n")
		}
	}

	if vet.Phone != "" {
		message.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", vet.Phone))
	}

	if vet.Email.Valid && vet.Email.String != "" {
		message.WriteString(fmt.Sprintf("📧 *Email:* %s\n", vet.Email.String))
	}

	if vet.ExperienceYears.Valid {
		message.WriteString(fmt.Sprintf("⏳ *Опыт работы:* %d лет\n", vet.ExperienceYears.Int64))
	}

	if vet.Description.Valid && vet.Description.String != "" {
		message.WriteString(fmt.Sprintf("📝 *Описание:* %s\n", vet.Description.String))
	}

	// Информация о городе
	if vet.City != nil {
		message.WriteString(fmt.Sprintf("🏙️ *Город:* %s", vet.City.Name))
		if vet.City.Region != "" {
			message.WriteString(fmt.Sprintf(" (%s)", vet.City.Region))
		}
		message.WriteString("\n")
	}

	// Специализации
	if len(vet.Specializations) > 0 {
		message.WriteString("🎯 *Специализации:* ")
		specNames := make([]string, len(vet.Specializations))
		for i, spec := range vet.Specializations {
			specNames[i] = spec.Name
		}
		message.WriteString(strings.Join(specNames, ", "))
		message.WriteString("\n")
	}

	// Расписание в конкретной клинике
	message.WriteString(fmt.Sprintf("\n🏥 *Расписание в клинике \"%s\":*\n", clinic.Name))

	hasSchedule := false
	schedules, err := h.db.GetSchedulesByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil {
		// Группируем по дням для этой клиники
		daysMap := make(map[int][]string)
		for _, schedule := range schedules {
			if schedule.ClinicID == clinic.ID {
				timeSlot := fmt.Sprintf("%s-%s", schedule.StartTime, schedule.EndTime)
				daysMap[schedule.DayOfWeek] = append(daysMap[schedule.DayOfWeek], timeSlot)
			}
		}

		if len(daysMap) > 0 {
			hasSchedule = true
			// Сортируем дни недели по порядку
			for day := 1; day <= 7; day++ {
				if timeSlots, exists := daysMap[day]; exists && len(timeSlots) > 0 {
					dayName := getDayName(day)
					message.WriteString(fmt.Sprintf("• %s: %s\n", dayName, strings.Join(timeSlots, ", ")))
				}
			}
		}
	}

	if !hasSchedule {
		message.WriteString("📅 Расписание не указано\n")
	}

	return message.String()
}

// createVetDetailsKeyboardForClinic создает клавиатуру для детального просмотра врача в контексте клиники
func (h *VetHandlers) createVetDetailsKeyboardForClinic(vetID int, clinicID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", vetID)),
			tgbotapi.NewInlineKeyboardButtonData("💬 Оставить отзыв", fmt.Sprintf("add_review_%d", vetID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 К списку врачей", fmt.Sprintf("search_clinic_%d", clinicID)),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// // sendVetWithClinicDetailsAndReviews отправляет врача с детальной информацией и кнопками отзывов для клиник
// func (h *VetHandlers) sendVetWithClinicDetailsAndReviews(chatID int64, vet *models.Veterinarian, index int, clinicID int) error {
// 	var sb strings.Builder

// 	sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", index, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
// 	sb.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", html.EscapeString(vet.Phone)))

// 	if vet.Email.Valid && vet.Email.String != "" {
// 		sb.WriteString(fmt.Sprintf("📧 *Email:* %s\n", html.EscapeString(vet.Email.String)))
// 	}

// 	if vet.ExperienceYears.Valid {
// 		sb.WriteString(fmt.Sprintf("💼 *Опыт:* %d лет\n", vet.ExperienceYears.Int64))
// 	}

// 	// Специализации врача
// 	specs, err := h.db.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(vet))
// 	if err == nil && len(specs) > 0 {
// 		sb.WriteString("🎯 *Специализации:* ")
// 		specNames := make([]string, len(specs))
// 		for j, spec := range specs {
// 			specNames[j] = html.EscapeString(spec.Name)
// 		}
// 		sb.WriteString(strings.Join(specNames, ", "))
// 		sb.WriteString("\n")
// 	}

// 	// Расписание врача в этой клинике
// 	schedules, err := h.db.GetSchedulesByVetID(models.GetVetIDAsIntOrZero(vet))
// 	if err == nil {
// 		hasSchedule := false
// 		for _, schedule := range schedules {
// 			if schedule.ClinicID == clinicID {
// 				if !hasSchedule {
// 					sb.WriteString("🕐 *Расписание в этой клинике:*\n")
// 					hasSchedule = true
// 				}
// 				scheduleDayName := getDayName(schedule.DayOfWeek)
// 				startTime := schedule.StartTime
// 				endTime := schedule.EndTime
// 				if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
// 					sb.WriteString(fmt.Sprintf("• %s: %s-%s\n", scheduleDayName, startTime, endTime))
// 				}
// 			}
// 		}
// 	}

// 	// Создаем клавиатуру с кнопками отзывов
// 	keyboard := tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
// 			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", models.GetVetIDAsIntOrZero(vet))),
// 		),
// 	)

// 	msg := tgbotapi.NewMessage(chatID, sb.String())
// 	msg.ParseMode = "Markdown"
// 	msg.ReplyMarkup = keyboard

// 	_, err = h.bot.Send(msg)
// 	return err
// }

// HandleCallback обрабатывает все inline callback запросы
func (h *VetHandlers) HandleCallback(update tgbotapi.Update) {
	InfoLog.Printf("HandleCallback called")

	callback := update.CallbackQuery
	data := callback.Data

	InfoLog.Printf("Callback data: %s", data)

	// Обрабатываем разные типы callback данных
	switch {
	case data == "main_menu":
		// Сохраняем текущее состояние перед переходом в главное меню
		currentState := h.stateManager.GetUserState(callback.From.ID)
		if currentState != "" {
			h.stateManager.PushState(callback.From.ID, currentState)
		}
		h.showMainMenu(callback)
	case data == "main_specializations":
		h.stateManager.PushState(callback.From.ID, "main_menu")
		h.HandleSpecializations(update)
	case data == "main_time":
		h.stateManager.PushState(callback.From.ID, "main_menu")
		h.HandleSearch(update)
	case data == "main_clinics":
		h.stateManager.PushState(callback.From.ID, "main_menu")
		h.HandleClinics(update)
	case data == "main_city":
		h.stateManager.PushState(callback.From.ID, "main_menu")
		h.HandleSearchByCity(update)
	case data == "main_help":
		h.stateManager.PushState(callback.From.ID, "main_menu")
		h.HandleHelp(update)
	case strings.HasPrefix(data, "search_spec_"):
		h.stateManager.PushState(callback.From.ID, "main_specializations")
		h.handleSearchSpecCallback(callback)
	case strings.HasPrefix(data, "search_day_"):
		h.stateManager.PushState(callback.From.ID, "main_time")
		h.handleDaySelection(callback)
	case strings.HasPrefix(data, "search_clinic_"):
		h.stateManager.PushState(callback.From.ID, "main_clinics")
		h.handleSearchClinicCallback(callback)
	case strings.HasPrefix(data, "search_city_"):
		h.stateManager.PushState(callback.From.ID, "main_city")
		h.handleSearchCityCallback(callback)
	case strings.HasPrefix(data, "vet_details_"):
		h.handleVetDetailsCallback(callback)
	case strings.HasPrefix(data, "show_reviews_"):
		h.handleShowReviewsCallback(callback)
	case strings.HasPrefix(data, "add_review_"):
		h.handleAddReviewCallback(callback)
	case strings.HasPrefix(data, "review_rate_"):
		h.handleReviewRatingCallback(update)
	case data == "review_cancel":
		h.handleReviewCancelCallback(update)
	case strings.HasPrefix(data, "vet_details_clinic_"):
		h.handleVetDetailsFromClinicCallback(callback)
	case strings.HasPrefix(data, "add_clinic_review_"):
		h.handleAddClinicReviewCallback(callback)
	default:
		// Неизвестный callback
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Неизвестная команда")
		h.bot.Request(callbackConfig)
	}
}

// handleAddClinicReviewCallback обрабатывает добавление отзыва о клинике
func (h *VetHandlers) handleAddClinicReviewCallback(callback *tgbotapi.CallbackQuery) {
	clinicIDStr := strings.TrimPrefix(callback.Data, "add_clinic_review_")
	clinicID, err := strconv.Atoi(clinicIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing clinic ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "❌ Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	// Получаем информацию о клинике для отображения в сообщении
	clinic, err := h.db.GetClinicByID(clinicID)
	var clinicName string
	if err != nil {
		ErrorLog.Printf("Error getting clinic: %v", err)
		clinicName = "этой клинике"
	} else {
		clinicName = clinic.Name
	}

	// Пока просто сообщение что функционал в разработке
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("⭐ *Отзывы о клинике \"%s\"*\n\nФункционал отзывов о клиниках находится в разработке и будет доступен в ближайшее время!", html.EscapeString(clinicName)))
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	callbackConfig := tgbotapi.NewCallback(callback.ID, "📝 Функционал в разработке")
	h.bot.Request(callbackConfig)
}

// handleVetDetailsFromClinicCallback обрабатывает callback для детального просмотра врача из клиники
func (h *VetHandlers) handleVetDetailsFromClinicCallback(callback *tgbotapi.CallbackQuery) {
	data := strings.TrimPrefix(callback.Data, "vet_details_clinic_")
	parts := strings.Split(data, "_")
	if len(parts) != 2 {
		ErrorLog.Printf("Invalid vet_details_clinic callback data: %s", callback.Data)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	vetID, err := strconv.Atoi(parts[0])
	if err != nil {
		ErrorLog.Printf("Error parsing vet ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	clinicID, err := strconv.Atoi(parts[1])
	if err != nil {
		ErrorLog.Printf("Error parsing clinic ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Showing details for vet ID: %d in clinic ID: %d", vetID, clinicID)

	err = h.HandleVetDetailsFromClinic(callback.Message.Chat.ID, vetID, clinicID, callback.Message.MessageID)
	if err != nil {
		ErrorLog.Printf("Error showing vet details: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка при загрузке данных")
		h.bot.Request(callbackConfig)
		return
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	h.bot.Request(callbackConfig)
}

// handleVetDetailsCallback обрабатывает callback для детального просмотра врача
func (h *VetHandlers) handleVetDetailsCallback(callback *tgbotapi.CallbackQuery) {
	vetIDStr := strings.TrimPrefix(callback.Data, "vet_details_")
	vetID, err := strconv.Atoi(vetIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing vet ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Showing details for vet ID: %d", vetID)

	err = h.HandleVetDetails(callback.Message.Chat.ID, vetID, callback.Message.MessageID)
	if err != nil {
		ErrorLog.Printf("Error showing vet details: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка при загрузке данных")
		h.bot.Request(callbackConfig)
		return
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	h.bot.Request(callbackConfig)
}

// showMainMenu показывает главное меню
func (h *VetHandlers) showMainMenu(callback *tgbotapi.CallbackQuery) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Поиск по специализациям", "main_specializations"),
			tgbotapi.NewInlineKeyboardButtonData("🕐 Поиск по времени", "main_time"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏥 Поиск по клиникам", "main_clinics"),
			tgbotapi.NewInlineKeyboardButtonData("🏙️ Поиск по городу", "main_city"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Помощь", "main_help"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
		`🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Выберите способ поиска:`)
	editMsg.ReplyMarkup = &keyboard

	_, err := h.bot.Send(editMsg)
	if err != nil {
		ErrorLog.Printf("Error editing message to main menu: %v", err)
		// Если редактирование не удалось, отправляем новое сообщение
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
			`🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Выберите способ поиска:`)
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	h.bot.Request(callbackConfig)
}

// handleSearchSpecCallback обрабатывает callback поиска по специализации
func (h *VetHandlers) handleSearchSpecCallback(callback *tgbotapi.CallbackQuery) {
	specIDStr := strings.TrimPrefix(callback.Data, "search_spec_")
	specID, err := strconv.Atoi(specIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing specialization ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Searching for specialization ID: %d", specID)

	// Создаем update для передачи в HandleSearchBySpecialization
	update := tgbotapi.Update{
		CallbackQuery: callback,
	}
	h.HandleSearchBySpecialization(update, specID)
}

// handleSearchClinicCallback обрабатывает callback поиска по клинике
func (h *VetHandlers) handleSearchClinicCallback(callback *tgbotapi.CallbackQuery) {
	clinicIDStr := strings.TrimPrefix(callback.Data, "search_clinic_")
	clinicID, err := strconv.Atoi(clinicIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing clinic ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Searching for clinic ID: %d", clinicID)

	update := tgbotapi.Update{
		CallbackQuery: callback,
	}
	h.HandleSearchByClinic(update, clinicID)
}

// handleSearchCityCallback обрабатывает callback поиска по городу
func (h *VetHandlers) handleSearchCityCallback(callback *tgbotapi.CallbackQuery) {
	cityIDStr := strings.TrimPrefix(callback.Data, "search_city_")
	cityID, err := strconv.Atoi(cityIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing city ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Searching for city ID: %d", cityID)

	criteria := &models.SearchCriteria{
		CityID: cityID,
	}

	vets, err := h.db.FindVetsByCity(criteria)
	if err != nil {
		ErrorLog.Printf("Error finding vets by city: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка при поиске врачей")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Found %d vets for city %d", len(vets), cityID)

	// Получаем информацию о городе
	city, err := h.db.GetCityByID(cityID)
	if err != nil {
		ErrorLog.Printf("Error getting city: %v", err)
		city = &models.City{Name: "Неизвестный город"}
	}

	// Клавиатура с кнопками навигации
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 К городам", "main_city"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if len(vets) == 0 {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
			fmt.Sprintf("🏙️ *Врачи в городе \"%s\" не найдены*\n\nПопробуйте выбрать другой город.", city.Name))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
		h.bot.Request(callbackConfig)
		return
	}

	// Отправляем заголовок
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("🏙️ *Врачи в городе \"%s\":*\n\nНайдено врачей: %d\n\nВыберите врача для просмотра отзывов:", city.Name, len(vets)))
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Отправляем каждого врача с детальной информацией и кнопками отзывов
	for i, vet := range vets {
		err := h.sendVetWithDetailsAndReviewsButtons(callback.Message.Chat.ID, vet, i+1)
		if err != nil {
			ErrorLog.Printf("Error sending vet info: %v", err)
		}
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
	h.bot.Request(callbackConfig)
}

// sendVetWithDetailsAndReviewsButtons отправляет врача с детальной информацией и кнопками отзывов для городов
func (h *VetHandlers) sendVetWithDetailsAndReviewsButtons(chatID int64, vet *models.Veterinarian, index int) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", index, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
	sb.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", html.EscapeString(vet.Phone)))

	if vet.Email.Valid && vet.Email.String != "" {
		sb.WriteString(fmt.Sprintf("📧 *Email:* %s\n", html.EscapeString(vet.Email.String)))
	}

	if vet.ExperienceYears.Valid {
		sb.WriteString(fmt.Sprintf("💼 *Опыт:* %d лет\n", vet.ExperienceYears.Int64))
	}

	// Специализации врача
	specs, err := h.db.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil && len(specs) > 0 {
		sb.WriteString("🎯 *Специализации:* ")
		specNames := make([]string, len(specs))
		for j, spec := range specs {
			specNames[j] = html.EscapeString(spec.Name)
		}
		sb.WriteString(strings.Join(specNames, ", "))
		sb.WriteString("\n")
	}

	// Информация о городе
	if vet.City != nil {
		sb.WriteString(fmt.Sprintf("🏙️ *Город:* %s\n", vet.City.Name))
	}

	// Создаем клавиатуру с кнопками отзывов
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", models.GetVetIDAsIntOrZero(vet))),
		),
	)

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err = h.bot.Send(msg)
	return err
}

// handleDaySelection обрабатывает выбор дня для поиска
func (h *VetHandlers) handleDaySelection(callback *tgbotapi.CallbackQuery) {
	InfoLog.Printf("handleDaySelection called")

	data := callback.Data
	dayStr := strings.TrimPrefix(data, "search_day_")
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		ErrorLog.Printf("Error parsing day: %v", err)
		return
	}

	InfoLog.Printf("Searching for day: %d", day)

	criteria := &models.SearchCriteria{
		DayOfWeek: day,
	}

	vets, err := h.db.FindAvailableVets(criteria)
	if err != nil {
		ErrorLog.Printf("Error finding vets: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка при поиске врачей")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Found %d vets for day %d", len(vets), day)

	// Клавиатура с кнопками навигации
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 К дням недели", "main_time"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if len(vets) == 0 {
		dayName := getDayName(day)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID,
			fmt.Sprintf("🕐 *Врачи, работающие в %s, не найдены*\n\nПопробуйте выбрать другой день.", dayName))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
		h.bot.Request(callbackConfig)
		return
	}

	// Отправляем заголовок
	dayName := getDayName(day)
	headerMsg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("🕐 *Врачи, работающие в %s:*\n\nНайдено врачей: %d\n\nВыберите врача для просмотра отзывов:", dayName, len(vets)))
	headerMsg.ParseMode = "Markdown"
	h.bot.Send(headerMsg)

	// Отправляем каждого врача с детальной информацией и кнопками отзывов
	for i, vet := range vets {
		err := h.sendVetWithDayDetailsAndReviews(callback.Message.Chat.ID, vet, i+1, day)
		if err != nil {
			ErrorLog.Printf("Error sending vet info: %v", err)
		}
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
	h.bot.Request(callbackConfig)
}

// sendVetWithDayDetailsAndReviews отправляет врача с детальной информацией и кнопками отзывов для дней недели
func (h *VetHandlers) sendVetWithDayDetailsAndReviews(chatID int64, vet *models.Veterinarian, index int, day int) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", index, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
	sb.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", html.EscapeString(vet.Phone)))

	if vet.Email.Valid && vet.Email.String != "" {
		sb.WriteString(fmt.Sprintf("📧 *Email:* %s\n", html.EscapeString(vet.Email.String)))
	}

	if vet.ExperienceYears.Valid {
		sb.WriteString(fmt.Sprintf("💼 *Опыт:* %d лет\n", vet.ExperienceYears.Int64))
	}

	// Специализации врача
	specs, err := h.db.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil && len(specs) > 0 {
		sb.WriteString("🎯 *Специализации:* ")
		specNames := make([]string, len(specs))
		for j, spec := range specs {
			specNames[j] = html.EscapeString(spec.Name)
		}
		sb.WriteString(strings.Join(specNames, ", "))
		sb.WriteString("\n")
	}

	// Информация о городе
	if vet.City != nil {
		sb.WriteString(fmt.Sprintf("🏙️ *Город:* %s\n", vet.City.Name))
	}

	// Расписание для выбранного дня
	schedules, err := h.db.GetSchedulesByVetID(models.GetVetIDAsIntOrZero(vet))
	if err == nil {
		hasSchedule := false
		for _, schedule := range schedules {
			if schedule.DayOfWeek == day || day == 0 {
				scheduleDayName := getDayName(schedule.DayOfWeek)
				startTime := schedule.StartTime
				endTime := schedule.EndTime
				// Проверяем, что время корректное
				if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
					if !hasSchedule {
						sb.WriteString("🕐 *Расписание:*\n")
						hasSchedule = true
					}
					sb.WriteString(fmt.Sprintf("• %s: %s-%s", scheduleDayName, startTime, endTime))
					if schedule.Clinic != nil && schedule.Clinic.Name != "" {
						sb.WriteString(fmt.Sprintf(" (%s)", html.EscapeString(schedule.Clinic.Name)))
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	// Создаем клавиатуру с кнопками отзывов
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", models.GetVetIDAsIntOrZero(vet))),
		),
	)

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err = h.bot.Send(msg)
	return err
}

// HandleTest для тестирования
func (h *VetHandlers) HandleTest(update tgbotapi.Update) {
	InfoLog.Printf("HandleTest called")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Тестовое сообщение: бот работает!")
	_, err := h.bot.Send(msg)
	if err != nil {
		ErrorLog.Printf("Error sending test message: %v", err)
	} else {
		InfoLog.Printf("Test message sent successfully")
	}
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

// min возвращает минимальное из двух чисел
// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// handleShowReviewsCallback обрабатывает показ отзывов
func (h *VetHandlers) handleShowReviewsCallback(callback *tgbotapi.CallbackQuery) {
	vetIDStr := strings.TrimPrefix(callback.Data, "show_reviews_")
	vetID, err := strconv.Atoi(vetIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing vet ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	// Создаем update для передачи в ReviewHandlers
	update := tgbotapi.Update{
		CallbackQuery: callback,
	}

	// Здесь нужно будет передать управление в ReviewHandlers
	// Пока используем существующий функционал
	h.reviewHandlers.HandleShowReviews(update, vetID)

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	h.bot.Request(callbackConfig)
}

// handleAddReviewCallback обрабатывает добавление отзыва
func (h *VetHandlers) handleAddReviewCallback(callback *tgbotapi.CallbackQuery) {
	vetIDStr := strings.TrimPrefix(callback.Data, "add_review_")
	vetID, err := strconv.Atoi(vetIDStr)
	if err != nil {
		ErrorLog.Printf("Error parsing vet ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "❌ Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	// Создаем update для передачи в ReviewHandlers
	update := tgbotapi.Update{
		CallbackQuery: callback,
	}

	// Передаем управление в ReviewHandlers
	h.reviewHandlers.HandleAddReview(update, vetID)

	// Подтверждаем callback - убираем "часики" на кнопке
	callbackConfig := tgbotapi.NewCallback(callback.ID, "📝 Начинаем добавление отзыва...")
	h.bot.Request(callbackConfig)
}

// formatVeterinarianInfoCompact форматирует информацию о враче в компактном виде
func (h *VetHandlers) formatVeterinarianInfoCompact(vet *models.Veterinarian, index int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", index, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
	sb.WriteString(fmt.Sprintf("📞 `%s`", html.EscapeString(vet.Phone)))

	if vet.Email.Valid && vet.Email.String != "" {
		sb.WriteString(fmt.Sprintf(" 📧 %s", html.EscapeString(vet.Email.String)))
	}

	if vet.ExperienceYears.Valid {
		sb.WriteString(fmt.Sprintf(" 💼 %d лет", vet.ExperienceYears.Int64))
	}

	// Информация о городе
	if vet.City != nil {
		sb.WriteString(fmt.Sprintf(" 🏙️ %s", vet.City.Name))
	}

	// Специализации врача (только названия)
	if len(vet.Specializations) > 0 {
		sb.WriteString(" 🎯 ")
		specNames := make([]string, len(vet.Specializations))
		for j, spec := range vet.Specializations {
			specNames[j] = html.EscapeString(spec.Name)
		}
		sb.WriteString(strings.Join(specNames, ", "))
	}

	sb.WriteString("\n\n")
	return sb.String()
}

// sendVetWithReviewsButtons отправляет врача с кнопками отзывов в компактном формате
// func (h *VetHandlers) sendVetWithReviewsButtons(chatID int64, vet *models.Veterinarian, index int) error {
// 	message := h.formatVeterinarianInfoCompact(vet, index)

// 	// Создаем клавиатуру с кнопками "Подробнее" и "Отзывы"
// 	keyboard := tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData("📋 Подробнее", fmt.Sprintf("vet_details_%d", models.GetVetIDAsIntOrZero(vet))),
// 			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
// 		),
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", models.GetVetIDAsIntOrZero(vet))),
// 		),
// 	)

// 	msg := tgbotapi.NewMessage(chatID, message)
// 	msg.ParseMode = "Markdown"
// 	msg.ReplyMarkup = keyboard

// 	_, err := h.bot.Send(msg)
// 	return err
// }

// sendVetWithDetailsButton отправляет врача с кнопкой "Подробнее" в компактном формате
func (h *VetHandlers) sendVetWithDetailsButton(chatID int64, vet *models.Veterinarian, index int) error {
	message := h.formatVeterinarianInfoCompact(vet, index)

	// Создаем клавиатуру с кнопкой "Подробнее" и "Отзывы"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Детальная информация", fmt.Sprintf("vet_details_%d", models.GetVetIDAsIntOrZero(vet))),
			tgbotapi.NewInlineKeyboardButtonData("⭐ Отзывы", fmt.Sprintf("show_reviews_%d", models.GetVetIDAsIntOrZero(vet))),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", models.GetVetIDAsIntOrZero(vet))),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

// splitVetsIntoMessagesBySpecialization разбивает список врачей по специализации на несколько сообщений
// func (h *VetHandlers) splitVetsIntoMessagesBySpecialization(vets []*models.Veterinarian, spec *models.Specialization) []string {
// 	var messages []string
// 	var currentMessage strings.Builder

// 	// Заголовок для первого сообщения
// 	if spec != nil {
// 		currentMessage.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\":*\n\n", html.EscapeString(spec.Name)))
// 	} else {
// 		currentMessage.WriteString("👨‍⚕️ *Найденные врачи:*\n\n")
// 	}

// 	maxDisplay := 10 // Ограничиваем первое сообщение 10 врачами
// 	displayCount := min(len(vets), maxDisplay)

// 	for i := 0; i < displayCount; i++ {
// 		vet := vets[i]
// 		vetText := h.formatVeterinarianInfoCompact(vet, i+1)

// 		// Проверяем не превысит ли добавление нового врача лимит
// 		if currentMessage.Len()+len(vetText) > 3500 { // Оставляем запас
// 			messages = append(messages, currentMessage.String())
// 			currentMessage.Reset()
// 			if spec != nil {
// 				currentMessage.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\" (продолжение):*\n\n", html.EscapeString(spec.Name)))
// 			} else {
// 				currentMessage.WriteString("👨‍⚕️ *Найденные врачи (продолжение):*\n\n")
// 			}
// 		}

// 		currentMessage.WriteString(vetText)
// 	}

// 	// Добавляем информацию если есть еще врачи
// 	if len(vets) > maxDisplay {
// 		currentMessage.WriteString(fmt.Sprintf("\n📄 *Показано %d из %d врачей*. Для детального просмотра используйте поиск по конкретным критериям.",
// 			maxDisplay, len(vets)))
// 	}

// 	// Добавляем первое сообщение
// 	if currentMessage.Len() > 0 {
// 		messages = append(messages, currentMessage.String())
// 	}

// 	// Если врачей больше 10, создаем дополнительные сообщения
// 	if len(vets) > maxDisplay {
// 		for i := maxDisplay; i < len(vets); i += 10 {
// 			var continuationBuilder strings.Builder

// 			if spec != nil {
// 				continuationBuilder.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\" (продолжение %d):*\n\n",
// 					html.EscapeString(spec.Name), (i/10)+1))
// 			} else {
// 				continuationBuilder.WriteString(fmt.Sprintf("👨‍⚕️ *Найденные врачи (продолжение %d):*\n\n", (i/10)+1))
// 			}

// 			endIndex := min(i+10, len(vets))
// 			for j := i; j < endIndex; j++ {
// 				vet := vets[j]
// 				vetText := h.formatVeterinarianInfoCompact(vet, j+1)
// 				continuationBuilder.WriteString(vetText)
// 			}

// 			messages = append(messages, continuationBuilder.String())
// 		}
// 	}

// 	return messages
// }

// handleReviewRatingCallback обрабатывает выбор рейтинга для отзыва
func (h *VetHandlers) handleReviewRatingCallback(update tgbotapi.Update) {
	callback := update.CallbackQuery
	data := callback.Data

	// Извлекаем рейтинг из callback data (review_rate_1, review_rate_2 и т.д.)
	ratingStr := strings.TrimPrefix(data, "review_rate_")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		ErrorLog.Printf("Invalid rating in callback: %s", data)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "❌ Неверный рейтинг")
		h.bot.Request(callbackConfig)
		return
	}

	InfoLog.Printf("Processing review rating: %d", rating)

	// Передаем обработку в ReviewHandlers
	h.reviewHandlers.HandleReviewRating(update, rating)

	// Подтверждаем callback
	callbackConfig := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("✅ Выбрано %d звезд", rating))
	h.bot.Request(callbackConfig)
}

// handleReviewCancelCallback обрабатывает отмену добавления отзыва
func (h *VetHandlers) handleReviewCancelCallback(update tgbotapi.Update) {
	InfoLog.Printf("Processing review cancellation")

	// Передаем обработку в ReviewHandlers
	h.reviewHandlers.HandleReviewCancel(update)

	// Подтверждаем callback
	callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "❌ Отмена")
	h.bot.Request(callbackConfig)
}
