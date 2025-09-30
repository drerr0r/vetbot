FROM golang:1.23-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata postgresql-client

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vetbot ./cmd/vetbot

# Финальный образ
FROM alpine:latest

# Устанавливаем зависимости для runtime
RUN apk --no-cache add ca-certificates tzdata postgresql-client bash

# Создаем пользователя app
RUN addgroup -S app && adduser -S app -G app

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/vetbot .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts ./scripts

# Даем права на выполнение скриптов ДО смены пользователя
RUN chmod +x scripts/apply_migrations_railway.sh

# Создаем пустую папку static если нужно
RUN mkdir -p static

# Устанавливаем права
RUN chown -R app:app /app

# Переключаемся на пользователя app
USER app

# Команда запуска с применением миграций
CMD ./scripts/apply_migrations_railway.sh && ./vetbot