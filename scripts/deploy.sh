#!/bin/bash

# Скрипт деплоя VetBot

set -e

echo "🚀 Деплой VetBot"

# Проверка переменных окружения
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ TELEGRAM_BOT_TOKEN не установлен"
    exit 1
fi

# Остановка текущих контейнеров
echo "🛑 Остановка текущих контейнеров"
docker-compose down

# Обновление кода
echo "📥 Обновление кода"
git pull origin main

# Пересборка образов
echo "🔨 Пересборка образов Docker"
docker-compose build

# Запуск контейнеров
echo "🚀 Запуск контейнеров"
docker-compose up -d

echo "✅ Деплой завершен!"
echo "📊 Проверьте логи: docker-compose logs -f vetbot"