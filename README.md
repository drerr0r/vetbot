# 🐾 VetBot - Telegram бот для поиска ветеринаров

Telegram бот на Go для поиска и управления контактами ветеринарных врачей.

## ✨ Возможности

- 🔍 Поиск врачей по специализации
- 📋 Просмотр полного списка врачей
- ⚡ Административная панель
- 🗃️ Хранение данных в PostgreSQL
- 🐳 Docker контейнеризация

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.21+
- PostgreSQL 15+
- Docker и Docker Compose (опционально)
- Telegram Bot Token от [@BotFather](https://t.me/BotFather)

### Установка

1. **Клонирование репозитория:**
   ```bash
   git clone https://github.com/drerr0r/vetbot.git
   cd vetbot
Настройка окружения:

bash
cp .env.example .env
# Отредактируйте .env файл, установите TELEGRAM_BOT_TOKEN
Установка зависимостей:

bash
make install
Запуск:

bash
make run
📋 Команды бота
/start - Начать работу

/help - Помощь по командам

/find [специализация] - Поиск врачей

/list - Список всех врачей

/admin - Административные функции

🛠 Разработка
bash
# Сборка
make build

# Тестирование
make test

# Запуск с Docker
make docker-up