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
ENV TELEGRAM_BOT_TOKEN=7996743445:AAF83G86x-J1CSJPq6uha-d_Q1xp1Y-sXnc
ENV ADMIN_ID=1093914606
ENV DATABASE_URL=./questions.db

# Запускаем приложение
CMD ["./telegram-anonymous-bot"]
