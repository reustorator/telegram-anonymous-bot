// internal/bot/bot.go
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"net/http"
	"telegram-anonymous-bot/internal/config"
	"telegram-anonymous-bot/internal/storage"
	"telegram-anonymous-bot/pkg/logger"
)

type TelegramBot struct {
	bot     *tgbotapi.BotAPI
	config  *config.Config
	storage storage.Storage
}

// Конструктор
func NewTelegramBot(cfg *config.Config, store storage.Storage) (*TelegramBot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = true
	logger.InfoLogger.Printf("Авторизован аккаунт %s", botAPI.Self.UserName)

	return &TelegramBot{
		bot:     botAPI,
		config:  cfg,
		storage: store,
	}, nil
}

func (t *TelegramBot) setBotCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начало работы с ботом"},
		{Command: "help", Description: "Получить справочную информацию"},
		{Command: "list", Description: "Показать список всех вопросов"},
		{Command: "media", Description: "Показать медиафайл по ID"},
		/*		{Command: "askgpt", Description: "Спросить ChatGPT"},
				todo Бесплатных умных нейросетей мало, поэтому пока что оставлю
		*/
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)
	_, err := t.bot.Request(cfg)
	if err != nil {
		log.Printf("Ошибка при установке команд бота: %v\n", err)
	} else {
		log.Println("Команды бота успешно установлены.")
	}
}

// Запуск
func (t *TelegramBot) Start() {
	// Устанавливаем команды
	t.setBotCommands() // Здесь просто вызываем метод

	// Далее логика запуска обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				t.handleCommand(update.Message)
			} else {
				t.handleMessage(update.Message)
			}
		} else if update.CallbackQuery != nil {
			t.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func (t *TelegramBot) sendInlineMenu(chatID int64) {
	button1 := tgbotapi.NewInlineKeyboardButtonData("Список вопросов", "list")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Помощь", "help")

	inlineKb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button1, button2))

	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")
	msg.ReplyMarkup = inlineKb

	t.bot.Send(msg)
}

func (t *TelegramBot) handleHelpCommand(message *tgbotapi.Message) {
	t.sendInlineMenu(message.Chat.ID)
}

func (t *TelegramBot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	switch callback.Data {
	case "list":
		t.sendMessage(callback.Message.Chat.ID, "Вы запросили список вопросов.")
	case "help":
		t.sendMessage(callback.Message.Chat.ID, "Справочная информация.")
	default:
		t.sendMessage(callback.Message.Chat.ID, "Неизвестная команда.")
	}
}

type HuggingFaceRequest struct {
	Inputs string `json:"inputs"`
}

type HuggingFaceResponse struct {
	GeneratedText string `json:"generated_text"`
}

func (t *TelegramBot) queryHuggingFace(model, userInput string) (string, error) {
	apiURL := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", model) //выносить потом всё в конфиг
	apiKey := t.config.HuggingFaceKey                                              // Убедитесь, что ключ загружен из .env

	// Формируем запрос
	payload := HuggingFaceRequest{Inputs: userInput}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	// Отправляем запрос
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("ошибка при создании HTTP-запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("ошибка API: %s", body)
	}

	// Читаем ответ
	var responses []HuggingFaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return "", fmt.Errorf("ошибка при обработке ответа: %v", err)
	}

	// Проверяем наличие текста
	if len(responses) == 0 || responses[0].GeneratedText == "" {
		return "", fmt.Errorf("пустой ответ от модели")
	}

	return responses[0].GeneratedText, nil
}
