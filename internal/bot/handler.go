package bot

import (
	"strconv"
	"strings"

	"telegram-anonymous-bot/internal/models"
	"telegram-anonymous-bot/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (t *TelegramBot) handleMessage(message *tgbotapi.Message) {
	question := message.Text
	userID := message.From.ID

	q := &models.Question{
		UserID: userID,
		Text:   question,
	}

	if err := t.storage.SaveQuestion(q); err != nil {
		logger.ErrorLogger.Println("Error saving question:", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при сохранении вашего вопроса.")
		t.bot.Send(msg)
		return
	}

	lastID, err := t.storage.GetLastQuestionID()
	if err != nil {
		logger.ErrorLogger.Println("Error getting last question ID:", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при обработке вашего вопроса.")
		t.bot.Send(msg)
		return
	}

	// Отправляем вопрос администратору
	adminMsg := tgbotapi.NewMessage(int64(t.config.AdminID),
		"Новый анонимный вопрос:\n"+question+"\nID вопроса: "+strconv.Itoa(lastID))
	t.bot.Send(adminMsg)

	// Подтверждение пользователю
	userMsg := tgbotapi.NewMessage(message.Chat.ID, "Ваш вопрос отправлен. Спасибо!")
	t.bot.Send(userMsg)
}

func (t *TelegramBot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		t.handleStartCommand(message)
	case "answer":
		t.handleAnswerCommand(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда.")
		t.bot.Send(msg)
	}
}

func (t *TelegramBot) handleStartCommand(message *tgbotapi.Message) {
	welcomeMsg := "Привет! Задайте свой вопрос, и он будет отправлен анонимно."
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeMsg)
	t.bot.Send(msg)
}

func (t *TelegramBot) handleAnswerCommand(message *tgbotapi.Message) {
	if message.From.ID != t.config.AdminID {
		msg := tgbotapi.NewMessage(message.Chat.ID, "У вас нет доступа к этой команде.")
		t.bot.Send(msg)
		return
	}

	// Ожидаемый формат: /answer <id_вопроса> <ответ>
	args := strings.SplitN(message.Text, " ", 3)
	if len(args) < 3 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Использование: /answer <id_вопроса> <ответ>")
		t.bot.Send(msg)
		return
	}

	questionID, err := strconv.Atoi(args[1])
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неверный ID вопроса.")
		t.bot.Send(msg)
		return
	}

	answer := args[2]

	question, err := t.storage.GetQuestion(questionID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вопрос не найден.")
		t.bot.Send(msg)
		return
	}

	if question.Answered {
		msg := tgbotapi.NewMessage(message.Chat.ID, "На этот вопрос уже был дан ответ.")
		t.bot.Send(msg)
		return
	}

	// Отправляем ответ пользователю
	respMsg := tgbotapi.NewMessage(int64(question.UserID), "Ответ на ваш вопрос:\n"+answer)
	t.bot.Send(respMsg)

	// Обновляем статус вопроса
	question.Answered = true
	question.Answer = answer
	if err := t.storage.UpdateQuestion(question); err != nil {
		logger.ErrorLogger.Println("Error updating question:", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при обновлении вопроса.")
		t.bot.Send(msg)
		return
	}

	// Подтверждение администратору
	confirmMsg := tgbotapi.NewMessage(message.Chat.ID, "Ответ отправлен.")
	t.bot.Send(confirmMsg)
}
