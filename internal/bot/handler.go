package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"telegram-anonymous-bot/internal/models"
	"telegram-anonymous-bot/pkg/logger"
)

// handleMessage обрабатывает обычное сообщение (вопрос).
func (t *TelegramBot) handleMessage(message *tgbotapi.Message) {
	// Извлекаем текст вопроса, username и userID
	questionText := message.Text
	username := message.From.UserName // Может быть пустым, если пользователь не задал username
	userID := message.From.ID

	// Создаём структуру вопроса
	q := &models.Question{
		UserID:   userID,
		Username: username,
		Text:     questionText,
	}

	if err := t.storage.SaveQuestion(q); err != nil {
		logger.ErrorLogger.Println("Error saving question:", err)
		t.sendErrorToUser(message.Chat.ID, "Произошла ошибка при сохранении вашего вопроса.")
		return
	}

	lastID, err := t.storage.GetLastQuestionID()
	if err != nil {
		logger.ErrorLogger.Println("Error getting last question ID:", err)
		t.sendErrorToUser(message.Chat.ID, "Произошла ошибка при обработке вашего вопроса.")
		return
	}

	// Отправляем вопрос администратору
	adminText := fmt.Sprintf(
		"Новый анонимный вопрос:\n%s\nID вопроса: %d\n(from: @%s / %d)",
		questionText, lastID, username, userID,
	)
	if err := t.sendMessage(int64(t.config.AdminID), adminText); err != nil {
		logger.ErrorLogger.Println("Error sending message to admin:", err)
	}

	// Подтверждение пользователю
	if err := t.sendMessage(message.Chat.ID, "Ваш вопрос отправлен. Спасибо!"); err != nil {
		logger.ErrorLogger.Println("Error sending confirmation to user:", err)
	}
}

// handleCommand обрабатывает команды: /start, /answer, /list и т.д.
func (t *TelegramBot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		t.handleStartCommand(message)
	case "answer":
		t.handleAnswerCommand(message)
	case "list":
		t.handleListCommand(message)
	default:
		_ = t.sendMessage(message.Chat.ID, "Неизвестная команда.")
	}
}

// handleStartCommand отправляет приветственное сообщение пользователю.
func (t *TelegramBot) handleStartCommand(message *tgbotapi.Message) {
	welcomeMsg := "Привет! Задайте свой вопрос, и он будет отправлен анонимно."
	if err := t.sendMessage(message.Chat.ID, welcomeMsg); err != nil {
		logger.ErrorLogger.Println("Error sending start message:", err)
	}
}

// handleAnswerCommand обрабатывает команду /answer <questionID> <answer> и отвечает на вопрос.
func (t *TelegramBot) handleAnswerCommand(message *tgbotapi.Message) {
	// Проверяем, что это администратор
	if message.From.ID != t.config.AdminID {
		t.sendErrorToUser(message.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	// Парсим аргументы команды
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

	question, err := t.storage.GetQuestion(questionID)
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Вопрос не найден или произошла ошибка: "+err.Error())
		return
	}

	if question.Answered {
		t.sendErrorToUser(message.Chat.ID, "На этот вопрос уже был дан ответ.")
		return
	}

	// Отправляем ответ пользователю
	responseText := fmt.Sprintf("Ответ на ваш вопрос:\n%s", answerText)
	if err := t.sendMessage(int64(question.UserID), responseText); err != nil {
		logger.ErrorLogger.Println("Error sending answer to user:", err)
	}

	// Обновляем статус вопроса
	question.Answered = true
	question.Answer = answerText
	if err := t.storage.UpdateQuestion(question); err != nil {
		logger.ErrorLogger.Println("Error updating question:", err)
		t.sendErrorToUser(message.Chat.ID, "Произошла ошибка при обновлении вопроса.")
		return
	}

	// Подтверждаем администратору
	_ = t.sendMessage(message.Chat.ID, "Ответ отправлен.")
}

// handleListCommand обрабатывает команду /list и выводит список всех вопросов.
func (t *TelegramBot) handleListCommand(message *tgbotapi.Message) {
	if message.From.ID != t.config.AdminID {
		t.sendErrorToUser(message.Chat.ID, "У вас нет доступа к этой команде.")
		return
	}

	questions, err := t.storage.GetAllQuestions()
	if err != nil {
		t.sendErrorToUser(message.Chat.ID, "Ошибка при получении списка вопросов: "+err.Error())
		return
	}
	if len(questions) == 0 {
		t.sendMessage(message.Chat.ID, "Вопросы отсутствуют.")
		return
	}

	var idsList strings.Builder
	for _, q := range questions {
		idsList.WriteString(fmt.Sprintf(
			"ID: %d, Пользователь: %s, Текст: %s, Ответ: %t\n",
			q.ID, q.Username, q.Text, q.Answered,
		))
	}

	if err := t.sendMessage(message.Chat.ID, "Список вопросов:\n"+idsList.String()); err != nil {
		logger.ErrorLogger.Println("Error sending question list:", err)
	}
}

// sendMessage отправляет сообщение пользователю (чату).
func (t *TelegramBot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	return err
}

// sendErrorToUser упрощает отправку сообщения об ошибке пользователю.
func (t *TelegramBot) sendErrorToUser(chatID int64, text string) {
	if err := t.sendMessage(chatID, text); err != nil {
		logger.ErrorLogger.Printf("Failed to send error message to user %d: %v", chatID, err)
	}
}
