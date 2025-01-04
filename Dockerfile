# Используем официальный образ Go для сборки
FROM golang:1.20-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь код в рабочую директорию
COPY . .

# Собираем приложение
RUN go build -o telegram-anonymous-bot ./cmd/bot

# Используем минимальный образ для запуска
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем бинарный файл из стадии сборки
COPY --from=builder /app/telegram-anonymous-bot .

# Копируем файл базы данных (если необходимо)
COPY --from=builder /app/questions.db .

# Копируем файл конфигурации (можно использовать переменные окружения вместо этого)
COPY --from=builder /app/config.env .

# Устанавливаем переменные окружения (можно настроить через docker-compose)
ENV TELEGRAM_BOT_TOKEN=
ENV ADMIN_ID=
ENV DATABASE_URL=

# Запускаем приложение
CMD ["./telegram-anonymous-bot"]
