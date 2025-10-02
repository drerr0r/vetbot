🐾 VetBot - Ветеринарный Telegram бот
Технологический стек: Go, PostgreSQL, Docker, Railway
Сложность деплоя: низкая (15-20 минут)

🚀 Быстрый деплой на Railway
Railway — это PaaS-платформа (как Heroku), которая автоматизирует деплой контейнеров и управление инфраструктурой.

📋 Предварительные требования
Telegram Bot Token от @BotFather

Ваш Telegram ID (можно получить через @userinfobot)

GitHub аккаунт (для хранения кода)

🎯 Пошаговая настройка
1. Подготовка репозитория
bash
# Клонируйте и настройте свой репозиторий
git clone https://github.com/drerr0r/vetbot.git
cd vetbot
# Настройте remote на ваш репозиторий
git remote set-url origin https://github.com/ВАШ_АККАУНТ/vetbot.git
git push -u origin main
Или через Fork: Просто нажмите "Fork" в интерфейсе GitHub.

2. Создание проекта в Railway
Регистрация: https://railway.app → Sign in with GitHub

Создание проекта: "New Project" → "Deploy from GitHub repo"

Выбор репозитория: Авторизуйте доступ и выберите ваш vetbot репозиторий

Что происходит: Railway автоматически определяет Dockerfile и начинает сборку.

3. Настройка базы данных
В панели Railway: "New" → "Database" → PostgreSQL

Дождитесь создания (1-2 минуты)

Перейдите в базу → "Connect" → скопируйте Postgres Connection URL

Важно: Этот URL будет автоматически доступен приложению как DATABASE_URL.

4. Конфигурация приложения
В Railway Dashboard перейдите в раздел Variables и добавьте:

Переменная	Значение	Описание
TELEGRAM_TOKEN	ваш_токен_от_BotFather	Токен бота Telegram
ADMIN_IDS	ваш_telegram_id	Ваш ID для админ-панели
DEBUG	false	Режим отладки
Как получить:

Telegram Token: /newbot в @BotFather

Telegram ID: напишите @userinfobot

5. Применение миграций БД
После деплоя нужно создать структуру БД:

bash
# Установите Railway CLI
npm install -g @railway/cli

# Авторизация и подключение
railway login
railway link

# Применение миграций
railway run psql $DATABASE_URL -f migrations/001_init.sql
Альтернативно: Через Railway Dashboard → Database → Query → выполните SQL из migrations/001_init.sql

6. Проверка работоспособности
Мониторинг логов:

bash
railway logs
Ожидаемый вывод:

text
Authorized on account YourBotName
Database connected successfully
Database columns check completed successfully
Bot started. Press Ctrl+C to stop.
Тестирование: Найдите бота в Telegram, команда /start должна работать.

📊 Загрузка данных из Excel/CSV файлов
🔧 Формат файлов для загрузки ветеринаров
Требуемые колонки:

text
Имя	Фамилия	Телефон	Email	Опыт работы	Описание	Специализации	Город	Регион
Пример заполнения:

text
Иван	Петров	+79161234567	ivan.petrov@mail.ru	5	Опытный хирург	Хирургия|Терапия	Москва	Московская область
Мария	Сидорова	+79031234568	maria.s@mail.ru	3	Специалист по мелким животным	Терапия|Стоматология	Санкт-Петербург	Ленинградская область
📝 Форматы файлов
Excel (.xlsx, .xls)

CSV (.csv) с разделителем табуляции или запятой

🔄 Отличие от локальной разработки
Важно: При деплое на Railway .env файл не используется. Вместо этого:

Локальная разработка	Railway деплой
Конфигурация через .env файл	Environment Variables в панели
База данных - ручная настройка	Автоматический DATABASE_URL
Миграции - ручной запуск	Автоматически при деплое
Не нужно:

Заливать .env файл в репозиторий

Редактировать .env для продакшена

Вручную настраивать подключение к БД

🔧 Архитектура и конфигурация
Docker-образ
dockerfile
# Многостадийная сборка Go приложения
FROM golang:1.21-alpine AS builder
FROM alpine:latest

# Автоматические миграции при запуске
CMD ./scripts/apply_migrations_railway.sh && ./vetbot
Переменные окружения
env
# Для Railway (Environment Variables в панели)
TELEGRAM_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
ADMIN_IDS=123456789
DEBUG=false
# DATABASE_URL создается автоматически
Структура БД
Миграции автоматически создают:

users - пользователи бота

clinics - ветеринарные клиники

veterinarians - врачи

specializations - специализации

schedules - расписания

🛠 Локальная разработка
1. Настройка окружения
bash
# Клонирование репозитория
git clone https://github.com/ВАШ_АККАУНТ/vetbot.git
cd vetbot

# Копирование и настройка .env файла
cp .env.example .env
2. Редактирование .env файла
env
# .env файл для локальной разработки
TELEGRAM_TOKEN=your_telegram_bot_token_here
DATABASE_URL=postgresql://user:password@localhost:5432/vetbot?sslmode=disable
ADMIN_IDS=your_telegram_id_here
DEBUG=true
3. Запуск базы данных
bash
# Запуск PostgreSQL через Docker
docker-compose up -d postgres

# Или используйте свою локальную PostgreSQL
# Убедитесь что БД существует и доступна по DATABASE_URL
4. Применение миграций (локально)
bash
# Создание структуры БД
psql $DATABASE_URL -f migrations/001_init.sql

# Или через Railway CLI если настроен
railway run psql $DATABASE_URL -f migrations/001_init.sql
5. Запуск приложения
bash
# Способ 1: Напрямую через Go
go run ./cmd/vetbot

# Способ 2: Через Docker
docker-compose up vetbot

# Способ 3: Сборка и запуск
go build -o vetbot ./cmd/vetbot
./vetbot
6. Тестирование локально
Бот должен запуститься и показать: Authorized on account YourBotName

Проверьте подключение к базе: Database connected successfully

Протестируйте команды бота в Telegram

❌ Диагностика проблем
Бот не запускается
bash
# Проверка логов
railway logs

# Проверка переменных
railway vars list
Частые причины:

Неверный TELEGRAM_TOKEN

Отсутствует DATABASE_URL

Не применены миграции БД

Ошибки базы данных
bash
# Проверка подключения к БД
railway run psql $DATABASE_URL -c "SELECT version();"

# Принудительное применение миграций
railway run psql $DATABASE_URL -f migrations/001_init.sql
Бот не отвечает
Проверьте что бот активирован через @BotFather

Убедитесь что в логах нет ошибок авторизации

Локальные проблемы
bash
# Проверка .env файла
cat .env

# Проверка подключения к локальной БД
psql $DATABASE_URL -c "\l"

# Пересборка Docker образов
docker-compose build --no-cache
🔄 Обновление и мониторинг
Обновление кода
bash
git pull origin main
# Railway автоматически перезапустит контейнер
Мониторинг
Логи: Railway Dashboard → Logs

Метрики: Railway Dashboard → Metrics

База данных: Railway Dashboard → Database → Metrics

Масштабирование
Автоматическое масштабирование включено по умолчанию

Ручное изменение ресурсов: Settings → Resources

📞 Техническая поддержка
Проверьте перед обращением:

Логи в Railway Dashboard

Применены миграции БД

Корректные значения переменных окружения

Для отладки предоставьте:

Вывод railway logs

Содержимое railway vars list (без чувствительных данных)

Результат проверки БД: railway run psql $DATABASE_URL -c "\dt"

⏱️ Время настройки: 15-20 минут
💸 Стоимость: Бесплатно на Railway
🚀 Результат: Production-ready бот с авто-масштабированием