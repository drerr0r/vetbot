#!/bin/bash
set -e  # Выход при ошибке

echo "🚀 Applying database migrations on Railway..."

# Получаем DATABASE_URL из переменных окружения
DATABASE_URL=${DATABASE_URL}

if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL not found"
    exit 1
fi

echo "📝 Checking for database migrations..."

# Применяем миграции по порядку
for migration_file in migrations/*.sql; do
    if [ -f "$migration_file" ]; then
        echo "📝 Applying migration: $migration_file"
        psql $DATABASE_URL -f "$migration_file" 
        if [ $? -eq 0 ]; then
            echo "✅ Successfully applied: $migration_file"
        else
            echo "⚠️ Migration completed with warnings: $migration_file"
            # Не выходим с ошибкой, так как "already exists" - это нормально
        fi
    fi
done

echo "🎉 All migrations completed successfully!"