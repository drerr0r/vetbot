🐾 VetBot - Ветеринарный Telegram бот
VetBot помогает пользователям находить ветеринарных врачей по специализации и расписанию.

✨ Возможности
🔍 Поиск врачей по специализации
🕐 Поиск по расписанию и дням недели
📞 Контакты врачей и клиник
👥 Административная панель
📊 Статистика использования

## 🚀 Быстрый старт (для пользователей)

### Предварительные требования
- **Docker** и **Docker Compose** ([скачать здесь](https://www.docker.com/products/docker-desktop/))
- **Telegram Bot Token** (получить у [@BotFather](https://t.me/BotFather))
- **Ваш Telegram ID** (получить у [@userinfobot](https://t.me/userinfobot))

### 📥 Установка и запуск

#### 1. Получите файлы проекта
```bash
# Если проект в Git репозитории
git clone https://github.com/drerr0r/vetbot.git
cd vetbot

# Или распакуйте архив с проектом и перейдите в папку
2. Настройка окружения
bash
# Скопируйте шаблон настроек
cp .env.example .env
Отредактируйте файл .env:

env
# Токен от @BotFather (ОБЯЗАТЕЛЬНО замените!)
TELEGRAM_TOKEN=ваш_настоящий_токен_от_botfather

# Настройки базы данных (можно оставить как есть)
DATABASE_URL=postgres://vetbot_user:vetbot_password@postgres:5432/vetbot?sslmode=disable

# Режим отладки
DEBUG=false

# Ваш Telegram ID (можно несколько через запятую)
ADMIN_IDS=ваш_telegram_id_от_userinfobot
3. Запуск бота
bash
# Автоматический запуск (рекомендуется)
chmod +x scripts/deploy.sh
./scripts/deploy.sh

# Или ручной запуск
docker-compose up -d
4. Проверка работоспособности
bash
# Проверьте статус сервисов
docker-compose ps

# Посмотрите логи
docker-compose logs -f vetbot
В логах должно появиться:

text
Authorized on account YourBotName
Database connected successfully  
Bot started. Press Ctrl+C to stop.
5. Тестирование
Найдите вашего бота в Telegram

Отправьте команду /start

Проверьте ответ бота

🐳 Управление Docker контейнерами
Основные команды
bash
# Остановить бота
docker-compose down

# Перезапустить бота
docker-compose restart

# Обновить бота (после обновления файлов)
docker-compose up -d --build

# Просмотр логов
docker-compose logs -f vetbot

# Проверить статус
docker-compose ps
Резервное копирование данных
bash
# Создать резервную копию базы данных
docker-compose exec postgres pg_dump -U vetbot_user vetbot > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановить из резервной копии
cat backup_file.sql | docker-compose exec -T postgres psql -U vetbot_user vetbot
🛠️ Администрирование
Доступ к базе данных
bash
# Подключиться к базе данных
docker-compose exec postgres psql -U vetbot_user -d vetbot
Мониторинг
bash
# Использование ресурсов
docker stats

# Логи всех сервисов
docker-compose logs
❌ Решение проблем
Бот не запускается
Проверьте логи: docker-compose logs vetbot

Убедитесь, что токен бота корректен

Проверьте, что Docker запущен

Проблемы с базой данных
bash
# Пересоздать базу данных (ВНИМАНИЕ: удалит все данные!)
docker-compose down
docker volume rm vetbot-project_postgres_data
docker-compose up -d
Порт занят
Если порт 5432 занят, измените в docker-compose.yml:

yaml
ports:
  - "5433:5432"  # вместо "5432:5432"
📁 Структура проекта
text
vetbot/
├── .env.example          # Шаблон настроек
├── docker-compose.yml    # Конфигурация Docker
├── Dockerfile           # Сборка приложения
├── migrations/          # Миграции базы данных
├── scripts/            # Скрипты для деплоя
└── ...
🤖 Команды бота
/start - Начало работы

/specializations - Список специализаций

/search - Поиск врача

/clinics - Список клиник

/help - Помощь

/admin - Админ-панель (только для админов)

📞 Поддержка
Если у вас возникли пробле