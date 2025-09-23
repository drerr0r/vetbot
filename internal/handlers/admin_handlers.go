package handlers

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/drerr0r/vetbot/internal/database"
	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminHandlers содержит обработчики для административных функций
type AdminHandlers struct {
	bot        *tgbotapi.BotAPI
	db         *database.Database
	adminState map[int64]string  // Хранит состояние админской сессии
	tempData   map[string]string // Хранит временные данные (ключ: "userID_field", значение: данные)
}

// NewAdminHandlers создает новый экземпляр AdminHandlers
func NewAdminHandlers(bot *tgbotapi.BotAPI, db *database.Database) *AdminHandlers {
	return &AdminHandlers{
		bot:        bot,
		db:         db,
		adminState: make(map[int64]string),
		tempData:   make(map[string]string),
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
			tgbotapi.NewKeyboardButton("📊 Статистика"),
			tgbotapi.NewKeyboardButton("⚙️ Настройки"),
		),
		tgbotapi.NewKeyboardButtonRow(
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
	case "add_vet_name":
		h.handleAddVetName(update, text)
	case "add_vet_phone":
		h.handleAddVetPhone(update, text)
	case "add_vet_specializations":
		h.handleAddVetSpecializations(update, text)
	default:
		h.handleMainMenu(update, text)
	}
}

// handleBackButton обрабатывает кнопку "Назад"
func (h *AdminHandlers) handleBackButton(update tgbotapi.Update) {
	userID := update.Message.From.ID

	// Определяем текущее состояние и возвращаемся на уровень выше
	switch h.adminState[userID] {
	case "vet_management", "clinic_management":
		// Возврат из подменю в главное меню
		h.adminState[userID] = "main_menu"
		h.HandleAdmin(update)
	case "add_vet_name", "add_vet_phone", "add_vet_specializations":
		// Возврат из процесса добавления врача в меню управления врачами
		h.adminState[userID] = "vet_management"

		// Очищаем временные данные
		userIDStr := strconv.FormatInt(userID, 10)
		delete(h.tempData, userIDStr+"_name")
		delete(h.tempData, userIDStr+"_phone")

		h.showVetManagement(update)
	default:
		// По умолчанию возвращаем в главное меню
		h.adminState[userID] = "main_menu"
		h.HandleAdmin(update)
	}
}

// handleMainMenu обрабатывает главное меню админки
func (h *AdminHandlers) handleMainMenu(update tgbotapi.Update, text string) {
	switch text {
	case "👥 Управление врачами":
		h.showVetManagement(update)
	case "🏥 Управление клиниками":
		h.showClinicManagement(update)
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

// handleVetManagement обрабатывает меню управления врачами
func (h *AdminHandlers) handleVetManagement(update tgbotapi.Update, text string) {
	switch text {
	case "➕ Добавить врача":
		h.startAddVet(update)
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

// handleClinicManagement обрабатывает меню управления клиниками
func (h *AdminHandlers) handleClinicManagement(update tgbotapi.Update, text string) {
	switch text {
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

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить врача"),
			tgbotapi.NewKeyboardButton("📋 Список врачей"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"👥 *Управление врачами*\n\nВыберите действие:")
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
	name := h.tempData[userIDStr+"_name"]
	phone := h.tempData[userIDStr+"_phone"]

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
	delete(h.tempData, userIDStr+"_name")
	delete(h.tempData, userIDStr+"_phone")

	// Возвращаем в меню управления врачами
	h.adminState[userID] = "vet_management"
	h.showVetManagement(update)
}

// isValidSpecializationIDs проверяет валидность введенных ID специализаций
func (h *AdminHandlers) isValidSpecializationIDs(input string) bool {
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

	return nil
}

// showVetList показывает список врачей
func (h *AdminHandlers) showVetList(update tgbotapi.Update) {
	// Получаем всех врачей через существующие методы
	specializations, err := h.db.GetAllSpecializations()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("👥 *Список врачей:*\n\n")

	for _, spec := range specializations {
		vets, err := h.db.GetVeterinariansBySpecialization(spec.ID)
		if err != nil {
			continue
		}

		if len(vets) > 0 {
			sb.WriteString(fmt.Sprintf("🏥 *%s:*\n", spec.Name))
			for _, vet := range vets {
				sb.WriteString(fmt.Sprintf("• %s %s - %s\n", vet.FirstName, vet.LastName, vet.Phone))
			}
			sb.WriteString("\n")
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// showClinicList показывает список клиник
func (h *AdminHandlers) showClinicList(update tgbotapi.Update) {
	clinics, err := h.db.GetAllClinics()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении клиник")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏥 *Список клиник:*\n\n")

	for i, clinic := range clinics {
		sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, clinic.Name))
		sb.WriteString(fmt.Sprintf("   Адрес: %s\n", clinic.Address))
		if clinic.Phone.Valid {
			sb.WriteString(fmt.Sprintf("   Телефон: %s\n", clinic.Phone.String))
		}
		if clinic.WorkingHours.Valid {
			sb.WriteString(fmt.Sprintf("   Часы работы: %s\n", clinic.WorkingHours.String))
		}
		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// showClinicManagement показывает меню управления клиниками
func (h *AdminHandlers) showClinicManagement(update tgbotapi.Update) {
	userID := update.Message.From.ID
	h.adminState[userID] = "clinic_management"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Список клиник"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)

	clinics, err := h.db.GetAllClinics()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении клиник")
		h.bot.Send(msg)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏥 *Управление клиниками*\n\n")
	sb.WriteString(fmt.Sprintf("Всего клиник: %d\n\n", len(clinics)))
	sb.WriteString("Выберите действие:")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// showSettings показывает настройки
func (h *AdminHandlers) showSettings(update tgbotapi.Update) {
	userCount, _ := h.getUserCount()
	vetCount, _ := h.getVetCount()
	clinicCount, _ := h.getClinicCount()

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf(`⚙️ *Настройки системы*

📊 Статистика:
• Пользователей: %d
• Врачей: %d
• Клиник: %d

Для изменения данных используйте прямые SQL-запросы к базе данных.`, userCount, vetCount, clinicCount))
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// HandleStats показывает статистику бота
func (h *AdminHandlers) HandleStats(update tgbotapi.Update) {
	userCount, _ := h.getUserCount()
	vetCount, _ := h.getVetCount()
	requestCount, _ := h.getRequestCount()

	statsMsg := fmt.Sprintf(`📊 *Статистика бота*

👥 Пользователей: %d
👨‍⚕️ Врачей в базе: %d
📞 Запросов: %d

Система работает стабильно ✅`, userCount, vetCount, requestCount)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, statsMsg)
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// closeAdmin закрывает админскую панель
func (h *AdminHandlers) closeAdmin(update tgbotapi.Update) {
	userID := update.Message.From.ID

	// Очищаем все временные данные пользователя
	userIDStr := strconv.FormatInt(userID, 10)
	delete(h.adminState, userID)
	delete(h.tempData, userIDStr+"_name")
	delete(h.tempData, userIDStr+"_phone")

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

func (h *AdminHandlers) getVetCount() (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians WHERE is_active = true"
	var count int
	err := h.db.GetDB().QueryRow(query).Scan(&count)
	return count, err
}

func (h *AdminHandlers) getClinicCount() (int, error) {
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
