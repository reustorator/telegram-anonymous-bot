package main

import (
	"log"
	"telegram-anonymous-bot/internal/bot"
	"telegram-anonymous-bot/internal/config"
	"telegram-anonymous-bot/internal/storage"
	"telegram-anonymous-bot/pkg/logger"
)

func main() {
	// Инициализация логгера
	logger.Init()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Инициализация хранилища
	store, err := storage.NewSQLiteStorage(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error initializing storage:", err)
	}

	// Инициализация бота
	telegramBot, err := bot.NewTelegramBot(cfg, store)
	if err != nil {
		log.Fatal("Error initializing Telegram bot:", err)
	}

	// Запуск бота
	telegramBot.Start()
}
