package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
)

type ListHandler struct {
	Core *core.BotCore
}

func (h *ListHandler) CanHandle(cmd string) bool {
	return cmd == "list"
}

func (h *ListHandler) Handle(msg *tgbotapi.Message) {
	// Проверка прав (админ)
	if int(msg.From.ID) != h.Core.Config.AdminID {
		h.Core.SendMessage(msg.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	questions, err := h.Core.Storage.GetAllQuestions()
	if err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Ошибка при получении списка вопросов: "+err.Error())
		return
	}
	if len(questions) == 0 {
		h.Core.SendMessage(msg.Chat.ID, "Вопросы отсутствуют.")
		return
	}

	var result string
	for _, q := range questions {
		result += fmt.Sprintf("ID: %d | User: %s | Ответ: %s | Ответил: %t\n",
			q.ID, q.Username, q.Answer, q.Answered)
	}
	h.Core.SendMessage(msg.Chat.ID, result)
}
