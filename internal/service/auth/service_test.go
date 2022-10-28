package auth

import (
	"strings"
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateAccessToken(t *testing.T) {
	tests := map[string]struct {
		timer  util.Timer
		userId int64
	}{
		"valid": {
			timer:  util.TestTimerFromUnix(0),
			userId: 245,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			jwt.TimeFunc = tcase.timer.Now
			util.SigningKey = func() []byte { return []byte("test") }
			cfg := cfg.TestConfig()
			cfg.SetTimer(tcase.timer)
			svc := NewAuthService(cfg, nil)
			tokenStr := svc.GenerateAccessToken(tcase.userId)
			parts := strings.Split(tokenStr, ".")
			assert.Equal(t, 3, len(parts))
		})
	}
}

func Test_GenerateRefreshToken(t *testing.T) {
	tests := map[string]struct {
		timer    util.Timer
		repo     AuthRepo
		userId   int64
		wantsErr bool
	}{
		"valid": {
			timer: util.TestTimerFromUnix(0),
			repo: &testAuthRepo{
				refreshToken: &RefreshToken{string: "hi"},
			},
			userId: 96,
		},
		"error from database": {
			timer: util.TestTimerFromUnix(0),
			repo: &testAuthRepo{
				err: resp.InternalServerError("boom"),
			},
			userId:   1254,
			wantsErr: true,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			jwt.TimeFunc = tcase.timer.Now
			util.SigningKey = func() []byte { return []byte("test") }
			cfg := cfg.TestConfig()
			cfg.SetTimer(tcase.timer)
			svc := NewAuthService(cfg, tcase.repo)
			tokenStr, err := svc.GenerateRefreshToken(tcase.userId)
			if tcase.wantsErr {
				assert.Error(t, err)
				assert.Equal(t, "", tokenStr)
			} else {
				assert.NoError(t, err)
				parts := strings.Split(tokenStr, ".")
				assert.Equal(t, 3, len(parts))
				refreshToken, _ := RefreshTokenFromTokenString(tokenStr)
				assert.Equal(t, tcase.userId, refreshToken.UserId())
			}
		})
	}
}

func Test_ValidateRefreshToken(t *testing.T) {
	tests := map[string]struct {
		timer        util.Timer
		repo         AuthRepo
		refreshToken *RefreshToken
		expectedErr  error
		expectedEOL  bool
	}{
		"valid": {
			timer: util.TestTimerFromUnix(0),
			repo:  &testAuthRepo{},
			refreshToken: &RefreshToken{
				issuedAt: 2,
			},
		},
		"error from database": {
			timer: util.TestTimerFromUnix(0),
			repo: &testAuthRepo{
				err: resp.ErrCouldNotValidateRefreshToken,
			},
			refreshToken: &RefreshToken{
				issuedAt: 2,
			},
			expectedErr: resp.ErrCouldNotValidateRefreshToken,
		},
		"error expired": {
			timer: util.TestTimerFromUnix(refreshTokenTimeLimit + 30),
			repo:  &testAuthRepo{},
			refreshToken: &RefreshToken{
				issuedAt: 30,
			},
			expectedErr: resp.ErrTokenIsExpired,
			expectedEOL: true,
		},
		"no error but close to EOL": {
			timer: util.TestTimerFromUnix(refreshTokenTimeLimit + 30),
			repo:  &testAuthRepo{},
			refreshToken: &RefreshToken{
				issuedAt: refreshTokenTimeLimit*.1 + 30,
			},
			expectedEOL: true,
		},
		"valid almost but not at EOL": {
			timer: util.TestTimerFromUnix(refreshTokenTimeLimit + 30),
			repo:  &testAuthRepo{},
			refreshToken: &RefreshToken{
				issuedAt: refreshTokenTimeLimit*.2 + 30,
			},
			expectedEOL: false,
		},
		"valid almost expired": {
			timer: util.TestTimerFromUnix(refreshTokenTimeLimit + 30),
			repo:  &testAuthRepo{},
			refreshToken: &RefreshToken{
				issuedAt: 31,
			},
			expectedEOL: true,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			jwt.TimeFunc = tcase.timer.Now
			util.SigningKey = func() []byte { return []byte("test") }
			cfg := cfg.TestConfig()
			cfg.SetTimer(tcase.timer)
			svc := NewAuthService(cfg, tcase.repo)
			err := svc.ValidateRefreshToken(tcase.refreshToken)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tcase.expectedEOL {
				assert.True(t, tcase.refreshToken.IsNearEndOfLife())
			} else {
				assert.False(t, tcase.refreshToken.IsNearEndOfLife())
			}
		})
	}
}

func Test_TokenFuncs(t *testing.T) {
	refreshToken, err := RefreshTokenFromTokenString("try parsing this")
	assert.Nil(t, refreshToken)
	assert.NotNil(t, err)
}

type testAuthRepo struct {
	err          error
	refreshToken *RefreshToken
}

func (trepo *testAuthRepo) storeRefreshToken(token *RefreshToken) error {
	return trepo.err
}

func (trepo *testAuthRepo) validateRefreshToken(token *RefreshToken) error {
	return trepo.err
}

func (trepo *testAuthRepo) getLatestRefreshToken(userId int64) *RefreshToken {
	return trepo.refreshToken
}
