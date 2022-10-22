package model

import (
	"github.com/golang-jwt/jwt"
)

type Wordbubble struct {
	Text string `json:"text" example:"Hello world, this is just an example of a wordbubble"`
}

type TokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}

type RefreshToken struct {
	Token string `json:"refresh_token"`
}
