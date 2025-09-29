package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/drerr0r/vetbot/internal/database"
	"github.com/drerr0r/vetbot/internal/handlers"
	"github.com/drerr0r/vetbot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загружаем конфигурацию
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Printf("Telegram token loaded: %s...", maskToken(config.TelegramToken))
	log.Printf("Database URL loaded: %s", maskDBPassword(config.DatabaseURL))
	log.Printf("Debug mode: %t", config.Debug)
	log.Printf("Admin IDs: %v", config.AdminIDs)

	// Инициализируем бота Telegram
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.Debug = config.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Инициализируем базу данных
	db, err := database.New(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// ДОБАВЛЕНО: Добавляем отсутствующие колонки в базу данных
	log.Println("Checking and adding missing database columns...")
	err = db.AddMissingColumns()
	if err != nil {
		log.Printf("Warning: could not add missing columns: %v", err)
	} else {
		log.Println("Database columns check completed successfully")
	}

	// Создаем адаптер для бота
	botAdapter := handlers.NewTelegramBotAdapter(bot)

	// Создаем основной обработчик
	mainHandler := handlers.NewMainHandler(botAdapter, db, config)

	// Настраиваем long polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Обрабатываем сигналы для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Bot started. Press Ctrl+C to stop.")

	// Основной цикл обработки сообщений
	for {
		select {
		case update := <-updates:
			// Обрабатываем обновление в той же горутине для сохранения порядка сообщений
			mainHandler.HandleUpdate(update)
		case <-sigChan:
			log.Println("Shutting down bot gracefully...")
			return
		}
	}
}

// maskToken маскирует токен для безопасного логирования
func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:10] + "..."
}

// maskDBPassword маскирует пароль в URL базы данных для логирования
func maskDBPassword(dbURL string) string {
	parts := strings.Split(dbURL, "@")
	if len(parts) != 2 {
		return dbURL
	}

	authPart := parts[0]
	if strings.Contains(authPart, ":") {
		authParts := strings.Split(authPart, ":")
		if len(authParts) >= 3 { // postgres://user:password@
			// Находим позицию пароля
			userPass := strings.Split(authParts[2], "@")
			if len(userPass) >= 2 {
				// Маскируем пароль
				return authParts[0] + "://" + authParts[1] + ":***@" + parts[1]
			}
		} else if len(authParts) == 2 && strings.Contains(authParts[0], "//") {
			// Формат user:password@host
			protocolParts := strings.Split(authParts[0], "//")
			if len(protocolParts) == 2 {
				return protocolParts[0] + "//" + protocolParts[1] + ":***@" + parts[1]
			}
		}
	}

	return dbURL
}
