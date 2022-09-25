package auth

import (
	"github.com/bchadwic/wordbubble/util"
)

type refreshToken struct {
	string
	issuedAt int64
	userId   int64
	nearEOL  bool
}

func RefreshTokenFromTokenString(tokenStr string) (*refreshToken, error) {
	claims, err := util.ParseWithClaims(tokenStr)
	if err != nil {
		return nil, err
	}
	return &refreshToken{
		string:   tokenStr,
		userId:   claims.UserId,
		issuedAt: claims.IssuedAt,
	}, nil
}

// returns true if this token is near the expiration time
func (tkn *refreshToken) IsAtEndOfLife() bool {
	return tkn.nearEOL
}

// returns the user id stored inside the token
func (tkn *refreshToken) UserId() int64 {
	return tkn.userId
}
