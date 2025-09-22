# 🐾 VetBot - Ветеринарный Telegram бот

VetBot помогает пользователям находить ветеринарных врачей по специализации и расписанию.

## ✨ Возможности

- 🔍 Поиск врачей по специализации
- 🕐 Поиск по расписанию и дням недели
- 📞 Контакты врачей и клиник
- 👥 Административная панель
- 📊 Статистика использования

## 🚀 Быстрый старт

### Требования

- Go 1.21+
- PostgreSQL 12+
- Telegram Bot Token

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/drerr0r/vetbot.git
cd vetbot
Настройте окружение:

bash
cp .env.example .env
# Отредактируйте .env файл
Запустите базу данных:

bash
docker-compose up -d postgres
Примените миграции:

bash
make migrate-up
Запустите приложение:

bash
make run
Docker запуск
bash
# Сборка и запуск
make docker-build
make docker-run

# Или используйте Docker Compose
make compose-up
📁 Структура проекта
text
vetbot/
├── cmd/vetbot/          # Точка входа
├── internal/            # Внутренние пакеты
│   ├── database/        # Работа с БД
│   ├── handlers/        # Обработчики Telegram
│   └── models/          # Модели данных
├── migrations/          # Миграции БД
├── pkg/utils/          # Утилиты
└── scripts/            # Скрипты развертывания
⚙️ Конфигурация
Создайте файл .env:

env
TELEGRAM_TOKEN=your_bot_token_here
DATABASE_URL=postgres://user:pass@localhost:5432/vetbot
DEBUG=false
ADMIN_IDS=123456789
🗄️ База данных
Проект использует PostgreSQL. Основные таблицы:

veterinarians - Врачи

specializations - Специализации

clinics - Клиники

schedules - Расписание

users - Пользователи бота

🤖 Команды бота
/start - Начало работы

/specializations - Список специализаций

/search - Поиск врача

/clinics - Список клиник

/help - Помощь

/admin - Админ-панель (только для админов)

🛠️ Разработка
bash
# Форматирование кода
make fmt

# Тестирование
make test

# Линтинг
make lint

# Запуск в режиме разработки
make dev