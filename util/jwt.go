package util

import (
	"log"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/golang-jwt/jwt"
)

var SigningKey func() []byte = func() []byte {
	log.Fatal("token signing key was not set")
	return nil
}

func GenerateSignedToken(iat, exp, userId int64) string {
	signedToken, _ := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &model.TokenClaims{StandardClaims: jwt.StandardClaims{IssuedAt: iat, ExpiresAt: exp}, UserId: userId},
	).SignedString(SigningKey())
	return signedToken
}

func GetUserIdFromTokenString(tokenStr string) (int64, error) {
	if tokenClaims, err := ParseWithClaims(tokenStr); err != nil {
		return 0, resp.ErrInvalidTokenSignature
	} else {
		return tokenClaims.UserId, nil
	}
}

func ParseWithClaims(tokenStr string) (*model.TokenClaims, error) {
	var tokenClaims model.TokenClaims
	_, err := jwt.ParseWithClaims(tokenStr, &tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return SigningKey(), nil
	})
	return &tokenClaims, err
}
