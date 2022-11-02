package auth

import (
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/stretchr/testify/assert"
)

func Test_HappyPath(t *testing.T) {
	repo := NewAuthRepo(cfg.TestConfig())
	tokenStr, userId, issuedAt := "im.a.token", int64(56), int64(234)

	// A user signs up or logins in. Refresh token is saved after being generated
	token := &RefreshToken{
		string:   tokenStr,
		userId:   userId,
		issuedAt: issuedAt,
	}
	err := repo.storeRefreshToken(token)
	assert.NoError(t, err)

	// A little while later a user needs to validate their refresh token to get a new access token
	token = &RefreshToken{
		string: tokenStr,
		userId: userId,
	}
	err = repo.validateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, issuedAt, token.issuedAt)

	// On a different device, a user uses a close to EOL token thus triggering a GetLatestRefreshToken
	token = repo.getLatestRefreshToken(userId)
	assert.NotNil(t, token)
	assert.Equal(t, tokenStr, token.string)
	assert.Equal(t, userId, token.UserId())
	assert.Equal(t, issuedAt, token.issuedAt)

	// The user has been away from some time, time to clean up the token
	repo.CleanupExpiredRefreshTokens(issuedAt + 1)
	// when the user comes back, they are faced with a login flow
	token = repo.getLatestRefreshToken(userId)
	assert.Nil(t, token)

	// malicious user tries to validate an old token
	token = &RefreshToken{
		string: tokenStr,
		userId: userId,
	}
	err = repo.validateRefreshToken(token)
	assert.NotNil(t, err)
	assert.Equal(t, resp.ErrCouldNotValidateRefreshToken.Error(), err.Error())
}

func Test_NotSoHappyPath(t *testing.T) {
	repo := NewAuthRepo(cfg.TestConfig())

	// db closed
	repo.db.Close()
	err := repo.storeRefreshToken(&RefreshToken{})
	assert.NotNil(t, err)
	assert.Equal(t, resp.ErrCouldNotStoreRefreshToken.Error(), err.Error())

	err = repo.CleanupExpiredRefreshTokens(0)
	assert.NotNil(t, err)
	assert.Equal(t, resp.ErrCouldNotCleanupTokens.Error(), err.Error())
}
