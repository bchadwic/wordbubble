package model

type WordBubble struct {
	Text string `json:"text"`
}

type User struct {
	Id       int64
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
