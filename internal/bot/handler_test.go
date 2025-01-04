package bot_test

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"

	"telegram-anonymous-bot/internal/bot" // <-- Пакет, где лежит TelegramBot
	"telegram-anonymous-bot/internal/models"
)

// -------------------- Моки -------------------- //

// MockStorage реализует интерфейс Storage для теста
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveQuestion(q *models.Question) error {
	args := m.Called(q)
	return args.Error(0)
}
func (m *MockStorage) GetQuestion(id int) (*models.Question, error) {
	args := m.Called(id)
	// Если нужно, возвращаем заранее заданный результат
	return args.Get(0).(*models.Question), args.Error(1)
}

// Если в вашем Storage есть и другие методы (UpdateQuestion, GetAllQuestions и т.д.),
// их тоже нужно замокать, иначе во время теста может быть паника.

// -------------------- Тест -------------------- //

func TestHandleMessage(t *testing.T) {
	// 1. Создаём мок хранилища
	storageMock := new(MockStorage)

	// Ожидаем, что при сохранении вопроса всё ок (возврат nil в качестве ошибки)
	storageMock.On("SaveQuestion", mock.Anything).Return(nil)

	// 2. Создаём "бота" — для теста достаточно минимальных полей
	telegramBot := &bot.TelegramBot{
		// bot.BotAPI можно замокать отдельно,
		// но если метод HandleMessage не требует реальных вызовов к Telegram, оставим nil
		Bot:     nil,
		Config:  &bot.Config{TelegramBotToken: "fake_token", AdminID: 999999},
		Storage: storageMock,
	}

	// 3. Подготовим тестовое сообщение (только текст)
	message := &tgbotapi.Message{
		Text: "Hello world!",
		From: &tgbotapi.User{
			ID:       12345,
			UserName: "testuser",
		},
		// Chat.ID — куда бот вернёт ответ (в вашем коде, если отправляете сообщения)
		Chat: &tgbotapi.Chat{ID: 11111},
	}

	// 4. Вызываем тестируемый метод
	telegramBot.HandleMessage(message)

	// 5. Проверяем, что метод SaveQuestion был вызван
	storageMock.AssertCalled(t, "SaveQuestion", mock.Anything)

	// (опционально) Проверяем конкретные поля:
	// storageMock.AssertCalled(t, "SaveQuestion", &models.Question{
	// 	UserID:   12345,
	// 	Username: "testuser",
	// 	Text:     "Hello world!",
	// })

	// Если хотим проверить точное совпадение, используйте:
	// storageMock.AssertCalled(t, "SaveQuestion", mock.MatchedBy(func(q *models.Question) bool {
	// 	return q.UserID == 12345 && q.Username == "testuser" && q.Text == "Hello world!"
	// }))
}
