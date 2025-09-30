package handlers

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/internal/models"
	"github.com/drerr0r/vetbot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminHandlers содержит обработчики для административных функций
type AdminHandlers struct {
	bot        BotAPI   // Используем интерфейс
	db         Database // Используем интерфейс
	config     *utils.Config
	adminState map[int64]string
	tempData   map[string]interface{} // Добавляем недостающее поле
}

// NewAdminHandlers создает новый экземпляр AdminHandlers
func NewAdminHandlers(bot BotAPI, db Database, config *utils.Config) *AdminHandlers {
	return &AdminHandlers{
		bot:        bot,
		db:         db,
		config:     config,
		adminState: make(map[int64]string),
		tempData:   make(map[string]interface{}), // Инициализируем
	}
}

// HandleAdmin показывает админскую панель
func (h *AdminHandlers) HandleAdmin(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "main_menu"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👥 Управление врачами"),
			tgbotapi.NewKeyboardButton("🏥 Управление клиниками"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📥 Импорт данных"),
			tgbotapi.NewKeyboardButton("📊 Статистика"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚙️ Настройки"),
			tgbotapi.NewKeyboardButton("❌ Выйти из админки"),
		),
	)
	keyboard.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"🔧 *Административная панель*\n\nВыберите раздел для управления:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// HandleAdminMessage обрабатывает текстовые сообщения в админском режиме
func (h *AdminHandlers) HandleAdminMessage(update tgbotapi.Update) {
	userID := update.Message.From.ID
	text := update.Message.Text
	state := h.adminState[userID]

	log.Printf("Admin message from %d: %s (state: %s)", userID, text, state)

	// Сначала проверяем кнопку "Назад" независимо от состояния
	if text == "🔙 Назад" {
		h.handleBackButton(update)
		return
	}

	switch state {
	case "main_menu":
		h.handleMainMenu(update, text)
	case "vet_management":
		h.handleVetManagement(update, text)
	case "clinic_management":
		h.handleClinicManagement(update, text)
	case "import_menu":
		h.handleImportMenu(update, text)
	case "add_vet_name":
		h.handleAddVetName(update, text)
	case "add_vet_phone":
		h.handleAddVetPhone(update, text)
	case "add_vet_specializations":
		h.handleAddVetSpecializations(update, text)
	case "vet_list":
		h.handleVetListSelection(update, text)
	case "vet_edit_menu":
		h.handleVetEditMenu(update, text)
	case "vet_edit_field":
		h.handleVetEditField(update, text)
	case "vet_edit_specializations":
		h.handleVetEditSpecializations(update, text)
	case "vet_confirm_delete":
		h.handleVetConfirmDelete(update, text)
	case "vet_toggle_active":
		h.handleVetToggleActive(update, text)
	case "clinic_list":
		h.handleClinicListSelection(update, text)
	case "clinic_edit_menu":
		h.handleClinicEditMenu(update, text)
	case "clinic_edit_field":
		h.handleClinicEditField(update, text)
	case "clinic_confirm_delete":
		h.handleClinicConfirmDelete(update, text)
	case "clinic_toggle_active":
		h.handleClinicToggleActive(update, text)
	default:
		h.handleMainMenu(update, text)
	}
}

// handleBackButton обрабатывает кнопку "Назад"
func (h *AdminHandlers) handleBackButton(update tgbotapi.Update) {
	userID := update.Message.From.ID
	currentState := h.adminState[userID]

	// Определяем текущее состояние и возвращаемся на уровень выше
	switch currentState {
	case "vet_management", "clinic_management", "import_menu":
		// Возврат из подменю в главное меню
		h.adminState[userID] = "main_menu"
		h.HandleAdmin(update)
	case "vet_list", "vet_edit_menu", "vet_edit_field", "vet_edit_specializations", "vet_confirm_delete", "vet_toggle_active":
		// Возврат из операций с врачами в меню управления врачами
		h.adminState[userID] = "vet_management"
		h.showVetManagement(update)
	case "clinic_list", "clinic_edit_menu", "clinic_edit_field", "clinic_confirm_delete", "clinic_toggle_active":
		// Возврат из операций с клиниками в меню управления клиниками
		h.adminState[userID] = "clinic_management"
		h.showClinicManagement(update)
	case "add_vet_name", "add_vet_phone", "add_vet_specializations":
		// Возврат из процесса добавления врача в меню управления врачами
		h.adminState[userID] = "vet_management"
		h.cleanTempData(userID)
		h.showVetManagement(update)
	default:
		// По умолчанию возвращаем в главное меню
		h.adminState[userID] = "main_menu"
		h.HandleAdmin(update)
	}
}

// cleanTempData очищает временные данные пользователя
func (h *AdminHandlers) cleanTempData(userID int64) {
	userIDStr := strconv.FormatInt(userID, 10)
	delete(h.tempData, userIDStr+"_name")
	delete(h.tempData, userIDStr+"_phone")
	delete(h.tempData, userIDStr+"_vet_edit")
	delete(h.tempData, userIDStr+"_clinic_edit")
}

// handleMainMenu обрабатывает главное меню админки
func (h *AdminHandlers) handleMainMenu(update tgbotapi.Update, text string) {
	switch text {
	case "👥 Управление врачами":
		h.showVetManagement(update)
	case "🏥 Управление клиниками":
		h.showClinicManagement(update)
	case "📥 Импорт данных":
		h.showImportMenu(update)
	case "📊 Статистика":
		h.HandleStats(update)
	case "⚙️ Настройки":
		h.showSettings(update)
	case "❌ Выйти из админки":
		h.closeAdmin(update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Используйте кнопки админской панели")
		h.bot.Send(msg)
	}
}

// showImportMenu показывает меню импорта данных
func (h *AdminHandlers) showImportMenu(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "import_menu"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏙️ Импорт городов"),
			tgbotapi.NewKeyboardButton("👥 Импорт врачей"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏥 Импорт клиник"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"📥 *Импорт данных*\n\nВыберите тип данных для импорта. Поддерживаются CSV и Excel файлы.\n\n"+
			"*Формат файлов:*\n"+
			"• CSV: разделитель - точка с запятой\n"+
			"• Excel: первый лист с данными")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleImportMenu обрабатывает меню импорта
func (h *AdminHandlers) handleImportMenu(update tgbotapi.Update, text string) {
	switch text {
	case "🏙️ Импорт городов":
		h.handleImportCities(update)
	case "👥 Импорт врачей":
		h.handleImportVeterinarians(update)
	case "🏥 Импорт клиник":
		h.handleImportClinics(update)
	case "🔙 Назад":
		h.handleBackButton(update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте кнопки меню импорта")
		h.bot.Send(msg)
	}
}

// handleImportCities обрабатывает импорт городов
func (h *AdminHandlers) handleImportCities(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"📤 Для импорта городов отправьте CSV или Excel файл со следующими колонками:\n\n"+
			"1. *Название города* (обязательно)\n"+
			"2. *Регион* (обязательно)\n\n"+
			"*Пример CSV:*\n"+
			"Москва;Центральный федеральный округ\n"+
			"Санкт-Петербург;Северо-Западный федеральный округ")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// handleImportVeterinarians обрабатывает импорт врачей
func (h *AdminHandlers) handleImportVeterinarians(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"📤 Для импорта врачей отправьте CSV или Excel файл со следующими колонками:\n\n"+
			"1. *Имя* (обязательно)\n"+
			"2. *Фамилия* (обязательно)\n"+
			"3. *Телефон* (обязательно)\n"+
			"4. *Email* (опционально)\n"+
			"5. *Опыт работы* (опционально, число)\n"+
			"6. *Описание* (опционально)\n"+
			"7. *Специализации* (опционально, через запятую)\n\n"+
			"*Пример CSV:*\n"+
			"Иван;Петров;+79161234567;ivan@vet.ru;10;Опытный хирург;Хирург,Терапевт")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// handleImportClinics обрабатывает импорт клиник
func (h *AdminHandlers) handleImportClinics(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"📤 Для импорта клиник отправьте CSV или Excel файл со следующими колонками:\n\n"+
			"1. *Название* (обязательно)\n"+
			"2. *Город* (обязательно)\n"+
			"3. *Адрес* (обязательно)\n"+
			"4. *Телефон* (опционально)\n"+
			"5. *Часы работы* (опционально)\n"+
			"6. *Район* (опционально)\n"+
			"7. *Станция метро* (опционально)\n\n"+
			"*Пример CSV:*\n"+
			"ВетКлиника Центр;Москва;ул. Центральная, д.1;+74950000001;Пн-Пт 9-21;Центральный;Охотный ряд")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// handleVetManagement обрабатывает меню управления врачами
func (h *AdminHandlers) handleVetManagement(update tgbotapi.Update, text string) {
	switch text {
	case "➕ Добавить врача":
		h.startAddVet(update)
	case "🌍 Поиск по городу":
		h.handleVetSearchByCity(update)
	case "📋 Список врачей":
		h.showVetList(update)
	case "🔙 Назад":
		h.handleBackButton(update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Используйте кнопки меню управления врачами")
		h.bot.Send(msg)
	}
}

// handleVetSearchByCity обрабатывает поиск врачей по городу
func (h *AdminHandlers) handleVetSearchByCity(update tgbotapi.Update) {
	cities, err := h.db.GetAllCities()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка городов")
		h.bot.Send(msg)
		return
	}

	if len(cities) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Городы не найдены. Сначала импортируйте города.")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏙️ *Выберите город для поиска врачей:*\n\n")

	for i, city := range cities {
		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, city.Name, city.Region))
	}

	sb.WriteString("\nВведите номер города:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleClinicManagement обрабатывает меню управления клиниками
func (h *AdminHandlers) handleClinicManagement(update tgbotapi.Update, text string) {
	switch text {
	case "➕ Добавить клинику":
		h.startAddClinic(update)
	case "📋 Список клиник":
		h.showClinicList(update)
	case "🔙 Назад":
		h.handleBackButton(update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Используйте кнопки меню управления клиниками")
		h.bot.Send(msg)
	}
}

// showVetManagement показывает меню управления врачами
func (h *AdminHandlers) showVetManagement(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "vet_management"

	// Получаем статистику врачей
	activeVets, _ := h.getActiveVetCount()
	totalVets, _ := h.getTotalVetCount()

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить врача"),
			tgbotapi.NewKeyboardButton("🌍 Поиск по городу"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Список врачей"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("👥 *Управление врачами*\n\nАктивных врачей: %d/%d\n\nВыберите действие:", activeVets, totalVets))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// showClinicManagement показывает меню управления клиниками
func (h *AdminHandlers) showClinicManagement(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "clinic_management"

	// Получаем статистику клиник
	activeClinics, _ := h.getActiveClinicCount()
	totalClinics, _ := h.getTotalClinicCount()

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить клинику"),
			tgbotapi.NewKeyboardButton("📋 Список клиник"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("🏥 *Управление клиниками*\n\nАктивных клиник: %d/%d\n\nВыберите действие:", activeClinics, totalClinics))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// startAddVet начинает процесс добавления врача
func (h *AdminHandlers) startAddVet(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "add_vet_name"

	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"👨‍⚕️ *Добавление нового врача*\n\nВведите имя врача:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = removeKeyboard

	h.bot.Send(msg)
}

// handleAddVetName обрабатывает ввод имени врача
func (h *AdminHandlers) handleAddVetName(update tgbotapi.Update, name string) {
	userID := update.Message.From.ID
	h.adminState[userID] = "add_vet_phone"

	// Сохраняем имя во временное хранилище
	userIDStr := strconv.FormatInt(userID, 10)
	h.tempData[userIDStr+"_name"] = name

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"📞 Теперь введите телефон врача:")
	msg.ParseMode = "Markdown"

	h.bot.Send(msg)
}

// handleAddVetPhone обрабатывает ввод телефона врача
func (h *AdminHandlers) handleAddVetPhone(update tgbotapi.Update, phone string) {
	userID := update.Message.From.ID
	h.adminState[userID] = "add_vet_specializations"

	// Сохраняем телефон
	userIDStr := strconv.FormatInt(userID, 10)
	h.tempData[userIDStr+"_phone"] = phone

	// Получаем список специализаций для выбора
	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Ошибка при получении специализаций")
		h.bot.Send(msg)
		return
	}

	// Сортируем специализации по ID для предсказуемости
	sort.Slice(specializations, func(i, j int) bool {
		return specializations[i].ID < specializations[j].ID
	})

	var sb strings.Builder
	sb.WriteString("🎯 Выберите специализации врача (введите ID через запятую):\n\n")

	for _, spec := range specializations {
		sb.WriteString(fmt.Sprintf("ID %d: %s\n", spec.ID, spec.Name))
	}

	sb.WriteString("\nПример: 1,3,5")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	h.bot.Send(msg)
}

// handleAddVetSpecializations обрабатывает ввод специализаций
func (h *AdminHandlers) handleAddVetSpecializations(update tgbotapi.Update, specsText string) {
	userID := update.Message.From.ID

	// Получаем сохраненные данные
	userIDStr := strconv.FormatInt(userID, 10)
	name := h.getStringTempData(userIDStr + "_name")
	phone := h.getStringTempData(userIDStr + "_phone")

	if name == "" || phone == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Ошибка: данные врача не найдены. Начните заново.")
		h.bot.Send(msg)
		h.startAddVet(update)
		return
	}

	// Валидация введенных ID специализаций
	if !h.isValidSpecializationIDs(specsText) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Неверный формат ID специализаций. Введите существующие ID через запятую (например: 1,3,5)")
		h.bot.Send(msg)
		return
	}

	// Создаем врача
	vet := &models.Veterinarian{
		FirstName: name,
		LastName:  "", // Можно добавить поле для фамилии
		Phone:     phone,
		IsActive:  true,
	}

	// Добавляем врача в базу
	err := h.addVeterinarian(vet, specsText)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Ошибка при добавлении врача: %v", err))
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"✅ Врач успешно добавлен!")
		h.bot.Send(msg)
	}

	// Очищаем временные данные
	h.cleanTempData(userID)

	// Возвращаем в меню управления врачами
	h.adminState[userID] = "vet_management"
	h.showVetManagement(update)
}

// showVetList показывает список врачей с возможностью выбора
func (h *AdminHandlers) showVetList(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "vet_list"

	vets, err := h.db.GetAllVeterinarians()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка врачей")
		h.bot.Send(msg)
		return
	}

	if len(vets) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Врачи не найдены")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("👥 *Список врачей:*\n\n")

	for i, vet := range vets {
		status := "✅"
		if !vet.IsActive {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s %d. %s %s - %s\n", status, i+1, vet.FirstName, vet.LastName, vet.Phone))
	}

	sb.WriteString("\nВведите номер врача для управления:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleVetListSelection обрабатывает выбор врача из списка
func (h *AdminHandlers) handleVetListSelection(update tgbotapi.Update, text string) {
	// Парсим номер врача
	index, err := strconv.Atoi(text)
	if err != nil || index < 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите корректный номер врача")
		h.bot.Send(msg)
		return
	}

	vets, err := h.db.GetAllVeterinarians()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка врачей")
		h.bot.Send(msg)
		return
	}

	if index > len(vets) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Врач с таким номером не найден")
		h.bot.Send(msg)
		return
	}

	vet := vets[index-1]
	h.showVetEditMenu(update, vet)
}

// showVetEditMenu показывает меню редактирования врача
func (h *AdminHandlers) showVetEditMenu(update tgbotapi.Update, vet *models.Veterinarian) {
	userID := update.Message.From.ID
	h.adminState[userID] = "vet_edit_menu"

	// Сохраняем ID врача во временные данные
	userIDStr := strconv.FormatInt(userID, 10)
	h.tempData[userIDStr+"_vet_edit"] = &models.VetEditData{
		VetID: vet.ID,
	}

	// Получаем специализации врача
	specs, err := h.db.GetSpecializationsByVetID(vet.ID)
	specsText := ""
	if err == nil && len(specs) > 0 {
		var specIDs []string
		for _, spec := range specs {
			specIDs = append(specIDs, strconv.Itoa(spec.ID))
		}
		specsText = strings.Join(specIDs, ",")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👨‍⚕️ *Управление врачом:* %s %s\n\n", vet.FirstName, vet.LastName))
	sb.WriteString(fmt.Sprintf("📞 Телефон: %s\n", vet.Phone))

	if vet.Email.Valid {
		sb.WriteString(fmt.Sprintf("📧 Email: %s\n", vet.Email.String))
	}

	if vet.ExperienceYears.Valid {
		sb.WriteString(fmt.Sprintf("💼 Опыт: %d лет\n", vet.ExperienceYears.Int64))
	}

	sb.WriteString("📊 Статус: ")
	if vet.IsActive {
		sb.WriteString("✅ Активен\n")
	} else {
		sb.WriteString("❌ Неактивен\n")
	}

	sb.WriteString(fmt.Sprintf("🎯 Специализации: %s\n\n", specsText))
	sb.WriteString("Выберите действие:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✏️ Редактировать имя"),
			tgbotapi.NewKeyboardButton("📞 Редактировать телефон"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎯 Редактировать специализации"),
			tgbotapi.NewKeyboardButton("📧 Редактировать email"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💼 Редактировать опыт"),
			tgbotapi.NewKeyboardButton("⚡ Изменить статус"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗑️ Удалить врача"),
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleVetEditMenu обрабатывает выбор действия для врача
func (h *AdminHandlers) handleVetEditMenu(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	vetData, ok := h.tempData[userIDStr+"_vet_edit"].(*models.VetEditData)
	if !ok || vetData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные врача не найдены")
		h.bot.Send(msg)
		h.showVetList(update)
		return
	}

	// Получаем актуальные данные врача
	vet, err := h.db.GetVeterinarianByID(vetData.VetID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных врача")
		h.bot.Send(msg)
		h.showVetList(update)
		return
	}

	switch text {
	case "✏️ Редактировать имя":
		h.adminState[userID] = "vet_edit_field"
		vetData.Field = "first_name"
		vetData.CurrentValue = vet.FirstName
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новое имя врача:")
		h.bot.Send(msg)

	case "📞 Редактировать телефон":
		h.adminState[userID] = "vet_edit_field"
		vetData.Field = "phone"
		vetData.CurrentValue = vet.Phone
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новый телефон врача:")
		h.bot.Send(msg)

	case "📧 Редактировать email":
		h.adminState[userID] = "vet_edit_field"
		vetData.Field = "email"
		if vet.Email.Valid {
			vetData.CurrentValue = vet.Email.String
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новый email врача (или '-' для очистки):")
		h.bot.Send(msg)

	case "💼 Редактировать опыт":
		h.adminState[userID] = "vet_edit_field"
		vetData.Field = "experience_years"
		if vet.ExperienceYears.Valid {
			vetData.CurrentValue = strconv.FormatInt(vet.ExperienceYears.Int64, 10)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новый опыт работы в годах (или '-' для очистки):")
		h.bot.Send(msg)

	case "🎯 Редактировать специализации":
		h.adminState[userID] = "vet_edit_specializations"
		specs, err := h.db.GetSpecializationsByVetID(vet.ID)
		if err == nil && len(specs) > 0 {
			var specIDs []string
			for _, spec := range specs {
				specIDs = append(specIDs, strconv.Itoa(spec.ID))
			}
			vetData.Specializations = strings.Join(specIDs, ",")
		}

		// Показываем список специализаций
		specializations, err := h.db.GetAllSpecializations()
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении специализаций")
			h.bot.Send(msg)
			return
		}

		var sb strings.Builder
		sb.WriteString("🎯 Текущие специализации: ")
		if vetData.Specializations != "" {
			sb.WriteString(vetData.Specializations)
		} else {
			sb.WriteString("не указаны")
		}
		sb.WriteString("\n\nДоступные специализации:\n")

		for _, spec := range specializations {
			sb.WriteString(fmt.Sprintf("ID %d: %s\n", spec.ID, spec.Name))
		}

		sb.WriteString("\nВведите ID специализаций через запятую (например: 1,3,5):")

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
		h.bot.Send(msg)

	case "⚡ Изменить статус":
		h.adminState[userID] = "vet_toggle_active"
		newStatus := !vet.IsActive
		statusText := "активен"
		if !newStatus {
			statusText = "неактивен"
		}

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("✅ Подтвердить"),
				tgbotapi.NewKeyboardButton("❌ Отмена"),
			),
		)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Вы уверены, что хотите сделать врача %s %s?", vet.FirstName, statusText))
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)

	case "🗑️ Удалить врача":
		h.adminState[userID] = "vet_confirm_delete"
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("✅ Подтвердить удаление"),
				tgbotapi.NewKeyboardButton("❌ Отмена"),
			),
		)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("⚠️ *ВНИМАНИЕ!* \n\nВы собираетесь удалить врача %s %s.\nЭто действие нельзя отменить!\n\nПодтвердите удаление:", vet.FirstName, vet.LastName))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)

	case "🔙 Назад":
		h.handleBackButton(update)

	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте кнопки для управления")
		h.bot.Send(msg)
	}
}

// handleVetEditField обрабатывает ввод нового значения для поля врача
func (h *AdminHandlers) handleVetEditField(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	vetData, ok := h.tempData[userIDStr+"_vet_edit"].(*models.VetEditData)
	if !ok || vetData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные врача не найдены")
		h.bot.Send(msg)
		h.showVetList(update)
		return
	}

	// Обработка специальных значений
	if text == "-" {
		text = "" // Очистка поля
	}

	// Обновляем поле в базе данных
	err := h.updateVeterinarianField(vetData.VetID, vetData.Field, text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Ошибка при обновлении данных: %v", err))
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Данные успешно обновлены!")
		h.bot.Send(msg)
	}

	// Возвращаем в меню редактирования врача
	vet, err := h.db.GetVeterinarianByID(vetData.VetID)
	if err == nil {
		h.showVetEditMenu(update, vet)
	} else {
		h.showVetList(update)
	}
}

// handleVetEditSpecializations обрабатывает ввод специализаций врача
func (h *AdminHandlers) handleVetEditSpecializations(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	vetData, ok := h.tempData[userIDStr+"_vet_edit"].(*models.VetEditData)
	if !ok || vetData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные врача не найдены")
		h.bot.Send(msg)
		h.showVetList(update)
		return
	}

	// Валидация ID специализаций
	if text != "" && !h.isValidSpecializationIDs(text) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Неверный формат ID специализаций. Введите существующие ID через запятую")
		h.bot.Send(msg)
		return
	}

	// Обновляем специализации врача
	err := h.updateVeterinarianSpecializations(vetData.VetID, text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Ошибка при обновлении специализаций: %v", err))
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Специализации успешно обновлены!")
		h.bot.Send(msg)
	}

	// Возвращаем в меню редактирования врача
	vet, err := h.db.GetVeterinarianByID(vetData.VetID)
	if err == nil {
		h.showVetEditMenu(update, vet)
	} else {
		h.showVetList(update)
	}
}

// handleVetConfirmDelete обрабатывает подтверждение удаления врача
func (h *AdminHandlers) handleVetConfirmDelete(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	vetData, ok := h.tempData[userIDStr+"_vet_edit"].(*models.VetEditData)
	if !ok || vetData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные врача не найдены")
		h.bot.Send(msg)
		h.showVetList(update)
		return
	}

	if text == "✅ Подтвердить удаление" {
		err := h.deleteVeterinarian(vetData.VetID)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Ошибка при удалении врача: %v", err))
			h.bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Врач успешно удален!")
			h.bot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Удаление отменено")
		h.bot.Send(msg)
	}

	// Возвращаем к списку врачей
	h.showVetList(update)
}

// handleVetToggleActive обрабатывает изменение статуса врача
func (h *AdminHandlers) handleVetToggleActive(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	vetData, ok := h.tempData[userIDStr+"_vet_edit"].(*models.VetEditData)
	if !ok || vetData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные врача не найдены")
		h.bot.Send(msg)
		h.showVetList(update)
		return
	}

	if text == "✅ Подтвердить" {
		// Получаем текущего врача
		vet, err := h.db.GetVeterinarianByID(vetData.VetID)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных врача")
			h.bot.Send(msg)
			h.showVetList(update)
			return
		}

		// Меняем статус
		newStatus := !vet.IsActive
		err = h.updateVeterinarianField(vetData.VetID, "is_active", strconv.FormatBool(newStatus))
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Ошибка при изменении статуса: %v", err))
			h.bot.Send(msg)
		} else {
			statusText := "активен"
			if !newStatus {
				statusText = "неактивен"
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("✅ Статус врача изменен на: %s", statusText))
			h.bot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Изменение статуса отменено")
		h.bot.Send(msg)
	}

	// Возвращаем в меню редактирования врача
	vet, err := h.db.GetVeterinarianByID(vetData.VetID)
	if err == nil {
		h.showVetEditMenu(update, vet)
	} else {
		h.showVetList(update)
	}
}

// showClinicList показывает список клиник с возможностью выбора
func (h *AdminHandlers) showClinicList(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "clinic_list"

	clinics, err := h.db.GetAllClinics()
	if err != nil {
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
	sb.WriteString("🏥 *Список клиник:*\n\n")

	for i, clinic := range clinics {
		status := "✅"
		if !clinic.IsActive {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s %d. %s - %s\n", status, i+1, clinic.Name, clinic.Address))
	}

	sb.WriteString("\nВведите номер клиники для управления:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleClinicListSelection обрабатывает выбор клиники из списка
func (h *AdminHandlers) handleClinicListSelection(update tgbotapi.Update, text string) {
	// Парсим номер клиники
	index, err := strconv.Atoi(text)
	if err != nil || index < 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите корректный номер клиники")
		h.bot.Send(msg)
		return
	}

	clinics, err := h.db.GetAllClinics()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка клиник")
		h.bot.Send(msg)
		return
	}

	if index > len(clinics) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Клиника с таким номером не найдена")
		h.bot.Send(msg)
		return
	}

	clinic := clinics[index-1]
	h.showClinicEditMenu(update, clinic)
}

// showClinicEditMenu показывает меню редактирования клиники
func (h *AdminHandlers) showClinicEditMenu(update tgbotapi.Update, clinic *models.Clinic) {
	userID := update.Message.From.ID
	h.adminState[userID] = "clinic_edit_menu"

	// Сохраняем ID клиники во временные данные
	userIDStr := strconv.FormatInt(userID, 10)
	h.tempData[userIDStr+"_clinic_edit"] = &models.ClinicEditData{
		ClinicID: clinic.ID,
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏥 *Управление клиникой:* %s\n\n", clinic.Name))
	sb.WriteString(fmt.Sprintf("📍 Адрес: %s\n", clinic.Address))

	if clinic.Phone.Valid {
		sb.WriteString(fmt.Sprintf("📞 Телефон: %s\n", clinic.Phone.String))
	}

	if clinic.WorkingHours.Valid {
		sb.WriteString(fmt.Sprintf("🕐 Часы работы: %s\n", clinic.WorkingHours.String))
	}

	sb.WriteString("📊 Статус: ")
	if clinic.IsActive {
		sb.WriteString("✅ Активна\n")
	} else {
		sb.WriteString("❌ Неактивна\n")
	}

	sb.WriteString("\nВыберите действие:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✏️ Редактировать название"),
			tgbotapi.NewKeyboardButton("📍 Редактировать адрес"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📞 Редактировать телефон"),
			tgbotapi.NewKeyboardButton("🕐 Редактировать часы работы"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚡ Изменить статус"),
			tgbotapi.NewKeyboardButton("🗑️ Удалить клинику"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// handleClinicEditMenu обрабатывает выбор действия для клиники
func (h *AdminHandlers) handleClinicEditMenu(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	clinicData, ok := h.tempData[userIDStr+"_clinic_edit"].(*models.ClinicEditData)
	if !ok || clinicData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные клиники не найдены")
		h.bot.Send(msg)
		h.showClinicList(update)
		return
	}

	// Получаем актуальные данные клиники
	clinic, err := h.db.GetClinicByID(clinicData.ClinicID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных клиники")
		h.bot.Send(msg)
		h.showClinicList(update)
		return
	}

	switch text {
	case "✏️ Редактировать название":
		h.adminState[userID] = "clinic_edit_field"
		clinicData.Field = "name"
		clinicData.CurrentValue = clinic.Name
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новое название клиники:")
		h.bot.Send(msg)

	case "📍 Редактировать адрес":
		h.adminState[userID] = "clinic_edit_field"
		clinicData.Field = "address"
		clinicData.CurrentValue = clinic.Address
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новый адрес клиники:")
		h.bot.Send(msg)

	case "📞 Редактировать телефон":
		h.adminState[userID] = "clinic_edit_field"
		clinicData.Field = "phone"
		if clinic.Phone.Valid {
			clinicData.CurrentValue = clinic.Phone.String
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новый телефон клиники (или '-' для очистки):")
		h.bot.Send(msg)

	case "🕐 Редактировать часы работы":
		h.adminState[userID] = "clinic_edit_field"
		clinicData.Field = "working_hours"
		if clinic.WorkingHours.Valid {
			clinicData.CurrentValue = clinic.WorkingHours.String
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите новые часы работы клиники (или '-' для очистки):")
		h.bot.Send(msg)

	case "⚡ Изменить статус":
		h.adminState[userID] = "clinic_toggle_active"
		newStatus := !clinic.IsActive
		statusText := "активна"
		if !newStatus {
			statusText = "неактивна"
		}

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("✅ Подтвердить"),
				tgbotapi.NewKeyboardButton("❌ Отмена"),
			),
		)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Вы уверены, что хотите сделать клинику %s %s?", clinic.Name, statusText))
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)

	case "🗑️ Удалить клинику":
		h.adminState[userID] = "clinic_confirm_delete"
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("✅ Подтвердить удаление"),
				tgbotapi.NewKeyboardButton("❌ Отмена"),
			),
		)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("⚠️ *ВНИМАНИЕ!* \n\nВы собираетесь удалить клинику %s.\nЭто действие нельзя отменить!\n\nПодтвердите удаление:", clinic.Name))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)

	case "🔙 Назад":
		h.handleBackButton(update)

	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте кнопки для управления")
		h.bot.Send(msg)
	}
}

// handleClinicEditField обрабатывает ввод нового значения для поля клиники
func (h *AdminHandlers) handleClinicEditField(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	clinicData, ok := h.tempData[userIDStr+"_clinic_edit"].(*models.ClinicEditData)
	if !ok || clinicData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные клиники не найдены")
		h.bot.Send(msg)
		h.showClinicList(update)
		return
	}

	// Обработка специальных значений
	if text == "-" {
		text = "" // Очистка поля
	}

	// Обновляем поле в базе данных
	err := h.updateClinicField(clinicData.ClinicID, clinicData.Field, text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Ошибка при обновлении данных: %v", err))
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Данные успешно обновлены!")
		h.bot.Send(msg)
	}

	// Возвращаем в меню редактирования клиники
	clinic, err := h.db.GetClinicByID(clinicData.ClinicID)
	if err == nil {
		h.showClinicEditMenu(update, clinic)
	} else {
		h.showClinicList(update)
	}
}

// handleClinicConfirmDelete обрабатывает подтверждение удаления клиники
func (h *AdminHandlers) handleClinicConfirmDelete(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	clinicData, ok := h.tempData[userIDStr+"_clinic_edit"].(*models.ClinicEditData)
	if !ok || clinicData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные клиники не найдены")
		h.bot.Send(msg)
		h.showClinicList(update)
		return
	}

	if text == "✅ Подтвердить удаление" {
		err := h.deleteClinic(clinicData.ClinicID)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Ошибка при удалении клиники: %v", err))
			h.bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Клиника успешно удалена!")
			h.bot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Удаление отменено")
		h.bot.Send(msg)
	}

	// Возвращаем к списку клиник
	h.showClinicList(update)
}

// handleClinicToggleActive обрабатывает изменение статуса клиники
func (h *AdminHandlers) handleClinicToggleActive(update tgbotapi.Update, text string) {
	userID := update.Message.From.ID
	userIDStr := strconv.FormatInt(userID, 10)

	clinicData, ok := h.tempData[userIDStr+"_clinic_edit"].(*models.ClinicEditData)
	if !ok || clinicData == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: данные клиники не найдены")
		h.bot.Send(msg)
		h.showClinicList(update)
		return
	}

	if text == "✅ Подтвердить" {
		// Получаем текущую клинику
		clinic, err := h.db.GetClinicByID(clinicData.ClinicID)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных клиники")
			h.bot.Send(msg)
			h.showClinicList(update)
			return
		}

		// Меняем статус
		newStatus := !clinic.IsActive
		err = h.updateClinicField(clinicData.ClinicID, "is_active", strconv.FormatBool(newStatus))
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Ошибка при изменении статуса: %v", err))
			h.bot.Send(msg)
		} else {
			statusText := "активна"
			if !newStatus {
				statusText = "неактивна"
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("✅ Статус клиники изменен на: %s", statusText))
			h.bot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Изменение статуса отменено")
		h.bot.Send(msg)
	}

	// Возвращаем в меню редактирования клиники
	clinic, err := h.db.GetClinicByID(clinicData.ClinicID)
	if err == nil {
		h.showClinicEditMenu(update, clinic)
	} else {
		h.showClinicList(update)
	}
}

// startAddClinic начинает процесс добавления клиники
func (h *AdminHandlers) startAddClinic(update tgbotapi.Update) {
	// TODO: Реализовать добавление клиники
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Функция добавления клиники в разработке")
	h.bot.Send(msg)
}

// showSettings показывает настройки
func (h *AdminHandlers) showSettings(update tgbotapi.Update) {
	userCount, _ := h.getUserCount()
	activeVets, _ := h.getActiveVetCount()
	totalVets, _ := h.getTotalVetCount()
	activeClinics, _ := h.getActiveClinicCount()
	totalClinics, _ := h.getTotalClinicCount()

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf(`⚙️ *Настройки системы*

📊 Статистика:
• Пользователей: %d
• Врачей: %d/%d активных
• Клиник: %d/%d активных

Для изменения данных используйте админские функции или прямые SQL-запросы к базе данных.`,
			userCount, activeVets, totalVets, activeClinics, totalClinics))
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleStats показывает статистику бота
func (h *AdminHandlers) HandleStats(update tgbotapi.Update) {
	userCount, _ := h.getUserCount()
	activeVets, _ := h.getActiveVetCount()
	totalVets, _ := h.getTotalVetCount()
	activeClinics, _ := h.getActiveClinicCount()
	totalClinics, _ := h.getTotalClinicCount()
	requestCount, _ := h.getRequestCount()

	statsMsg := fmt.Sprintf(`📊 *Статистика бота*

👥 Пользователей: %d
👨‍⚕️ Врачей: %d/%d активных
🏥 Клиник: %d/%d активных
📞 Запросов: %d

Система работает стабильно ✅`, userCount, activeVets, totalVets, activeClinics, totalClinics, requestCount)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, statsMsg)
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// closeAdmin закрывает админскую панель
func (h *AdminHandlers) closeAdmin(update tgbotapi.Update) {
	userID := update.Message.From.ID

	// Очищаем все временные данные пользователя
	h.cleanTempData(userID)
	delete(h.adminState, userID)

	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Админская панель закрыта")
	msg.ReplyMarkup = removeKeyboard
	h.bot.Send(msg)
}

// ========== ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С БАЗОЙ ДАННЫХ ==========

// isValidSpecializationIDs проверяет валидность введенных ID специализаций
func (h *AdminHandlers) isValidSpecializationIDs(input string) bool {
	if input == "" {
		return true // Пустая строка допустима (очистка специализаций)
	}

	// Получаем максимальный ID специализации для проверки
	maxID, err := h.getMaxSpecializationID()
	if err != nil {
		log.Printf("Error getting max specialization ID: %v", err)
		return false
	}

	ids := strings.Split(input, ",")
	for _, idStr := range ids {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil || id <= 0 || id > maxID {
			return false
		}

		// Дополнительная проверка существования специализации в БД
		exists, err := h.db.SpecializationExists(id)
		if err != nil || !exists {
			return false
		}
	}
	return true
}

// getMaxSpecializationID возвращает максимальный ID специализации
func (h *AdminHandlers) getMaxSpecializationID() (int, error) {
	var maxID int
	err := h.db.GetDB().QueryRow("SELECT COALESCE(MAX(id), 0) FROM specializations").Scan(&maxID)
	return maxID, err
}

// addVeterinarian добавляет врача в базу данных
func (h *AdminHandlers) addVeterinarian(vet *models.Veterinarian, specsText string) error {
	// Добавляем врача в базу
	query := `INSERT INTO veterinarians (first_name, last_name, phone, is_active) 
	          VALUES ($1, $2, $3, $4) RETURNING id`

	err := h.db.GetDB().QueryRow(query, vet.FirstName, vet.LastName, vet.Phone, vet.IsActive).
		Scan(&vet.ID)
	if err != nil {
		return err
	}

	// Обрабатываем специализации
	if specsText != "" {
		specIDs := strings.Split(specsText, ",")
		log.Printf("Adding vet ID %d with specializations: %v", vet.ID, specIDs)

		for _, specIDStr := range specIDs {
			specID, err := strconv.Atoi(strings.TrimSpace(specIDStr))
			if err == nil && specID > 0 {
				// Проверяем существование специализации
				exists, err := h.db.SpecializationExists(specID)
				if err != nil {
					log.Printf("Error checking specialization %d: %v", specID, err)
					continue
				}

				if exists {
					// Добавляем связь врач-специализация
					_, err = h.db.GetDB().Exec(
						"INSERT INTO vet_specializations (vet_id, specialization_id) VALUES ($1, $2)",
						vet.ID, specID,
					)
					if err != nil {
						log.Printf("Error adding specialization %d: %v", specID, err)
					} else {
						log.Printf("Successfully added specialization %d for vet %d", specID, vet.ID)
					}
				} else {
					log.Printf("Specialization %d does not exist", specID)
				}
			}
		}
	}

	return nil
}

// updateVeterinarianField обновляет поле врача в базе данных
func (h *AdminHandlers) updateVeterinarianField(vetID int, field string, value string) error {
	var query string
	var err error

	switch field {
	case "first_name":
		query = "UPDATE veterinarians SET first_name = $1 WHERE id = $2"
		_, err = h.db.GetDB().Exec(query, value, vetID)
	case "phone":
		query = "UPDATE veterinarians SET phone = $1 WHERE id = $2"
		_, err = h.db.GetDB().Exec(query, value, vetID)
	case "email":
		if value == "" {
			query = "UPDATE veterinarians SET email = NULL WHERE id = $1"
			_, err = h.db.GetDB().Exec(query, vetID)
		} else {
			query = "UPDATE veterinarians SET email = $1 WHERE id = $2"
			_, err = h.db.GetDB().Exec(query, value, vetID)
		}
	case "experience_years":
		if value == "" {
			query = "UPDATE veterinarians SET experience_years = NULL WHERE id = $1"
			_, err = h.db.GetDB().Exec(query, vetID)
		} else {
			exp, convErr := strconv.ParseInt(value, 10, 64)
			if convErr != nil {
				return convErr
			}
			query = "UPDATE veterinarians SET experience_years = $1 WHERE id = $2"
			_, err = h.db.GetDB().Exec(query, exp, vetID)
		}
	case "is_active":
		active, convErr := strconv.ParseBool(value)
		if convErr != nil {
			return convErr
		}
		query = "UPDATE veterinarians SET is_active = $1 WHERE id = $2"
		_, err = h.db.GetDB().Exec(query, active, vetID)
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return err
}

// updateVeterinarianSpecializations обновляет специализации врача
func (h *AdminHandlers) updateVeterinarianSpecializations(vetID int, specsText string) error {
	// Удаляем все текущие специализации врача
	_, err := h.db.GetDB().Exec("DELETE FROM vet_specializations WHERE vet_id = $1", vetID)
	if err != nil {
		return err
	}

	// Добавляем новые специализации, если они указаны
	if specsText != "" {
		specIDs := strings.Split(specsText, ",")
		for _, specIDStr := range specIDs {
			specID, err := strconv.Atoi(strings.TrimSpace(specIDStr))
			if err == nil && specID > 0 {
				// Проверяем существование специализации
				exists, err := h.db.SpecializationExists(specID)
				if err == nil && exists {
					_, err = h.db.GetDB().Exec(
						"INSERT INTO vet_specializations (vet_id, specialization_id) VALUES ($1, $2)",
						vetID, specID,
					)
					if err != nil {
						log.Printf("Error adding specialization %d: %v", specID, err)
					}
				}
			}
		}
	}

	return nil
}

// deleteVeterinarian удаляет врача из базы данных
func (h *AdminHandlers) deleteVeterinarian(vetID int) error {
	// Удаляем связи с специализациями
	_, err := h.db.GetDB().Exec("DELETE FROM vet_specializations WHERE vet_id = $1", vetID)
	if err != nil {
		return err
	}

	// Удаляем расписание врача
	_, err = h.db.GetDB().Exec("DELETE FROM schedules WHERE vet_id = $1", vetID)
	if err != nil {
		return err
	}

	// Удаляем врача
	_, err = h.db.GetDB().Exec("DELETE FROM veterinarians WHERE id = $1", vetID)
	return err
}

// updateClinicField обновляет поле клиники в базе данных
func (h *AdminHandlers) updateClinicField(clinicID int, field string, value string) error {
	var query string
	var err error

	switch field {
	case "name":
		query = "UPDATE clinics SET name = $1 WHERE id = $2"
		_, err = h.db.GetDB().Exec(query, value, clinicID)
	case "address":
		query = "UPDATE clinics SET address = $1 WHERE id = $2"
		_, err = h.db.GetDB().Exec(query, value, clinicID)
	case "phone":
		if value == "" {
			query = "UPDATE clinics SET phone = NULL WHERE id = $1"
			_, err = h.db.GetDB().Exec(query, clinicID)
		} else {
			query = "UPDATE clinics SET phone = $1 WHERE id = $2"
			_, err = h.db.GetDB().Exec(query, value, clinicID)
		}
	case "working_hours":
		if value == "" {
			query = "UPDATE clinics SET working_hours = NULL WHERE id = $1"
			_, err = h.db.GetDB().Exec(query, clinicID)
		} else {
			query = "UPDATE clinics SET working_hours = $1 WHERE id = $2"
			_, err = h.db.GetDB().Exec(query, value, clinicID)
		}
	case "is_active":
		active, convErr := strconv.ParseBool(value)
		if convErr != nil {
			return convErr
		}
		query = "UPDATE clinics SET is_active = $1 WHERE id = $2"
		_, err = h.db.GetDB().Exec(query, active, clinicID)
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return err
}

// deleteClinic удаляет клинику из базы данных
func (h *AdminHandlers) deleteClinic(clinicID int) error {
	// Удаляем расписание, связанное с клиникой
	_, err := h.db.GetDB().Exec("DELETE FROM schedules WHERE clinic_id = $1", clinicID)
	if err != nil {
		return err
	}

	// Удаляем клинику
	_, err = h.db.GetDB().Exec("DELETE FROM clinics WHERE id = $1", clinicID)
	return err
}

// getStringTempData получает строковые данные из временного хранилища
func (h *AdminHandlers) getStringTempData(key string) string {
	if value, exists := h.tempData[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// ========== МЕТОДЫ ДЛЯ СТАТИСТИКИ ==========

func (h *AdminHandlers) getUserCount() (int, error) {
	query := "SELECT COUNT(*) FROM users"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getActiveVetCount() (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians WHERE is_active = true"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getTotalVetCount() (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getActiveClinicCount() (int, error) {
	query := "SELECT COUNT(*) FROM clinics WHERE is_active = true"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getTotalClinicCount() (int, error) {
	query := "SELECT COUNT(*) FROM clinics"
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
