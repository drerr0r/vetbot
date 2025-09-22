package main

import (
	"log"
	"os"
	"os/signal"
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

	// Инициализируем бота Telegram
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.Debug = config.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Инициализируем базу данных
	db, err := database.NewDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Создаем основной обработчик
	mainHandler := handlers.NewMainHandler(bot, db)

	// Настраиваем вебхук или long polling
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
			go mainHandler.HandleUpdate(update)
		case <-sigChan:
			log.Println("Shutting down bot...")
			return
		}
	}
}
