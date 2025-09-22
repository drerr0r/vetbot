package handlers

import (
	"fmt"
	"log"
	"strings"

	"vetbot/internal/database"
	"vetbot/pkg/utils"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotHandlers содержит обработчики команд Telegram бота
type BotHandlers struct {
	bot    *telegram.BotAPI
	db     *database.Database
	config *utils.Config
}

// NewBotHandlers создает новый экземпляр обработчиков бота
func NewBotHandlers(bot *telegram.BotAPI, db *database.Database, config *utils.Config) *BotHandlers {
	return &BotHandlers{
		bot:    bot,
		db:     db,
		config: config,
	}
}

// HandleUpdate обрабатывает входящие обновления от Telegram
func (h *BotHandlers) HandleUpdate(update telegram.Update) {
	// Игнорируем обновления без сообщений
	if update.Message == nil {
		return
	}

	// Регистрируем пользователя если он новый
	h.RegisterUser(update.Message.From.UserName, update.Message.Chat.ID)

	// Обрабатываем команды
	switch {
	case update.Message.IsCommand():
		h.HandleCommand(update)
	case strings.HasPrefix(update.Message.Text, "/find"):
		h.HandleFindCommand(update)
	default:
		h.HandleDefaultMessage(update)
	}
}

// HandleCommand обрабатывает команды бота
func (h *BotHandlers) HandleCommand(update telegram.Update) {
	command := update.Message.Command()
	chatID := update.Message.Chat.ID

	log.Printf("Received command: %s from chat ID: %d", command, chatID)

	switch command {
	case "start":
		h.HandleStartCommand(chatID)
	case "help":
		h.HandleHelpCommand(chatID)
	case "list":
		h.HandleListCommand(chatID)
	case "find":
		h.HandleFindWithEmptyQuery(chatID)
	case "admin":
		// Административные команды обрабатываются отдельно
		h.SendMessage(chatID, "⚙️ Административные команды обрабатываются через /admin")
	default:
		h.HandleUnknownCommand(chatID)
	}
}

// HandleStartCommand обрабатывает команду /start
func (h *BotHandlers) HandleStartCommand(chatID int64) {
	message := `🐾 Добро пожаловать в VetBot!

Я помогу вам найти контакты ветеринарных врачей.

📋 Доступные команды:
/start - начать работу с ботом
/help - показать справку по командам
/find - найти врача по специализации
/list - показать всех врачей

💡 Пример использования:
/find терапевт - найти всех терапевтов
/find хирург - найти хирургов`

	h.SendMessage(chatID, message)
}

// HandleHelpCommand обрабатывает команду /help
func (h *BotHandlers) HandleHelpCommand(chatID int64) {
	message := `📋 Помощь по командам VetBot:

🔍 Поиск врачей:
/find [специализация] - поиск врачей по специализации
Пример: /find терапевт

📋 Просмотр данных:
/list - показать всех врачей в базе

💡 Советы:
• Используйте частичные названия специализаций
• Регистр не имеет значения при поиске`

	h.SendMessage(chatID, message)
}

// HandleListCommand обрабатывает команду /list
func (h *BotHandlers) HandleListCommand(chatID int64) {
	// Получаем всех врачей из базы данных
	veterinarians, err := h.db.GetAllVeterinarians()
	if err != nil {
		log.Printf("Error getting veterinarians: %v", err)
		h.SendMessage(chatID, "❌ Ошибка при получении данных из базы")
		return
	}

	if len(veterinarians) == 0 {
		h.SendMessage(chatID, "📭 В базе данных нет врачей")
		return
	}

	// Форматируем список врачей
	var message strings.Builder
	message.WriteString("👨‍⚕️ Список всех ветеринарных врачей:\n\n")

	for i, vet := range veterinarians {
		message.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, vet.Name, vet.Specialty))
		message.WriteString(fmt.Sprintf("   📍 Адрес: %s\n", vet.Address))
		message.WriteString(fmt.Sprintf("   📞 Телефон: %s\n", vet.Phone))
		message.WriteString(fmt.Sprintf("   🕐 Часы работы: %s\n\n", vet.WorkHours))
	}

	h.SendMessage(chatID, message.String())
}

// HandleFindCommand обрабатывает команду поиска /find
func (h *BotHandlers) HandleFindCommand(update telegram.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	// Извлекаем поисковый запрос из команды
	query := strings.TrimSpace(strings.TrimPrefix(text, "/find"))
	if query == "" {
		h.HandleFindWithEmptyQuery(chatID)
		return
	}

	// Ищем врачей по специализации
	veterinarians, err := h.db.FindVeterinariansBySpecialty(query)
	if err != nil {
		log.Printf("Error searching veterinarians: %v", err)
		h.SendMessage(chatID, "❌ Ошибка при поиске в базе данных")
		return
	}

	if len(veterinarians) == 0 {
		message := fmt.Sprintf("🔍 По запросу \"%s\" ничего не найдено\n\nПопробуйте другой запрос или используйте /list для просмотра всех врачей", query)
		h.SendMessage(chatID, message)
		return
	}

	// Форматируем результаты поиска
	var message strings.Builder
	message.WriteString(fmt.Sprintf("🔍 Результаты поиска по запросу \"%s\":\n\n", query))

	for i, vet := range veterinarians {
		message.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, vet.Name, vet.Specialty))
		message.WriteString(fmt.Sprintf("   📍 Адрес: %s\n", vet.Address))
		message.WriteString(fmt.Sprintf("   📞 Телефон: %s\n", vet.Phone))
		message.WriteString(fmt.Sprintf("   🕐 Часы работы: %s\n\n", vet.WorkHours))
	}

	h.SendMessage(chatID, message.String())
}

// HandleFindWithEmptyQuery обрабатывает команду /find без параметров
func (h *BotHandlers) HandleFindWithEmptyQuery(chatID int64) {
	message := `🔍 Поиск ветеринарных врачей

Используйте команду в формате:
/find [специализация]

Примеры:
/find терапевт - поиск терапевтов
/find хирург - поиск хирургов
/find стоматолог - поиск стоматологов

💡 Вы также можете использовать частичные совпадения:
/find тер - найдет терапевтов
/find хир - найдет хирургов`

	h.SendMessage(chatID, message)
}

// HandleUnknownCommand обрабатывает неизвестные команды
func (h *BotHandlers) HandleUnknownCommand(chatID int64) {
	message := `❌ Неизвестная команда

Доступные команды:
/start - начать работу
/help - помощь по командам
/find - поиск врачей
/list - список всех врачей

Введите /help для подробной справки`

	h.SendMessage(chatID, message)
}

// HandleDefaultMessage обрабатывает обычные сообщения (не команды)
func (h *BotHandlers) HandleDefaultMessage(update telegram.Update) {
	chatID := update.Message.Chat.ID
	message := `💬 Я понимаю только команды

Используйте следующие команды:
/start - начать работу
/help - помощь по командам
/find - поиск врачей
/list - список всех врачей

Введите /help для подробной справки`

	h.SendMessage(chatID, message)
}

// RegisterUser регистрирует нового пользователя в системе
func (h *BotHandlers) RegisterUser(username string, chatID int64) {
	// Проверяем, существует ли пользователь
	exists, err := h.db.UserExists(chatID)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		return
	}

	if !exists {
		// Определяем, является ли пользователь администратором
		isAdmin := chatID == h.config.AdminChatID

		// Создаем нового пользователя
		err := h.db.CreateUser(username, chatID, isAdmin)
		if err != nil {
			log.Printf("Error creating user: %v", err)
		} else {
			log.Printf("New user registered: %s (chat ID: %d, admin: %v)", username, chatID, isAdmin)
		}
	}
}

// SendMessage отправляет сообщение в Telegram чат
func (h *BotHandlers) SendMessage(chatID int64, text string) {
	msg := telegram.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
