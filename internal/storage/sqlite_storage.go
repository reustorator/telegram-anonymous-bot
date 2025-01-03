package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"telegram-anonymous-bot/internal/models"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(databaseURL string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return nil, err
	}

	createTable := `
        CREATE TABLE IF NOT EXISTS questions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
			username STRING NOT NULL,
            text TEXT NOT NULL,
            answered INTEGER DEFAULT 0, -- 0 = false, 1 = true
            answer TEXT
        );
    `
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) SaveQuestion(question *models.Question) error {
	stmt, err := s.db.Prepare("INSERT INTO questions (user_id, username, text) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(question.UserID, question.Username, question.Text)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	question.ID = int(id)
	return nil
}

func (s *SQLiteStorage) GetQuestion(id int) (*models.Question, error) {
	row := s.db.QueryRow("SELECT id, user_id, username, text, answered, answer FROM questions WHERE id = ?", id)
	question := &models.Question{}
	var answeredInt int
	var answer sql.NullString

	if err := row.Scan(&question.ID, &question.UserID, &question.Username, &question.Text, &answeredInt, &answer); err != nil {
		return nil, err
	}
	// Преобразуем 0/1 в bool
	question.Answered = answeredInt != 0

	// Обработка NULL для answer
	if answer.Valid {
		question.Answer = answer.String
	} else {
		question.Answer = ""
	}

	return question, nil
}

func (s *SQLiteStorage) GetLastQuestionID() (int, error) {
	row := s.db.QueryRow("SELECT MAX(id) FROM questions")
	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	if id.Valid {
		return int(id.Int64), nil
	}
	return 0, fmt.Errorf("no questions found")
}

func (s *SQLiteStorage) UpdateQuestion(question *models.Question) error {
	stmt, err := s.db.Prepare("UPDATE questions SET answered = ?, answer = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Сохраняем 0 или 1 для поля answered
	answeredVal := 0
	if question.Answered {
		answeredVal = 1
	}

	_, err = stmt.Exec(answeredVal, question.Answer, question.ID)
	return err
}

func (s *SQLiteStorage) GetAllQuestions() ([]*models.Question, error) {
	rows, err := s.db.Query("SELECT id, user_id, username, text, answered, answer FROM questions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question := &models.Question{}
		var answeredInt int
		var answer sql.NullString

		if err := rows.Scan(&question.ID, &question.UserID, &question.Username, &question.Text, &answeredInt, &answer); err != nil {
			return nil, err
		}

		// Преобразуем 0/1 в bool
		question.Answered = answeredInt != 0

		// Обработка NULL для answer
		if answer.Valid {
			question.Answer = answer.String
		} else {
			question.Answer = ""
		}

		questions = append(questions, question)
	}

	// Проверяем ошибки, возникшие во время итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}
