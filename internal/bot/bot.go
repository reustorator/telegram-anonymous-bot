package bot

import (
	"telegram-anonymous-bot/internal/config"
	"telegram-anonymous-bot/internal/storage"
	"telegram-anonymous-bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBot struct {
	bot     *tgbotapi.BotAPI
	config  *config.Config
	storage storage.Storage
}

func NewTelegramBot(cfg *config.Config, storage storage.Storage) (*TelegramBot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = true
	logger.InfoLogger.Printf("Authorized on account %s", botAPI.Self.UserName)

	return &TelegramBot{
		bot:     botAPI,
		config:  cfg,
		storage: storage,
	}, nil
}

func (t *TelegramBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		logger.ErrorLogger.Panic(err)
	}

	for update := range updates {
		if update.Message == nil { // игнорируем обновления, не содержащие сообщений
			continue
		}

		if update.Message.IsCommand() {
			t.handleCommand(update.Message)
			continue
		}

		t.handleMessage(update.Message)
	}
}
