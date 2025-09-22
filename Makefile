.PHONY: install build test run clean migrate docker-up docker-down docker-logs

# Установка зависимостей
install:
	@echo "📦 Установка зависимостей..."
	go mod download

# Сборка приложения
build:
	@echo "🔨 Сборка приложения..."
	go build -o vetbot ./cmd/vetbot

# Запуск тестов
test:
	@echo "🧪 Запуск тестов..."
	go test -v ./...

# Запуск приложения
run: build
	@echo "🚀 Запуск приложения..."
	./vetbot

# Очистка
clean:
	@echo "🧹 Очистка..."
	rm -f vetbot
	go clean

# Запуск миграций
migrate:
	@echo "🔄 Запуск миграций..."
	go run cmd/vetbot/main.go -migrate

# Запуск Docker Compose
docker-up:
	@echo "🐳 Запуск Docker Compose..."
	docker-compose up -d

# Остановка Docker Compose
docker-down:
	@echo "🛑 Остановка Docker Compose..."
	docker-compose down

# Просмотр логов Docker
docker-logs:
	@echo "📊 Просмотр логов..."
	docker-compose logs -f vetbot

# Help
help:
	@echo "Доступные команды:"
	@echo "  make install    - Установить зависимости"
	@echo "  make build      - Собрать приложение"
	@echo "  make test       - Запустить тесты"
	@echo "  make run        - Запустить приложение"
	@echo "  make clean      - Очистить проект"
	@echo "  make migrate    - Запустить миграции"
	@echo "  make docker-up  - Запустить Docker"
	@echo "  make docker-down - Остановить Docker"
	@echo "  make docker-logs - Просмотр логов Docker"