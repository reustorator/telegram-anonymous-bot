package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
)

type MediaHandler struct {
	Core *core.BotCore
}

func (h *MediaHandler) CanHandle(cmd string) bool {
	return cmd == "media"
}

func (h *MediaHandler) Handle(msg *tgbotapi.Message) {
	if int(msg.From.ID) != h.Core.Config.AdminID {
		h.Core.SendMessage(msg.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	args := strings.SplitN(msg.Text, " ", 2)
	if len(args) < 2 {
		h.Core.SendMessage(msg.Chat.ID, "Использование: /media <id>")
		return
	}
	qID, err := strconv.Atoi(args[1])
	if err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Неверный ID вопроса.")
		return
	}

	q, err := h.Core.Storage.GetQuestion(qID)
	if err != nil {
		h.Core.SendMessage(msg.Chat.ID, "Вопрос не найден: "+err.Error())
		return
	}
	if q.FileID == "" {
		h.Core.SendMessage(msg.Chat.ID, fmt.Sprintf("У вопроса #%d нет медиафайла.", q.ID))
		return
	}

	switch q.MediaType {
	case "photo":
		photoMsg := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileID(q.FileID))
		photoMsg.Caption = fmt.Sprintf("Вопрос #%d: %s", q.ID, q.Text)
		h.Core.BotAPI.Send(photoMsg)
	case "video":
		videoMsg := tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FileID(q.FileID))
		videoMsg.Caption = fmt.Sprintf("Вопрос #%d: %s", q.ID, q.Text)
		h.Core.BotAPI.Send(videoMsg)
	default:
		h.Core.SendMessage(msg.Chat.ID, "Неизвестный тип медиа: "+q.MediaType)
	}
}
