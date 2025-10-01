package handlers

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// VetHandlers содержит обработчики для ветеринарного бота
type VetHandlers struct {
	bot BotAPI
	db  Database
}

// NewVetHandlers создает новый экземпляр VetHandlers
func NewVetHandlers(bot BotAPI, db Database) *VetHandlers {
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
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Помощь", "main_help"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		`🐾 Добро пожаловать в VetBot! 🐾

Я ваш помощник в поиске ветеринарных врачей. Выберите способ поиска:`)
	msg.ReplyMarkup = keyboard

	log.Printf("Sending start message with inline keyboard")
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	} else {
		log.Printf("Start message sent successfully")
	}
}

// HandleSpecializations показывает список специализаций с улучшенным интерфейсом
func (h *VetHandlers) HandleSpecializations(update tgbotapi.Update) {
	log.Printf("HandleSpecializations called")

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
		log.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		log.Printf("Error getting specializations: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении списка специализаций")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d specializations", len(specializations))

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

	log.Printf("Sending specializations menu to chat %d", chatID)
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending specializations menu: %v", err)
	}
}

// HandleSearch показывает меню поиска по времени
func (h *VetHandlers) HandleSearch(update tgbotapi.Update) {
	log.Printf("HandleSearch called")

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		log.Printf("Error: both CallbackQuery and Message are nil")
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

	log.Printf("Sending search by time menu to chat %d", chatID)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending search menu: %v", err)
	}
}

// HandleClinics показывает меню клиник
func (h *VetHandlers) HandleClinics(update tgbotapi.Update) {
	log.Printf("HandleClinics called")

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		log.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	clinics, err := h.db.GetAllClinics()
	if err != nil {
		log.Printf("Error getting clinics: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении списка клиник")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d clinics", len(clinics))

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

	log.Printf("Sending clinics menu to chat %d", chatID)
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending clinics menu: %v", err)
	}
}

// HandleSearchByCity показывает меню поиска по городам
func (h *VetHandlers) HandleSearchByCity(update tgbotapi.Update) {
	log.Printf("HandleSearchByCity called")

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		log.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	// Получаем список городов
	cities, err := h.db.GetAllCities()
	if err != nil {
		log.Printf("Error getting cities: %v", err)
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

	log.Printf("Sending cities menu to chat %d", chatID)
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending cities menu: %v", err)
	}
}

// HandleHelp показывает справку с кнопкой "Назад"
func (h *VetHandlers) HandleHelp(update tgbotapi.Update) {
	log.Printf("HandleHelp called")

	var chatID int64

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		log.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "main_menu"),
		),
	)

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
	msg.ReplyMarkup = keyboard

	log.Printf("Sending help message to chat %d", chatID)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

// HandleSearchBySpecialization ищет врачей по специализации с кнопкой "Назад"
func (h *VetHandlers) HandleSearchBySpecialization(update tgbotapi.Update, specializationID int) {
	log.Printf("HandleSearchBySpecialization called with ID: %d", specializationID)

	var chatID int64
	var messageID int

	// Определяем chatID в зависимости от типа update
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		messageID = update.CallbackQuery.Message.MessageID
		// Отвечаем на callback query чтобы убрать "часики" у кнопки
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		log.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	vets, err := h.db.GetVeterinariansBySpecialization(specializationID)
	if err != nil {
		log.Printf("Error getting veterinarians: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при поиске врачей")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d veterinarians for specialization ID: %d", len(vets), specializationID)

	spec, err := h.db.GetSpecializationByID(specializationID)
	if err != nil {
		log.Printf("Error getting specialization: %v", err)
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

	// Разбиваем результаты на сообщения
	messages := h.splitVetsIntoMessagesBySpecialization(vets, spec)

	// Отправляем первое сообщение с клавиатурой
	if len(messages) > 0 {
		firstMessage := messages[0]

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, firstMessage)
		editMsg.ParseMode = "Markdown"
		editMsg.ReplyMarkup = &keyboard
		_, err = h.bot.Send(editMsg)

		if err != nil {
			log.Printf("Error editing message: %v", err)
			// Если редактирование не удалось, отправляем новое сообщение
			msg := tgbotapi.NewMessage(chatID, firstMessage)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = keyboard
			h.bot.Send(msg)
		}

		// Отправляем остальные сообщения если есть
		for i := 1; i < len(messages); i++ {
			msg := tgbotapi.NewMessage(chatID, messages[i])
			msg.ParseMode = "Markdown"
			h.bot.Send(msg)
		}
	}
}

// splitVetsIntoMessagesBySpecialization разбивает список врачей по специализации на несколько сообщений
func (h *VetHandlers) splitVetsIntoMessagesBySpecialization(vets []*models.Veterinarian, spec *models.Specialization) []string {
	var messages []string
	var currentMessage strings.Builder

	// Заголовок для первого сообщения
	if spec != nil {
		currentMessage.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\":*\n\n", html.EscapeString(spec.Name)))
	} else {
		currentMessage.WriteString("👨‍⚕️ *Найденные врачи:*\n\n")
	}

	maxDisplay := 10 // Ограничиваем первое сообщение 10 врачами
	displayCount := min(len(vets), maxDisplay)

	for i := 0; i < displayCount; i++ {
		vet := vets[i]
		vetText := h.formatVeterinarianInfoCompact(vet, i+1)

		// Проверяем не превысит ли добавление нового врача лимит
		if currentMessage.Len()+len(vetText) > 3500 { // Оставляем запас
			messages = append(messages, currentMessage.String())
			currentMessage.Reset()
			if spec != nil {
				currentMessage.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\" (продолжение):*\n\n", html.EscapeString(spec.Name)))
			} else {
				currentMessage.WriteString("👨‍⚕️ *Найденные врачи (продолжение):*\n\n")
			}
		}

		currentMessage.WriteString(vetText)
	}

	// Добавляем информацию если есть еще врачи
	if len(vets) > maxDisplay {
		currentMessage.WriteString(fmt.Sprintf("\n📄 *Показано %d из %d врачей*. Для детального просмотра используйте поиск по конкретным критериям.",
			maxDisplay, len(vets)))
	}

	// Добавляем первое сообщение
	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}

	// Если врачей больше 10, создаем дополнительные сообщения
	if len(vets) > maxDisplay {
		for i := maxDisplay; i < len(vets); i += 10 {
			var continuationBuilder strings.Builder

			if spec != nil {
				continuationBuilder.WriteString(fmt.Sprintf("👨‍⚕️ *Врачи по специализации \"%s\" (продолжение %d):*\n\n",
					html.EscapeString(spec.Name), (i/10)+1))
			} else {
				continuationBuilder.WriteString(fmt.Sprintf("👨‍⚕️ *Найденные врачи (продолжение %d):*\n\n", (i/10)+1))
			}

			endIndex := min(i+10, len(vets))
			for j := i; j < endIndex; j++ {
				vet := vets[j]
				vetText := h.formatVeterinarianInfoCompact(vet, j+1)
				continuationBuilder.WriteString(vetText)
			}

			messages = append(messages, continuationBuilder.String())
		}
	}

	return messages
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

	// Специализации врача (только названия)
	specs, err := h.db.GetSpecializationsByVetID(vet.ID)
	if err == nil && len(specs) > 0 {
		sb.WriteString(" 🎯 ")
		specNames := make([]string, len(specs))
		for j, spec := range specs {
			specNames[j] = html.EscapeString(spec.Name)
		}
		sb.WriteString(strings.Join(specNames, ", "))
	}

	sb.WriteString("\n\n")
	return sb.String()
}

// HandleSearchByClinic ищет врачей по клинике
func (h *VetHandlers) HandleSearchByClinic(update tgbotapi.Update, clinicID int) {
	log.Printf("HandleSearchByClinic called with ID: %d", clinicID)

	var chatID int64
	var messageID int

	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		messageID = update.CallbackQuery.Message.MessageID
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Send(callback)
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		log.Printf("Error: both CallbackQuery and Message are nil")
		return
	}

	// Получаем врачей клиники через расписание
	criteria := &models.SearchCriteria{
		ClinicID: clinicID,
	}
	vets, err := h.db.FindAvailableVets(criteria)
	if err != nil {
		log.Printf("Error finding vets by clinic: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при поиске врачей")
		h.bot.Send(msg)
		return
	}

	log.Printf("Found %d veterinarians for clinic ID: %d", len(vets), clinicID)

	// Получаем информацию о клинике
	clinics, err := h.db.GetAllClinics()
	if err != nil {
		log.Printf("Error getting clinics: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при получении информации о клинике")
		h.bot.Send(msg)
		return
	}

	var clinicName string
	for _, c := range clinics {
		if c.ID == clinicID {
			clinicName = c.Name
			break
		}
	}

	// Если клиника не найдена, используем заглушку
	if clinicName == "" {
		clinicName = "Неизвестная клиника"
	}

	// Клавиатура с кнопками навигации
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 К клиникам", "main_clinics"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if len(vets) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("🏥 *Врачи в клинике \"%s\" не найдены*\n\nПопробуйте выбрать другую клинику.", clinicName))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏥 *Врачи в клинике \"%s\":*\n\n", html.EscapeString(clinicName)))

	for i, vet := range vets {
		sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", i+1, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
		sb.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", html.EscapeString(vet.Phone)))

		if vet.Email.Valid && vet.Email.String != "" {
			sb.WriteString(fmt.Sprintf("📧 *Email:* %s\n", html.EscapeString(vet.Email.String)))
		}

		if vet.ExperienceYears.Valid {
			sb.WriteString(fmt.Sprintf("💼 *Опыт:* %d лет\n", vet.ExperienceYears.Int64))
		}

		// Специализации врача
		specs, err := h.db.GetSpecializationsByVetID(vet.ID)
		if err == nil && len(specs) > 0 {
			sb.WriteString("🎯 *Специализации:* ")
			specNames := make([]string, len(specs))
			for j, spec := range specs {
				specNames[j] = html.EscapeString(spec.Name)
			}
			sb.WriteString(strings.Join(specNames, ", "))
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	if update.CallbackQuery != nil && messageID != 0 {
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
		editMsg.ParseMode = "Markdown"
		editMsg.ReplyMarkup = &keyboard
		h.bot.Send(editMsg)
	} else {
		msg := tgbotapi.NewMessage(chatID, sb.String())
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
	}
}

// HandleCallback обрабатывает все inline callback запросы
func (h *VetHandlers) HandleCallback(update tgbotapi.Update) {
	log.Printf("HandleCallback called")

	callback := update.CallbackQuery
	data := callback.Data

	log.Printf("Callback data: %s", data)

	// Обрабатываем разные типы callback данных
	switch {
	case data == "main_menu":
		h.showMainMenu(callback)
	case data == "main_specializations":
		h.HandleSpecializations(update)
	case data == "main_time":
		h.HandleSearch(update)
	case data == "main_clinics":
		h.HandleClinics(update)
	case data == "main_city":
		h.HandleSearchByCity(update)
	case data == "main_help":
		h.HandleHelp(update)
	case strings.HasPrefix(data, "search_spec_"):
		h.handleSearchSpecCallback(callback)
	case strings.HasPrefix(data, "search_day_"):
		h.handleDaySelection(callback)
	case strings.HasPrefix(data, "search_clinic_"):
		h.handleSearchClinicCallback(callback)
	case strings.HasPrefix(data, "search_city_"):
		h.handleSearchCityCallback(callback)
	default:
		// Неизвестный callback
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Неизвестная команда")
		h.bot.Request(callbackConfig)
	}
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
		log.Printf("Error editing message to main menu: %v", err)
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
		log.Printf("Error parsing specialization ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	log.Printf("Searching for specialization ID: %d", specID)

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
		log.Printf("Error parsing clinic ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	log.Printf("Searching for clinic ID: %d", clinicID)

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
		log.Printf("Error parsing city ID: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка обработки запроса")
		h.bot.Request(callbackConfig)
		return
	}

	log.Printf("Searching for city ID: %d", cityID)

	criteria := &models.SearchCriteria{
		CityID: cityID,
	}

	vets, err := h.db.FindVetsByCity(criteria)
	if err != nil {
		log.Printf("Error finding vets by city: %v", err)
		callbackConfig := tgbotapi.NewCallback(callback.ID, "Ошибка при поиске врачей")
		h.bot.Request(callbackConfig)
		return
	}

	log.Printf("Found %d vets for city %d", len(vets), cityID)

	// Получаем информацию о городе
	city, err := h.db.GetCityByID(cityID)
	if err != nil {
		log.Printf("Error getting city: %v", err)
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

	// Разбиваем результаты на части если слишком много
	messages := h.splitVetsIntoMessages(vets, city.Name)

	// Отправляем первое сообщение с клавиатурой
	if len(messages) > 0 {
		firstMessage := messages[0]
		if len(messages) > 1 {
			firstMessage += fmt.Sprintf("\n\n📄 *Показано %d из %d врачей*. Для просмотра всех результатов используйте поиск по специализациям.",
				min(10, len(vets)), len(vets))
		}

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, firstMessage)
		editMsg.ParseMode = "Markdown"
		editMsg.ReplyMarkup = &keyboard
		_, err = h.bot.Send(editMsg)

		if err != nil {
			log.Printf("Error editing message: %v", err)
			// Если редактирование не удалось, отправляем новое сообщение
			msg := tgbotapi.NewMessage(callback.Message.Chat.ID, firstMessage)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = keyboard
			h.bot.Send(msg)
		}

		// Отправляем остальные сообщения если есть
		for i := 1; i < len(messages); i++ {
			msg := tgbotapi.NewMessage(callback.Message.Chat.ID, messages[i])
			msg.ParseMode = "Markdown"
			h.bot.Send(msg)
		}
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
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

	var sb strings.Builder
	dayName := getDayName(day)
	sb.WriteString(fmt.Sprintf("🕐 *Врачи, работающие в %s:*\n\n", dayName))

	for i, vet := range vets {
		sb.WriteString(fmt.Sprintf("**%d. %s %s**\n", i+1, html.EscapeString(vet.FirstName), html.EscapeString(vet.LastName)))
		sb.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", html.EscapeString(vet.Phone)))

		if vet.Email.Valid && vet.Email.String != "" {
			sb.WriteString(fmt.Sprintf("📧 *Email:* %s\n", html.EscapeString(vet.Email.String)))
		}

		if vet.ExperienceYears.Valid {
			sb.WriteString(fmt.Sprintf("💼 *Опыт:* %d лет\n", vet.ExperienceYears.Int64))
		}

		// Расписание для выбранного дня
		// Расписание для выбранного дня
		schedules, err := h.db.GetSchedulesByVetID(vet.ID)
		if err == nil {
			for _, schedule := range schedules {
				if schedule.DayOfWeek == day || day == 0 {
					scheduleDayName := getDayName(schedule.DayOfWeek)
					startTime := schedule.StartTime
					endTime := schedule.EndTime
					// Проверяем, что время корректное
					if startTime != "" && endTime != "" && startTime != "00:00" && endTime != "00:00" {
						sb.WriteString(fmt.Sprintf("🕐 *%s:* %s-%s", scheduleDayName, startTime, endTime))
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

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, sb.String())
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = &keyboard

	_, err = h.bot.Send(editMsg)
	if err != nil {
		log.Printf("Error sending day search results: %v", err)
		// Если редактирование не удалось, отправляем новое сообщение
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, sb.String())
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "Поиск завершен")
	h.bot.Request(callbackConfig)
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

// splitVetsIntoMessages разбивает список врачей на несколько сообщений
func (h *VetHandlers) splitVetsIntoMessages(vets []*models.Veterinarian, cityName string) []string {
	var messages []string
	var currentMessage strings.Builder

	// Заголовок для первого сообщения
	currentMessage.WriteString(fmt.Sprintf("🏙️ *Врачи в городе \"%s\":*\n\n", cityName))

	for i, vet := range vets {
		vetText := h.formatVeterinarianInfo(vet, i+1)

		// Проверяем не превысит ли добавление нового врача лимит
		if currentMessage.Len()+len(vetText) > 3500 { // Оставляем запас
			messages = append(messages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(fmt.Sprintf("🏙️ *Врачи в городе \"%s\" (продолжение):*\n\n", cityName))
		}

		currentMessage.WriteString(vetText)

		// Ограничиваем первое сообщение 10 врачами для лучшего UX
		if i == 9 && len(vets) > 10 {
			currentMessage.WriteString(fmt.Sprintf("\n📄 ... и еще %d врачей. Для детального просмотра используйте поиск по специализациям.", len(vets)-10))
			break
		}
	}

	// Добавляем последнее сообщение
	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}

	return messages
}

// formatVeterinarianInfo форматирует информацию о враче
func (h *VetHandlers) formatVeterinarianInfo(vet *models.Veterinarian, index int) string {
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
	specs, err := h.db.GetSpecializationsByVetID(vet.ID)
	if err == nil && len(specs) > 0 {
		sb.WriteString("🎯 *Специализации:* ")
		specNames := make([]string, len(specs))
		for j, spec := range specs {
			specNames[j] = html.EscapeString(spec.Name)
		}
		sb.WriteString(strings.Join(specNames, ", "))
		sb.WriteString("\n")
	}

	// Расписание врача (только основные дни)
	schedules, err := h.db.GetSchedulesByVetID(vet.ID)
	if err == nil && len(schedules) > 0 {
		sb.WriteString("🕐 *Расписание:* ")
		scheduleDays := make([]string, 0)

		// Группируем расписание по дням
		daySchedules := make(map[int][]string)
		for _, schedule := range schedules {
			if schedule.StartTime != "" && schedule.EndTime != "" &&
				schedule.StartTime != "00:00" && schedule.EndTime != "00:00" {

				clinicName := ""
				if schedule.Clinic != nil && schedule.Clinic.Name != "" {
					clinicName = fmt.Sprintf(" (%s)", html.EscapeString(schedule.Clinic.Name))
				}

				scheduleInfo := fmt.Sprintf("%s-%s%s",
					schedule.StartTime, schedule.EndTime, clinicName)

				daySchedules[schedule.DayOfWeek] = append(daySchedules[schedule.DayOfWeek], scheduleInfo)
			}
		}

		// Формируем строку расписания
		for day := 1; day <= 7; day++ {
			if times, exists := daySchedules[day]; exists && len(times) > 0 {
				dayName := getDayName(day)
				scheduleDays = append(scheduleDays, fmt.Sprintf("%s %s", dayName, strings.Join(times, ", ")))
			}
		}

		if len(scheduleDays) > 0 {
			sb.WriteString(strings.Join(scheduleDays, "; "))
		} else {
			sb.WriteString("не указано")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// Вспомогательная функция для минимума
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
