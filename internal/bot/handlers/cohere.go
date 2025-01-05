package handlers

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
)

// CohereHandler обрабатывает команду /askcohere
type CohereHandler struct {
	Core *core.BotCore
}

func (h *CohereHandler) CanHandle(cmd string) bool {
	return cmd == "askcohere"
}

func (h *CohereHandler) Handle(msg *tgbotapi.Message) {
	userInput := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/askcohere"))
	if userInput == "" {
		h.Core.SendMessage(msg.Chat.ID, "Введите вопрос после /askcohere")
		return
	}

	answer, err := h.Core.QueryCohereWithProxy(userInput)
	if err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Ошибка Cohere API: "+err.Error())
		return
	}

	h.Core.SendMessage(msg.Chat.ID, answer)
}
