package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
)

type StartHandler struct {
	Core *core.BotCore
}

func (h *StartHandler) CanHandle(cmd string) bool {
	return cmd == "start"
}

func (h *StartHandler) Handle(msg *tgbotapi.Message) {
	h.Core.SendMessage(msg.Chat.ID, "Привет! Я анонимный бот. Введите /help для списка команд.")
}
