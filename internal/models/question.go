package models

type Question struct {
	ID        int
	UserID    int
	Username  string
	Text      string
	Answered  bool
	Answer    string
	FileID    string
	MediaType string
}
