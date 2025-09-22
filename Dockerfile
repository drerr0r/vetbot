# Используем официальный образ Go
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o vetbot ./cmd/vetbot

# Финальный образ
FROM alpine:3.18

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/vetbot .
COPY --from=builder /app/migrations ./migrations

# Создаем пользователя для безопасности
RUN adduser -D -g '' vetuser && \
    chown -R vetuser:vetuser /app

USER vetuser

# Экспортируем порт (если понадобится для health checks)
EXPOSE 8080

# Команда запуска
CMD ["./vetbot"]