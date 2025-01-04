package storage_test

import (
	"path/filepath"
	"testing"

	"telegram-anonymous-bot/internal/models"
	"telegram-anonymous-bot/internal/storage"
)

// Создаём временную базу данных для тестов
func createTestDB(t *testing.T) *storage.SQLiteStorage {
	t.Helper()

	// Создадим временный файл для SQLite
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	return store
}

func TestSaveAndGetQuestion(t *testing.T) {
	store := createTestDB(t)

	// Подготовим тестовый вопрос
	q := &models.Question{
		UserID:    123,
		Username:  "test_user",
		Text:      "Hello, world!",
		FileID:    "some-file-id",
		MediaType: "photo",
	}

	// Сохраняем вопрос
	if err := store.SaveQuestion(q); err != nil {
		t.Fatalf("SaveQuestion failed: %v", err)
	}
	if q.ID == 0 {
		t.Fatalf("Expected question ID to be set, got 0")
	}

	// Получаем вопрос из базы
	got, err := store.GetQuestion(q.ID)
	if err != nil {
		t.Fatalf("GetQuestion failed: %v", err)
	}

	if got.UserID != 123 {
		t.Errorf("Expected UserID=123, got %d", got.UserID)
	}
	if got.Username != "test_user" {
		t.Errorf("Expected Username=test_user, got %s", got.Username)
	}
	if got.Text != "Hello, world!" {
		t.Errorf("Expected Text=Hello, world!, got %s", got.Text)
	}
	if got.FileID != "some-file-id" {
		t.Errorf("Expected FileID=some-file-id, got %s", got.FileID)
	}
	if got.MediaType != "photo" {
		t.Errorf("Expected MediaType=photo, got %s", got.MediaType)
	}
}

func TestUpdateQuestion(t *testing.T) {
	store := createTestDB(t)

	q := &models.Question{
		UserID:   999,
		Username: "update_tester",
		Text:     "Update me!",
	}
	if err := store.SaveQuestion(q); err != nil {
		t.Fatalf("SaveQuestion failed: %v", err)
	}

	// Обновляем вопрос (помечаем как отвеченный)
	q.Answered = true
	q.Answer = "Updated answer"
	if err := store.UpdateQuestion(q); err != nil {
		t.Fatalf("UpdateQuestion failed: %v", err)
	}

	got, err := store.GetQuestion(q.ID)
	if err != nil {
		t.Fatalf("GetQuestion after update failed: %v", err)
	}
	if !got.Answered {
		t.Errorf("Expected Answered=true, got false")
	}
	if got.Answer != "Updated answer" {
		t.Errorf("Expected Answer='Updated answer', got %s", got.Answer)
	}
}

func TestGetAllQuestions(t *testing.T) {
	store := createTestDB(t)

	// Сохраняем несколько вопросов
	qs := []*models.Question{
		{UserID: 1, Username: "user1", Text: "Question1"},
		{UserID: 2, Username: "user2", Text: "Question2"},
		{UserID: 3, Username: "user3", Text: "Question3"},
	}
	for _, q := range qs {
		if err := store.SaveQuestion(q); err != nil {
			t.Fatalf("SaveQuestion failed for %v: %v", q, err)
		}
	}

	all, err := store.GetAllQuestions()
	if err != nil {
		t.Fatalf("GetAllQuestions failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 questions, got %d", len(all))
	}
}

func TestGetLastQuestionID(t *testing.T) {
	store := createTestDB(t)

	// Без вопросов
	_, err := store.GetLastQuestionID()
	if err == nil {
		t.Errorf("Expected error when no questions in DB")
	}

	// Добавим вопросы
	q1 := &models.Question{UserID: 10, Username: "u10", Text: "Q1"}
	q2 := &models.Question{UserID: 11, Username: "u11", Text: "Q2"}
	_ = store.SaveQuestion(q1)
	_ = store.SaveQuestion(q2)

	lastID, err := store.GetLastQuestionID()
	if err != nil {
		t.Fatalf("GetLastQuestionID failed: %v", err)
	}
	if lastID != q2.ID {
		t.Errorf("Expected lastID=%d, got %d", q2.ID, lastID)
	}
}
