package auth

import (
	"database/sql"
	"testing"

	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func NewTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func Test_HappyPath(t *testing.T) {
	repo := NewAuthRepo(util.TestLogger(), NewTestDB())
	tokenStr, userId, issuedAt := "im.a.token", int64(56), int64(234)

	// A user signs up or logins in. Refresh token is saved after being generated
	token := &refreshToken{
		string:   tokenStr,
		userId:   userId,
		issuedAt: issuedAt,
	}
	err := repo.StoreRefreshToken(token)
	assert.NoError(t, err)

	// A little while later a user needs to validate their refresh token to get a new access token
	token = &refreshToken{
		string: tokenStr,
		userId: userId,
	}
	err = repo.ValidateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, issuedAt, token.issuedAt)

	// On a different device, a user uses a close to EOL token thus triggering a GetLatestRefreshToken
	token = repo.GetLatestRefreshToken(userId)
	assert.NotNil(t, token)
	assert.Equal(t, tokenStr, token.string)
	assert.Equal(t, userId, token.UserId())
	assert.Equal(t, issuedAt, token.issuedAt)

	// The user has been away from some time, time to clean up the token
	repo.CleanupExpiredRefreshTokens(issuedAt + 1)
	// when the user comes back, they are faced with a login flow
	token = repo.GetLatestRefreshToken(userId)
	assert.Nil(t, token)

	// malicious user tries to validate an old token
	token = &refreshToken{
		string: tokenStr,
		userId: userId,
	}
	err = repo.ValidateRefreshToken(token)
	assert.NotNil(t, err)
	assert.Equal(t, "could not validate issued time of refresh token, please login again", err.Error())
}

func Test_NotSoHappyPath(t *testing.T) {
	repo := NewAuthRepo(util.TestLogger(), NewTestDB())

	// db closed
	repo.db.Close()
	err := repo.StoreRefreshToken(&refreshToken{})
	assert.NotNil(t, err)
	assert.Equal(t, "could not successfully store refresh token on server", err.Error())

	err = repo.CleanupExpiredRefreshTokens(0)
	assert.NotNil(t, err)
	assert.Equal(t, "an error occurred cleaning up expired refresh tokens", err.Error())
}
