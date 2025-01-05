package core

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
)

type BotCore struct {
	BotAPI  *tgbotapi.BotAPI
	Config  *config.Config
	Storage storage.Storage
}

// SendMessage отправляет обычное сообщение пользователю.
func (bc *BotCore) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bc.BotAPI.Send(msg); err != nil {
		log.Printf("SendMessage error: %v", err)
	}
}

func (bc *BotCore) sendCommandsKeyboard(chatID int64) {
	// Создаём кнопки
	row := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/answer "),
		tgbotapi.NewKeyboardButton("/askcohere "),
	)
	// Можно добавить больше кнопок, если нужно

	// Создаём клавиатуру
	replyKeyboard := tgbotapi.NewReplyKeyboard(row)
	// Параметры клавиатуры
	replyKeyboard.OneTimeKeyboard = true // клавиатура «схлопнется» после нажатия
	replyKeyboard.ResizeKeyboard = true  // подгоняем размер под кнопки

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(chatID, "Выберите команду:")
	msg.ReplyMarkup = replyKeyboard
	_, _ = bc.BotAPI.Send(msg)
}

// QueryCohereWithProxy пример интеграции с Cohere.
func (bc *BotCore) QueryCohereWithProxy(prompt string) (string, error) {
	const apiURL = "https://api.cohere.ai/generate"

	payload := struct {
		Model       string  `json:"model"`
		Prompt      string  `json:"prompt"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
	}{
		Model:       "command-xlarge-nightly",
		Prompt:      prompt,
		MaxTokens:   512,
		Temperature: 0.7,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка формирования JSON: %w", err)
	}

	client := bc.getHttpClientWithProxy(bc.Config.ProxyURL)

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+bc.Config.CohereKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса к Cohere API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Cohere API error: %s", string(body))
	}

	var cohereResp struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cohereResp); err != nil {
		return "", fmt.Errorf("ошибка обработки JSON-ответа: %w", err)
	}

	if cohereResp.Text == "" {
		return "", fmt.Errorf("пустой ответ от Cohere")
	}

	return cohereResp.Text, nil
}

// getHttpClientWithProxy возвращает http-клиент, учитывая прокси.
func (bc *BotCore) getHttpClientWithProxy(proxyURL string) *http.Client {
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
