package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
)

type AnswerHandler struct {
	Core *core.BotCore
}

func (h *AnswerHandler) CanHandle(cmd string) bool {
	return cmd == "answer"
}

func (h *AnswerHandler) Handle(msg *tgbotapi.Message) {
	// Проверка прав (админ)
	if int(msg.From.ID) != h.Core.Config.AdminID {
		h.Core.SendMessage(msg.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	args := strings.SplitN(msg.Text, " ", 3)
	if len(args) < 3 {
		h.Core.SendMessage(msg.Chat.ID, "Использование: /answer <id> <ответ>")
		return
	}

	qID, err := strconv.Atoi(args[1])
	if err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Неверный ID вопроса.")
		return
	}
	answerText := args[2]

	q, err := h.Core.Storage.GetQuestion(qID)
	if err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Вопрос не найден: "+err.Error())
		return
	}
	if q.Answered {
		h.Core.SendMessage(msg.Chat.ID, "На этот вопрос уже был дан ответ.")
		return
	}

	resp := fmt.Sprintf("Ответ на ваш вопрос (ID=%d):\n%s", qID, answerText)
	h.Core.SendMessage(int64(q.UserID), resp)

	q.Answered = true
	q.Answer = answerText
	if err := h.Core.Storage.UpdateQuestion(q); err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Ошибка при обновлении вопроса: "+err.Error())
		return
	}

	h.Core.SendMessage(msg.Chat.ID, fmt.Sprintf("Ответ для вопроса %d отправлен.", qID))
}
