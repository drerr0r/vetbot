#!/bin/bash
echo "🚀 Applying database migrations on Railway..."

# Получаем DATABASE_URL из переменных окружения
DATABASE_URL=${DATABASE_URL}

if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL not found"
    exit 1
fi

# Применяем миграции
echo "📝 Applying migrations from migrations/001_init.sql..."
psql $DATABASE_URL -f migrations/001_init.sql

if [ $? -eq 0 ]; then
    echo "✅ Migrations applied successfully!"
else
    echo "❌ Failed to apply migrations"
    exit 1
fi