package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"vetbot/internal/database"
	"vetbot/pkg/utils"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminHandlers содержит обработчики административных команд
type AdminHandlers struct {
	bot    *telegram.BotAPI
	db     *database.Database
	config *utils.Config
}

// NewAdminHandlers создает новый экземпляр административных обработчиков
func NewAdminHandlers(bot *telegram.BotAPI, db *database.Database, config *utils.Config) *AdminHandlers {
	return &AdminHandlers{
		bot:    bot,
		db:     db,
		config: config,
	}
}

// HandleAdminCommand обрабатывает административные команды
func (h *AdminHandlers) HandleAdminCommand(update telegram.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	// Проверяем права администратора
	isAdmin, err := h.db.IsAdmin(chatID)
	if err != nil {
		log.Printf("Error checking admin rights: %v", err)
		h.sendMessage(chatID, "❌ Ошибка проверки прав доступа")
		return
	}

	if !isAdmin {
		h.sendMessage(chatID, "⛔ У вас нет прав администратора")
		return
	}

	// Разбираем команду
	parts := strings.Fields(text)
	if len(parts) < 2 {
		h.showAdminHelp(chatID)
		return
	}

	subCommand := parts[1]

	switch subCommand {
	case "add":
		h.handleAddVet(chatID)
	case "edit":
		if len(parts) >= 3 {
			h.handleEditVet(chatID, parts[2])
		} else {
			h.sendMessage(chatID, "❌ Укажите ID врача для редактирования: /admin edit [id]")
		}
	case "delete":
		if len(parts) >= 3 {
			h.handleDeleteVet(chatID, parts[2])
		} else {
			h.sendMessage(chatID, "❌ Укажите ID врача для удаления: /admin delete [id]")
		}
	case "stats":
		h.handleStats(chatID)
	case "list":
		h.handleAdminList(chatID)
	default:
		h.showAdminHelp(chatID)
	}
}

// handleAddVet обрабатывает добавление нового врача
func (h *AdminHandlers) handleAddVet(chatID int64) {
	message := `👨‍⚕️ Добавление нового врача

Отправьте данные в формате:
Имя Фамилия
Специализация
Адрес
Телефон
Часы работы

Пример:
Иван Петров
Терапевт
ул. Центральная, 1
+7 (999) 123-45-67
09:00-18:00`

	h.sendMessage(chatID, message)

	// Здесь должна быть логика ожидания следующих сообщений с данными
	// В реальной реализации нужно использовать состояние бота (state machine)
	h.sendMessage(chatID, "⚠️ Функция добавления через многосообщенный ввод в разработке. Используйте прямой SQL для добавления врачей.")
}

// handleEditVet обрабатывает редактирование врача
func (h *AdminHandlers) handleEditVet(chatID int64, vetIDStr string) {
	// Парсим ID врача
	vetID, err := strconv.ParseInt(vetIDStr, 10, 64)
	if err != nil {
		h.sendMessage(chatID, "❌ Неверный формат ID врача")
		return
	}

	// Получаем данные врача
	vet, err := h.db.GetVeterinarianByID(vetID)
	if err != nil {
		h.sendMessage(chatID, fmt.Sprintf("❌ Врач с ID %d не найден", vetID))
		return
	}

	// Показываем текущие данные врача
	message := fmt.Sprintf(`✏️ Редактирование врача ID: %d

Текущие данные:
👨‍⚕️ Имя: %s
🎯 Специализация: %s
📍 Адрес: %s
📞 Телефон: %s
🕐 Часы работы: %s

Отправьте новые данные в формате:
имя=Новое Имя
специализация=Новая Специализация
адрес=Новый Адрес
телефон=Новый Телефон
часы=Новые Часы работы

Пример:
имя=Иван Иванов
специализация=Хирург`, vetID, vet.Name, vet.Specialty, vet.Address, vet.Phone, vet.WorkHours)

	h.sendMessage(chatID, message)
	h.sendMessage(chatID, "⚠️ Функция редактирования через сообщения в разработке. Используйте прямой SQL для изменений.")
}

// handleDeleteVet обрабатывает удаление врача
func (h *AdminHandlers) handleDeleteVet(chatID int64, vetIDStr string) {
	// Парсим ID врача
	vetID, err := strconv.ParseInt(vetIDStr, 10, 64)
	if err != nil {
		h.sendMessage(chatID, "❌ Неверный формат ID врача")
		return
	}

	// Получаем данные врача для подтверждения
	vet, err := h.db.GetVeterinarianByID(vetID)
	if err != nil {
		h.sendMessage(chatID, fmt.Sprintf("❌ Врач с ID %d не найден", vetID))
		return
	}

	// Показываем информацию о враче и запрашиваем подтверждение
	message := fmt.Sprintf(`🗑️ Подтверждение удаления врача:

👨‍⚕️ Имя: %s
🎯 Специализация: %s
📍 Адрес: %s
📞 Телефон: %s

Для подтверждения удаления отправьте: /confirm_delete %d
Для отмены отправьте: /cancel`, vet.Name, vet.Specialty, vet.Address, vet.Phone, vetID)

	h.sendMessage(chatID, message)
}

// handleStats показывает статистику бота
func (h *AdminHandlers) handleStats(chatID int64) {
	// Получаем всех врачей
	vets, err := h.db.GetAllVeterinarians()
	if err != nil {
		h.sendMessage(chatID, "❌ Ошибка получения статистики")
		return
	}

	// Получаем количество пользователей
	var userCount int
	err = h.db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		userCount = 0
	}

	var adminCount int
	err = h.db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = true").Scan(&adminCount)
	if err != nil {
		adminCount = 0
	}

	message := fmt.Sprintf(`📊 Статистика VetBot

👨‍⚕️ Врачей в базе: %d
👥 Пользователей: %d
⚙️ Администраторов: %d

💡 Система работает стабильно`, len(vets), userCount, adminCount)

	h.sendMessage(chatID, message)
}

// handleAdminList показывает список врачей с ID
func (h *AdminHandlers) handleAdminList(chatID int64) {
	// Получаем всех врачей из базы данных
	veterinarians, err := h.db.GetAllVeterinarians()
	if err != nil {
		log.Printf("Error getting veterinarians: %v", err)
		h.sendMessage(chatID, "❌ Ошибка при получении данных из базы")
		return
	}

	if len(veterinarians) == 0 {
		h.sendMessage(chatID, "📭 В базе данных нет врачей")
		return
	}

	// Форматируем список врачей с ID
	var message strings.Builder
	message.WriteString("👨‍⚕️ Список врачей (с ID):\n\n")

	for _, vet := range veterinarians {
		message.WriteString(fmt.Sprintf("🆔 ID: %d\n", vet.ID))
		message.WriteString(fmt.Sprintf("   👨‍⚕️ Имя: %s\n", vet.Name))
		message.WriteString(fmt.Sprintf("   🎯 Специализация: %s\n", vet.Specialty))
		message.WriteString(fmt.Sprintf("   📍 Адрес: %s\n", vet.Address))
		message.WriteString(fmt.Sprintf("   📞 Телефон: %s\n", vet.Phone))
		message.WriteString(fmt.Sprintf("   🕐 Часы работы: %s\n\n", vet.WorkHours))
	}

	message.WriteString("💡 Используйте ID для команд редактирования и удаления")

	h.sendMessage(chatID, message.String())
}

// showAdminHelp показывает справку по административным командам
func (h *AdminHandlers) showAdminHelp(chatID int64) {
	message := `⚙️ Справка по административным командам:

/admin add - добавить нового врача
/admin edit [id] - редактировать врача
/admin delete [id] - удалить врача
/admin stats - статистика бота
/admin list - список всех врачей (с ID)

Примеры:
/admin edit 1 - редактировать врача с ID 1
/admin delete 2 - удалить врача с ID 2

💡 Для просмотра ID врачей используйте: /admin list`

	h.sendMessage(chatID, message)
}

// sendMessage отправляет сообщение в Telegram чат
func (h *AdminHandlers) sendMessage(chatID int64, text string) {
	msg := telegram.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
