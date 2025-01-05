package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type CommandHandler interface {
	CanHandle(cmd string) bool
	Handle(msg *tgbotapi.Message)
}
