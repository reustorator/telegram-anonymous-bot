// internal/storage/sqlite_storage.go

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
    username TEXT NOT NULL,
    text TEXT NOT NULL,
    file_id TEXT,
    media_type TEXT,
    answered INTEGER DEFAULT 0, -- 0 = false, 1 = true
    answer TEXT
);
`
	if _, err = db.Exec(createTable); err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) SaveQuestion(q *models.Question) error {
	query := `
INSERT INTO questions (user_id, username, text, file_id, media_type)
VALUES (?, ?, ?, ?, ?)
`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(q.UserID, q.Username, q.Text, q.FileID, q.MediaType)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	q.ID = int(id)
	return nil
}

func (s *SQLiteStorage) GetQuestion(id int) (*models.Question, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, username, text, file_id, media_type, answered, answer
FROM questions
WHERE id = ?
`, id)

	q := &models.Question{}
	var answeredInt int
	var answer sql.NullString
	var fileID sql.NullString
	var mediaType sql.NullString

	if err := row.Scan(
		&q.ID,
		&q.UserID,
		&q.Username,
		&q.Text,
		&fileID,
		&mediaType,
		&answeredInt,
		&answer,
	); err != nil {
		return nil, err
	}

	if fileID.Valid {
		q.FileID = fileID.String
	}
	if mediaType.Valid {
		q.MediaType = mediaType.String
	}
	q.Answered = answeredInt != 0

	if answer.Valid {
		q.Answer = answer.String
	} else {
		q.Answer = ""
	}

	return q, nil
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

func (s *SQLiteStorage) UpdateQuestion(q *models.Question) error {
	stmt, err := s.db.Prepare(`
UPDATE questions
SET answered = ?, answer = ?
WHERE id = ?
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	answeredVal := 0
	if q.Answered {
		answeredVal = 1
	}

	_, err = stmt.Exec(answeredVal, q.Answer, q.ID)
	return err
}

func (s *SQLiteStorage) GetAllQuestions() ([]*models.Question, error) {
	rows, err := s.db.Query(`
SELECT id, user_id, username, text, file_id, media_type, answered, answer
FROM questions
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q := &models.Question{}
		var answeredInt int
		var answer sql.NullString
		var fileID sql.NullString
		var mediaType sql.NullString

		if err := rows.Scan(
			&q.ID,
			&q.UserID,
			&q.Username,
			&q.Text,
			&fileID,
			&mediaType,
			&answeredInt,
			&answer,
		); err != nil {
			return nil, err
		}

		if fileID.Valid {
			q.FileID = fileID.String
		}
		if mediaType.Valid {
			q.MediaType = mediaType.String
		}

		q.Answered = (answeredInt != 0)

		if answer.Valid {
			q.Answer = answer.String
		} else {
			q.Answer = ""
		}

		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}
