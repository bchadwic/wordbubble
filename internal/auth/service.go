package auth

import (
	"errors"
	"time"

	"github.com/bchadwic/wordbubble/util"
)

type authService struct {
	repo       AuthRepo
	log        util.Logger
	timer      util.Timer
	signingKey string
}

type refreshToken struct {
	string
	issuedAt int64
	userId   int64
	nearEOL  bool
}

func NewAuthService(log util.Logger, repo AuthRepo, timer util.Timer, signingKey string) *authService {
	util.SigningKey = func() []byte {
		return []byte(signingKey)
	}
	return &authService{
		log:        log,
		repo:       repo,
		timer:      timer,
		signingKey: signingKey,
	}
}

func (svc *authService) GenerateAccessToken(userId int64) string {
	now := svc.timer.Now()
	return util.GenerateSignedToken(now.Unix(), now.Add(accessTokenTimeLimit).Unix(), userId)
}

func (svc *authService) GenerateRefreshToken(userId int64) (string, error) {
	now := svc.timer.Now()
	token, _ := RefreshTokenFromTokenString(
		util.GenerateSignedToken(now.Unix(), now.Add(refreshTokenTimeLimit*time.Second).Unix(), userId),
	)
	if err := svc.repo.StoreRefreshToken(token); err != nil {
		return "", err
	}
	return token.string, nil
}

func (svc *authService) ValidateRefreshToken(token *refreshToken) (err error) {
	if err = svc.checkRefreshTokenExpiry(token); err != nil {
		return
	}
	if err = svc.repo.ValidateRefreshToken(token); err != nil {
		return
	}
	return
}

// sets EOL flag for token; returns error if token is expired
func (svc *authService) checkRefreshTokenExpiry(token *refreshToken) error {
	if timeLeft := refreshTokenTimeLimit - (svc.timer.Now().Unix() - token.issuedAt); timeLeft < ImminentExpirationWindow {
		token.nearEOL = true
		if timeLeft <= 0 {
			return errors.New("refresh token is expired, please login again")
		}
	}
	return nil
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
func (tkn *refreshToken) IsNearEndOfLife() bool {
	return tkn.nearEOL
}

// returns the user id stored inside the token
func (tkn *refreshToken) UserId() int64 {
	return tkn.userId
}
