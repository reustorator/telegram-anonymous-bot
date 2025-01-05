package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/bot/core"
	"telegram-anonymous-bot/internal/bot/handlers"
	"telegram-anonymous-bot/internal/config"
	"telegram-anonymous-bot/internal/storage"
	"telegram-anonymous-bot/pkg/logger"
)

// TelegramBot — главный объект, регистрирующий хендлеры и обрабатывающий входящие команды.
type TelegramBot struct {
	core     *core.BotCore
	handlers []handlers.CommandHandler
}

func NewTelegramBot(cfg *config.Config, store storage.Storage) (*TelegramBot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = false
	logger.InfoLogger.Printf("Авторизован бот: %s", botAPI.Self.UserName)

	bc := &core.BotCore{
		BotAPI:  botAPI,
		Config:  cfg,
		Storage: store,
	}

	return &TelegramBot{
		core: bc,
		handlers: []handlers.CommandHandler{
			&handlers.StartHandler{Core: bc},
			&handlers.AnswerHandler{Core: bc},
			&handlers.ListHandler{Core: bc},
			&handlers.MediaHandler{Core: bc},
			&handlers.CohereHandler{Core: bc},
			&handlers.HelpHandler{Core: bc},
			// ... при необходимости добавляйте новые
		},
	}, nil
}

func (t *TelegramBot) Start() {
	// Можно установить команды
	// u := tgbotapi.NewUpdate(0)
	// ...

	updates := t.core.BotAPI.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:  0,
		Timeout: 60,
	})

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			t.handleCommand(update.Message)
		}
		// Или обрабатывать CallbackQuery/просто текст
	}
}

// handleCommand ищет подходящий CommandHandler и вызывает его метод Handle(msg).
func (t *TelegramBot) handleCommand(msg *tgbotapi.Message) {
	cmd := msg.Command()
	for _, h := range t.handlers {
		if h.CanHandle(cmd) {
			h.Handle(msg)
			return
		}
	}
	t.core.SendMessage(msg.Chat.ID, "Неизвестная команда (через общий router).")
}
