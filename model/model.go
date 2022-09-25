package model

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type WordBubble struct {
	Text string `json:"text"`
}

type User struct {
	Id       int64
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Context struct {
	Time time.Time `json:"time"`
}

type TokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}
