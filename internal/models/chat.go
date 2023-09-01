package models

type Message struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	Token     string `json:"token"`
	Timestamp string `json:"timestamp"`
}
