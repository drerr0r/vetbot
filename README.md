🐾 VetBot - Ветеринарный Telegram бот
VetBot помогает пользователям находить ветеринарных врачей по специализации и расписанию.

✨ Возможности
🔍 Поиск врачей по специализации
🕐 Поиск по расписанию и дням недели
📞 Контакты врачей и клиник
👥 Административная панель
📊 Статистика использования

## 🚀 Быстрый старт на Railway

### 📋 Предварительные требования
- **GitHub аккаунт** - для хранения кода
- **Telegram Bot Token** от [@BotFather](https://t.me/BotFather)
- **Ваш Telegram ID** (можно получить через [@userinfobot](https://t.me/userinfobot))

---

## 🎯 Шаг 1: Создать свой репозиторий

### Вариант A: Fork репозитория (рекомендуется)
1. Перейдите в оригинальный репозиторий VetBot на GitHub
2. Нажмите кнопку **"Fork"** в правом верхнем углу
3. Выберите ваш аккаунт как место назначения
4. Дождитесь завершения процесса

### Вариант B: Ручное копирование
```bash
# Клонируйте оригинальный репозиторий
git clone https://github.com/drerr0r/vetbot.git
cd vetbot

# Создайте новый репозиторий на GitHub
# Добавьте remote вашего репозитория
git remote set-url origin https://github.com/ВАШ_АККАУНТ/vetbot.git
git push -u origin main
🎯 Шаг 2: Создать проект на Railway
Зарегистрируйтесь на railway.app через GitHub

Нажмите "New Project" → "Deploy from GitHub repo"

Авторизуйте доступ к GitHub если потребуется

Выберите ваш форкнутый репозиторий VetBot

Railway автоматически начнет деплой

🎯 Шаг 3: Создать базу данных PostgreSQL
В панели Railway нажмите "New" → "Database"

Выберите PostgreSQL

Дождитесь создания базы (1-2 минуты)

Перейдите в созданную БД → вкладка "Connect"

Скопируйте "Postgres Connection URL" - он понадобится позже

🎯 Шаг 4: Настройка переменных окружения
В панели Railway перейдите в раздел Variables и добавьте:

Переменная	Значение	Описание
TELEGRAM_TOKEN	ваш_токен_от_BotFather	Токен бота Telegram
DATABASE_URL	postgresql://...	Connection URL из шага 3
ADMIN_IDS	ваш_telegram_id	Ваш ID для админ-панели
DEBUG	false	Режим отладки
🔧 Как получить данные:
Telegram Token:

Напишите @BotFather в Telegram

Команда: /newbot

Следуйте инструкциям и получите токен

Telegram ID:

Напишите @userinfobot в Telegram

Он покажет ваш числовой ID

🎯 Шаг 5: Применить миграции базы данных
База данных создана, но таблицы нужно создать вручную:

Способ A: Через Railway CLI (рекомендуется)
bash
# Установите Railway CLI
npm install -g @railway/cli

# Авторизуйтесь
railway login

# Подключитесь к проекту
railway link

# Примените миграции
railway run psql $DATABASE_URL -f migrations/001_init.sql
Способ B: Через панель Railway Query
Перейдите в вашу базу данных в Railway

Откройте вкладку "Query"

Скопируйте содержимое файла migrations/001_init.sql

Вставьте SQL код и выполните его

🎯 Шаг 6: Проверка работоспособности
Проверка логов:
В панели Railway откройте вкладку "Logs"

Должны увидеть:

text
Authorized on account YourBotName
Database connected successfully# 🐾 VetBot - Ветеринарный Telegram бот

**Технологический стек:** Go, PostgreSQL, Docker, Railway
**Сложность деплоя:** низкая (15-20 минут)

## 🚀 Быстрый деплой на Railway

Railway — это PaaS-платформа (как Heroku), которая автоматизирует деплой контейнеров и управление инфраструктурой.

### 📋 Предварительные требования
- **Telegram Bot Token** от [@BotFather](https://t.me/BotFather)
- **Ваш Telegram ID** (можно получить через [@userinfobot](https://t.me/userinfobot))
- GitHub аккаунт (для хранения кода)

---

## 🎯 Пошаговая настройка

### 1. Подготовка репозитория
```bash
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
TELEGRAM_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
DATABASE_URL=postgresql://user:pass@host:port/dbname
ADMIN_IDS=123456789
DEBUG=false
Структура БД
Миграции автоматически создают:

users - пользователи бота

clinics - ветеринарные клиники

veterinarians - врачи

specializations - специализации

schedules - расписания

🐳 Локальная разработка
bash
# Клонирование и настройка
git clone https://github.com/ВАШ_АККАУНТ/vetbot.git
cd vetbot

# Локальная БД (требуется Docker)
docker-compose up -d postgres

# Настройка окружения
cp .env.example .env
# Отредактируйте .env с вашими значениями

# Запуск приложения
go run ./cmd/vetbot
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
Database columns check completed successfully
Bot started. Press Ctrl+C to stop.
Тестирование бота:
Найдите вашего бота в Telegram по username

Отправьте команду /start

Проверьте работу всех функций:

Список специализаций

Поиск по расписанию

Список клиник

🐳 Локальный запуск (опционально)
Если хотите протестировать бота локально:

bash
# Клонируйте ваш репозиторий
git clone https://github.com/ВАШ_АККАУНТ/vetbot.git
cd vetbot

# Настройте окружение
cp .env.example .env
# Отредактируйте .env файл

# Запустите через Docker
docker-compose up -d

# Проверьте логи
docker-compose logs -f vetbot
❌ Решение проблем
Бот не запускается:
Проверьте правильность TELEGRAM_TOKEN

Убедитесь что DATABASE_URL указан верно

Проверьте логи в Railway → Logs

Ошибки с базой данных:
Убедитесь что применили миграции (Шаг 5)

Проверьте что DATABASE_URL скопирован полностью

Бот не отвечает в Telegram:
Проверьте что бот активирован через @BotFather

Убедитесь что в логах нет ошибок авторизации

📞 Поддержка
Если возникли проблемы:

Проверьте логи в Railway

Убедитесь что все шаги выполнены правильно

Обратитесь к разработчику с полным логом ошибок

🔄 Обновление бота
При появлении новых версий:

Синхронизируйте ваш форк с оригинальным репозиторием

Railway автоматически перезапустит бота

При необходимости примените новые миграции

⏱️ Время настройки: 15-30 минут
💸 Стоимость: полностью бесплатно на Railway
🚀 Результат: ваш бот работает 24/7 в облаке!