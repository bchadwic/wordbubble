package util

import (
	"errors"

	"github.com/bchadwic/wordbubble/model"
	"github.com/golang-jwt/jwt"
)

var SigningKey func() []byte

func GenerateSignedToken(iat, exp, userId int64) string {
	signedToken, _ := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &model.TokenClaims{StandardClaims: jwt.StandardClaims{IssuedAt: iat, ExpiresAt: exp}, UserId: userId},
	).SignedString(SigningKey())
	return signedToken
}

func GetUserIdFromTokenString(tokenStr string) (int64, error) {
	if tokenClaims, err := ParseWithClaims(tokenStr); err != nil {
		return 0, errors.New("token signature was found to be invalid")
	} else {
		return tokenClaims.UserId, nil
	}
}

func ParseWithClaims(tokenStr string) (*model.TokenClaims, error) {
	tokenClaims := &model.TokenClaims{} // TODO come back to this mapping
	_, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return SigningKey(), nil
	})
	return tokenClaims, err
}
