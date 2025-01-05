package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
)

// HelpHandler обрабатывает команду /help
type HelpHandler struct {
	Core *core.BotCore
}

// CanHandle проверяет, является ли команда — "help"
func (h *HelpHandler) CanHandle(cmd string) bool {
	return cmd == "help"
}

// Handle реализует логику для /help
func (h *HelpHandler) Handle(msg *tgbotapi.Message) {
	// Можно отправить краткую справку по основным командам
	helpText := `Доступные команды:
    
/start — начало работы
/list  — список всех вопросов (админ)
/answer <id> <ответ> — ответ на вопрос (админ)
/media <id> — показать фото/видео (админ)
/askcohere <текст> — спросить Cohere AI
/help — показать эту справку
`
	h.Core.SendMessage(msg.Chat.ID, helpText)
}
