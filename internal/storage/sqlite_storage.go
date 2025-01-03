package storage

import (
	"database/sql"
	"fmt"
	"log"

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

	createTable := `CREATE TABLE IF NOT EXISTS questions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        text TEXT,
        answered BOOLEAN DEFAULT FALSE,
        answer TEXT
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) SaveQuestion(question *models.Question) error {
	stmt, err := s.db.Prepare("INSERT INTO questions(user_id, text) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(question.UserID, question.Text)
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
	row := s.db.QueryRow("SELECT id, user_id, text, answered, answer FROM questions WHERE id = ?", id)
	question := &models.Question{}
	var answeredInt int
	if err := row.Scan(&question.ID, &question.UserID, &question.Text, &answeredInt, &question.Answer); err != nil {
		return nil, err
	}
	question.Answered = answeredInt != 0
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

	_, err = stmt.Exec(question.Answered, question.Answer, question.ID)
	return err
}
