// internal/bot/bot.go
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/config"
	"telegram-anonymous-bot/internal/storage"
	"telegram-anonymous-bot/pkg/logger"
)

// TelegramBot содержит основные объекты для работы бота.
type TelegramBot struct {
	bot     *tgbotapi.BotAPI
	config  *config.Config
	storage storage.Storage
}

// NewTelegramBot создаёт нового бота с заданной конфигурацией и хранилищем.
func NewTelegramBot(cfg *config.Config, store storage.Storage) (*TelegramBot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = false
	logger.InfoLogger.Printf("Авторизован бот: %s", botAPI.Self.UserName)

	return &TelegramBot{
		bot:     botAPI,
		config:  cfg,
		storage: store,
	}, nil
}

// setBotCommands устанавливает для бота основные команды.
func (t *TelegramBot) setBotCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начало работы с ботом"},
		{Command: "list", Description: "Список вопросов"},
		{Command: "help", Description: "Получить помощь"},
		{Command: "askcohere", Description: "Задать вопрос через Cohere AI"},
	}

	cmdCfg := tgbotapi.NewSetMyCommands(commands...)
	if _, err := t.bot.Request(cmdCfg); err != nil {
		logger.ErrorLogger.Printf("Ошибка установки команд: %v", err)
	} else {
		logger.InfoLogger.Println("Команды успешно установлены.")
	}
}

// Start запускает бесконечный цикл чтения обновлений и их обработки.
func (t *TelegramBot) Start() {
	t.setBotCommands()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)
	for update := range updates {
		switch {
		case update.Message != nil:
			if update.Message.IsCommand() {
				t.handleCommand(update.Message)
			} else {
				t.handleMessage(update.Message)
			}
		case update.CallbackQuery != nil:
			t.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

// sendInlineMenu отправляет меню с кнопками (InlineKeyboard).
func (t *TelegramBot) sendInlineMenu(chatID int64) {
	menu := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Список вопросов", "list"),
			tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")
	msg.ReplyMarkup = menu
	_, _ = t.bot.Send(msg)
}

// handleCallbackQuery обрабатывает нажатие инлайн-кнопок.
func (t *TelegramBot) handleCallbackQuery(cb *tgbotapi.CallbackQuery) {
	switch cb.Data {
	case "list":
		t.sendMessage(cb.Message.Chat.ID, "Вы запросили список вопросов.")
	case "help":
		t.sendMessage(cb.Message.Chat.ID, "Справочная информация.")
	default:
		t.sendMessage(cb.Message.Chat.ID, "Неизвестная команда.")
	}
}

type CohereRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type CohereResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"` // Основной текстовый ответ
	Meta struct {
		APIVersion struct {
			Version      string `json:"version"`
			IsDeprecated bool   `json:"is_deprecated"`
		} `json:"api_version"`
		Warnings    []string `json:"warnings"`
		BilledUnits struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"billed_units"`
	} `json:"meta"`
	FinishReason string `json:"finish_reason"`
}

// queryCohereWithProxy выполняет запрос к Cohere API через (возможно) настроенный прокси.
func (t *TelegramBot) queryCohereWithProxy(prompt string) (string, error) {
	const apiURL = "https://api.cohere.ai/generate"

	payload := CohereRequest{
		Model:       "command-xlarge-nightly",
		Prompt:      prompt,
		MaxTokens:   512,
		Temperature: 0.7,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка формирования JSON: %w", err)
	}

	client := getHttpClientWithProxy(t.config.ProxyURL)

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.config.CohereKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса к Cohere API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("ошибка API Cohere: %s", string(body))
	}

	var cohereResp CohereResponse
	if err := json.NewDecoder(resp.Body).Decode(&cohereResp); err != nil {
		return "", fmt.Errorf("ошибка обработки JSON-ответа: %w", err)
	}

	if cohereResp.Text == "" {
		return "", fmt.Errorf("пустой ответ от Cohere API")
	}
	return cohereResp.Text, nil
}

// getHttpClientWithProxy возвращает кастомный *http.Client, учитывая прокси.
func getHttpClientWithProxy(proxyURL string) *http.Client {
	if proxyURL == "" {
		log.Println("Прокси не указан, используем стандартный клиент")
		return http.DefaultClient
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Fatalf("Некорректный адрес прокси: %v", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		},
	}
}
