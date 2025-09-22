package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"vetbot/internal/database"
	"vetbot/internal/handlers"
	"vetbot/pkg/utils"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загрузка конфигурации
	config := utils.LoadConfig()

	// Валидация конфигурации
	if err := config.Validate(); err != nil {
		log.Fatalf("❌ Ошибка конфигурации: %v", err)
	}

	log.Println("✅ Конфигурация загружена успешно")

	// Инициализация базы данных
	db, err := database.InitDB(config)
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	log.Println("✅ База данных подключена успешно")

	// Запуск миграций
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("❌ Ошибка миграций: %v", err)
	}

	log.Println("✅ Миграции выполнены успешно")

	// Инициализация Telegram бота
	bot, err := telegram.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации бота: %v", err)
	}

	log.Printf("✅ Бот инициализирован: @%s", bot.Self.UserName)

	// Создание обработчиков
	botHandlers := handlers.NewBotHandlers(bot, db, config)
	adminHandlers := handlers.NewAdminHandlers(bot, db, config)

	// Создаем главный обработчик
	mainHandler := handlers.NewMainHandler(botHandlers, adminHandlers)

	// Настройка обновлений
	u := telegram.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("🚀 Бот запущен и готов к работе!")
	log.Println("💡 Используйте /start в Telegram для начала работы")

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Главный цикл обработки сообщений
	for {
		select {
		case update := <-updates:
			go mainHandler.HandleUpdate(update)
		case <-sigChan:
			log.Println("🛑 Получен сигнал завершения, останавливаем бота...")
			return
		}
	}
}
