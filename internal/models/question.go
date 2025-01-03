package models

type Question struct {
	ID       int
	UserID   int
	Text     string
	Answered bool
	Answer   string
}
