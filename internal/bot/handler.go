// internal/bot/handler.go
package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-anonymous-bot/internal/models"
	"telegram-anonymous-bot/pkg/logger"
)

// handleMessage — обрабатывает текстовые/медийные сообщения
func (t *TelegramBot) handleMessage(m *tgbotapi.Message) {
	q := &models.Question{
		UserID:   int(m.From.ID), // Приводим int64 к int
		Username: m.From.UserName,
		Text:     m.Text,
	}

	// Если это фото
	if len(m.Photo) > 0 { // Photo — это []tgbotapi.PhotoSize
		lastIndex := len(m.Photo) - 1
		q.FileID = m.Photo[lastIndex].FileID
		q.MediaType = "photo"

		if m.Caption != "" {
			q.Text = m.Caption
		}
	}

	// Если это видео
	if m.Video != nil {
		q.FileID = m.Video.FileID
		q.MediaType = "video"

		if m.Caption != "" {
			q.Text = m.Caption
		}
	}

	if err := t.storage.SaveQuestion(q); err != nil {
		logger.ErrorLogger.Println("Error saving question:", err)
		t.sendErrorToUser(m.Chat.ID, "Произошла ошибка при сохранении вашего вопроса.")
		return
	}

	lastID, err := t.storage.GetLastQuestionID()
	if err != nil {
		logger.ErrorLogger.Println("Error getting last question ID:", err)
		t.sendErrorToUser(m.Chat.ID, "Произошла ошибка при обработке вашего вопроса.")
		return
	}

	// Отправляем информацию администратору
	adminMsg := fmt.Sprintf("Новый анонимный вопрос:\n%s\nID вопроса: %d\n(from: @%s / %d)",
		q.Text, lastID, q.Username, q.UserID,
	)
	if err := t.sendMessage(int64(t.config.AdminID), adminMsg); err != nil {
		logger.ErrorLogger.Println("Error sending message to admin:", err)
	}

	// Подтверждаем пользователю
	if err := t.sendMessage(m.Chat.ID, "Ваш вопрос отправлен. Спасибо!"); err != nil {
		logger.ErrorLogger.Println("Error sending confirmation:", err)
	}
}

// handleCommand — обрабатывает команды (/start, /answer, /list, /media и т.д.)
func (t *TelegramBot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		t.handleStartCommand(message)
	case "answer":
		t.handleAnswerCommand(message)
	case "list":
		t.handleListCommand(message)
	case "media":
		t.handleMediaCommand(message)
	case "askcohere":
		t.handleAskCohereCommand(message)
	default:
		_ = t.sendMessage(message.Chat.ID, "Неизвестная команда.")
	}
}

func (t *TelegramBot) handleListCommand(message *tgbotapi.Message) {
	// Проверяем, является ли отправитель администратором
	if int(message.From.ID) != t.config.AdminID {
		t.sendErrorToUser(message.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	// Получаем список всех вопросов из хранилища
	questions, err := t.storage.GetAllQuestions()
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Ошибка при получении списка вопросов: "+err.Error())
		return
	}

	// Если вопросов нет, уведомляем администратора
	if len(questions) == 0 {
		t.sendMessage(message.Chat.ID, "Вопросы отсутствуют.")
		return
	}

	// Формируем список вопросов
	var builder strings.Builder
	for _, q := range questions {
		builder.WriteString(fmt.Sprintf(
			"ID: %d, Пользователь: %s, Ответ: %s, Ответил: %t\n",
			q.ID, q.Username, q.Answer, q.Answered,
		))
	}

	// Отправляем список вопросов администратору
	if err := t.sendMessage(message.Chat.ID, "Список вопросов:\n"+builder.String()); err != nil {
		logger.ErrorLogger.Println("Error sending question list:", err)
	}
}

// handleStartCommand — команда /start
func (t *TelegramBot) handleStartCommand(message *tgbotapi.Message) {
	welcomeMsg := "Привет! Задайте свой вопрос, и он будет отправлен анонимно."
	if err := t.sendMessage(message.Chat.ID, welcomeMsg); err != nil {
		logger.ErrorLogger.Println("Error sending start message:", err)
	}
}

// handleAnswerCommand — команда /answer <id> <ответ>
func (t *TelegramBot) handleAnswerCommand(message *tgbotapi.Message) {
	if int(message.From.ID) != t.config.AdminID {
		t.sendErrorToUser(message.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	args := strings.SplitN(message.Text, " ", 3)
	if len(args) < 3 {
		t.sendErrorToUser(message.Chat.ID, "Использование: /answer <id_вопроса> <ответ>")
		return
	}

	questionID, err := strconv.Atoi(args[1])
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Неверный ID вопроса.")
		return
	}
	answerText := args[2]

	q, err := t.storage.GetQuestion(questionID)
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Вопрос не найден или произошла ошибка: "+err.Error())
		return
	}
	if q.Answered {
		t.sendErrorToUser(message.Chat.ID, "На этот вопрос уже был дан ответ.")
		return
	}

	respText := fmt.Sprintf("Ответ на ваш вопрос:\n%s", answerText)
	if err := t.sendMessage(int64(q.UserID), respText); err != nil {
		logger.ErrorLogger.Println("Error sending answer to user:", err)
	}

	q.Answered = true
	q.Answer = answerText
	if err := t.storage.UpdateQuestion(q); err != nil {
		logger.ErrorLogger.Println("Error updating question:", err)
		t.sendErrorToUser(message.Chat.ID, "Произошла ошибка при обновлении вопроса.")
		return
	}

	_ = t.sendMessage(message.Chat.ID, "Ответ отправлен.")
}

// handleMediaCommand — команда /media <id>
func (t *TelegramBot) handleMediaCommand(message *tgbotapi.Message) {
	if int(message.From.ID) != t.config.AdminID {
		t.sendErrorToUser(message.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	args := strings.SplitN(message.Text, " ", 2)
	if len(args) < 2 {
		t.sendErrorToUser(message.Chat.ID, "Использование: /media <id_вопроса>")
		return
	}

	questionID, err := strconv.Atoi(args[1])
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Неверный ID вопроса.")
		return
	}

	q, err := t.storage.GetQuestion(questionID)
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Вопрос не найден или произошла ошибка: "+err.Error())
		return
	}

	if q.FileID == "" {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("У вопроса #%d нет медиафайлов.", q.ID))
		return
	}

	switch q.MediaType {
	case "photo":
		photoMsg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileID(q.FileID))
		photoMsg.Caption = fmt.Sprintf("Вопрос #%d: %s", q.ID, q.Text)
		if _, err := t.bot.Send(photoMsg); err != nil {
			logger.ErrorLogger.Println("Ошибка при отправке фото:", err)
			t.sendErrorToUser(message.Chat.ID, "Не удалось отправить фото.")
			return
		}

	case "video":
		videoMsg := tgbotapi.NewVideo(message.Chat.ID, tgbotapi.FileID(q.FileID))
		videoMsg.Caption = fmt.Sprintf("Вопрос #%d: %s", q.ID, q.Text)
		if _, err := t.bot.Send(videoMsg); err != nil {
			logger.ErrorLogger.Println("Ошибка при отправке видео:", err)
			t.sendErrorToUser(message.Chat.ID, "Не удалось отправить видео.")
			return
		}

	default:
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Неизвестный тип медиа: %s", q.MediaType))
	}
}

// sendMessage — отправка текстового сообщения
func (t *TelegramBot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	return err
}

// sendErrorToUser — отправка сообщения об ошибке
func (t *TelegramBot) sendErrorToUser(chatID int64, text string) {
	if err := t.sendMessage(chatID, text); err != nil {
		logger.ErrorLogger.Printf("Failed to send error message: %v", err)
	}
}

func (t *TelegramBot) handleAskCohereCommand(message *tgbotapi.Message) {
	userInput := strings.TrimSpace(strings.TrimPrefix(message.Text, "/askcohere"))

	if userInput == "" {
		t.sendMessage(message.Chat.ID, "Введите вопрос после команды /askcohere.")
		return
	}

	// Запрос к Cohere API через прокси
	response, err := t.queryCohereWithProxy(userInput)
	if err != nil {
		t.sendMessage(message.Chat.ID, "Ошибка при обращении к Cohere API: "+err.Error())
		return
	}

	// Отправляем ответ пользователю
	t.sendMessage(message.Chat.ID, response)
}
