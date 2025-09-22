#!/bin/bash

# Скрипт установки и настройки VetBot

set -e

echo "🐾 Установка VetBot"

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Установите Go 1.21+"
    exit 1
fi

# Проверка наличия Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен. Установите Docker"
    exit 1
fi

# Создание .env файла если его нет
if [ ! -f .env ]; then
    echo "📝 Создание .env файла из примера"
    cp .env.example .env
    echo "⚠️ Отредактируйте .env файл перед запуском"
fi

# Установка зависимостей
echo "📦 Установка зависимостей Go"
go mod download

# Сборка приложения
echo "🔨 Сборка приложения"
go build -o vetbot ./cmd/vetbot

# Запуск Docker Compose
echo "🐳 Запуск PostgreSQL через Docker"
docker-compose up -d postgres

echo "✅ Установка завершена!"
echo "💡 Отредактируйте .env файл и запустите: ./vetbot"