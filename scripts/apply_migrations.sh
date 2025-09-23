#!/bin/bash
# scripts/apply_migrations.sh

echo "Applying database migrations..."

# Проверяем существование базы данных и применяем миграции
docker-compose exec db psql -U vetbot -d vetbot -f /migrations/001_init.sql

echo "Migrations applied successfully!"