package auth

import (
	"errors"
	"fmt"
	"strings"
	"testing"

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
			timer:  util.Unix(0),
			userId: 245,
		},
	}

	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			jwt.TimeFunc = tcase.timer.Now
			svc := NewAuthService(util.TestLogger(), nil, tcase.timer, "test signing key")
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
			timer: util.Unix(0),
			repo: &TestAuthRepo{
				refreshToken: &refreshToken{string: "hi"},
			},
			userId: 96,
		},
		"error from database": {
			timer: util.Unix(0),
			repo: &TestAuthRepo{
				err: errors.New("explosion"),
			},
			userId:   1254,
			wantsErr: true,
		},
	}

	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			jwt.TimeFunc = tcase.timer.Now
			svc := NewAuthService(util.TestLogger(), tcase.repo, tcase.timer, "test signing key")
			tokenStr, err := svc.GenerateRefreshToken(tcase.userId)
			if tcase.wantsErr {
				assert.Error(t, err)
				assert.Equal(t, "", tokenStr)
			} else {
				assert.NoError(t, err)
				parts := strings.Split(tokenStr, ".")
				assert.Equal(t, 3, len(parts))
			}
		})
	}
}

func Test_ValidateRefreshToken(t *testing.T) {
	tests := map[string]struct {
		timer        util.Timer
		repo         AuthRepo
		refreshToken *refreshToken
		expectedErr  string
		expectedEOL  bool
	}{
		"valid": {
			timer: util.Unix(0),
			repo:  &TestAuthRepo{},
			refreshToken: &refreshToken{
				issuedAt: 2,
			},
		},
		"error from database": {
			timer: util.Unix(0),
			repo: &TestAuthRepo{
				err: fmt.Errorf("could not validate token"),
			},
			refreshToken: &refreshToken{
				issuedAt: 2,
			},
			expectedErr: "could not validate token",
		},
		"error expired": {
			timer: util.Unix(refreshTokenTimeLimit + 30),
			repo:  &TestAuthRepo{},
			refreshToken: &refreshToken{
				issuedAt: 30,
			},
			expectedErr: "refresh token is expired, please login again",
			expectedEOL: true,
		},
		"no error but close to EOL": {
			timer: util.Unix(refreshTokenTimeLimit + 30),
			repo:  &TestAuthRepo{},
			refreshToken: &refreshToken{
				issuedAt: refreshTokenTimeLimit*.1 + 30,
			},
			expectedEOL: true,
		},
		"valid almost but not at EOL": {
			timer: util.Unix(refreshTokenTimeLimit + 30),
			repo:  &TestAuthRepo{},
			refreshToken: &refreshToken{
				issuedAt: refreshTokenTimeLimit*.2 + 30,
			},
			expectedEOL: false,
		},
		"valid almost expired": {
			timer: util.Unix(refreshTokenTimeLimit + 30),
			repo:  &TestAuthRepo{},
			refreshToken: &refreshToken{
				issuedAt: 31,
			},
			expectedEOL: true,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			jwt.TimeFunc = tcase.timer.Now
			svc := NewAuthService(nil, tcase.repo, tcase.timer, "test signing key")
			err := svc.ValidateRefreshToken(tcase.refreshToken)
			if tcase.expectedErr != "" {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tcase.expectedEOL {
				assert.True(t, tcase.refreshToken.IsAtEndOfLife())
			} else {
				assert.False(t, tcase.refreshToken.IsAtEndOfLife())
			}
		})
	}
}
