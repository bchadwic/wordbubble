package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	minPasswordLength = 6
	maxUsernameLength = 40
	maxEmailLength    = 320
)

type Auth interface {
	GenerateToken(logger Logger, user *User) (string, error)
	ValidateTokenAndReceiveId(logger Logger, tokenStr string) (int64, error)
}

type auth struct {
	signingKey string
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
}

func NewAuth(signingKey string) *auth {
	return &auth{signingKey}
}

func (auth *auth) GenerateToken(logger Logger, user *User) (string, error) {
	logger.Info("users.GenerateToken: generating token for %s", user.Username)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.UserId,
		user.Username,
	}) // constructing payload of the jwt token before signing

	logger.Info("users.GenerateToken: successfully generated token for %s", user.Username)
	return token.SignedString([]byte(auth.signingKey))
}

func (auth *auth) ValidateTokenAndReceiveId(logger Logger, tokenStr string) (int64, error) {
	tokenClaims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(auth.signingKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return 0, fmt.Errorf("token's signature was found to be invalid")
		}
		return 0, fmt.Errorf("could not parse the token sent to authorize")
	}
	if !token.Valid {
		return 0, fmt.Errorf("token is expired")
	}
	return tokenClaims.UserId, nil
}
