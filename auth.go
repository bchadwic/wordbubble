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
	ValidateToken(logger Logger, tokenStr string) error
}

type auth struct {
	signingKey string
}

type tokenClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Email    string `json:"email"`
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
		user.Username,
		user.Email,
	}) // constructing payload of the jwt token before signing

	logger.Info("users.GenerateToken: successfully generated token for %s", user.Username)
	return token.SignedString([]byte(auth.signingKey))
}

func (auth *auth) ValidateToken(logger Logger, tokenStr string) error {
	tokenClaims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(auth.signingKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return fmt.Errorf("token's signature was found to be invalid")
		}
		return fmt.Errorf("could not parse the token sent to authorize")
	}
	if !token.Valid {
		return fmt.Errorf("token is expired")
	}
	return nil
}
