package storage

import "telegram-anonymous-bot/internal/models"

// Storage интерфейс для операций с данными
type Storage interface {
	SaveQuestion(question *models.Question) error
	GetQuestion(id int) (*models.Question, error)
	GetAllQuestions() ([]*models.Question, error) // Новый метод
	GetLastQuestionID() (int, error)
	UpdateQuestion(question *models.Question) error
}
