// cmd/bot/main.go
package main

import (
	"github.com/joho/godotenv"
	"log"
	"telegram-anonymous-bot/internal/bot"
	"telegram-anonymous-bot/internal/config"
	"telegram-anonymous-bot/internal/storage"
	"telegram-anonymous-bot/pkg/logger"
)

func main() {
	logger.Init()
	LoadEnv()

	// Загрузка конфигурации (токен, admin_id, database_url из .env)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Инициализация SQLite хранилища
	store, err := storage.NewSQLiteStorage(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error initializing SQLite storage:", err)
	}

	// Инициализация бота
	telegramBot, err := bot.NewTelegramBot(cfg, store)
	if err != nil {
		log.Fatal("Error initializing Telegram bot:", err)
	}

	// Запуск бота
	telegramBot.Start()
}

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Ошибка загрузки .env файла:", err)
	}
}
