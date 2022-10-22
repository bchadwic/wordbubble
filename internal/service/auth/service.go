package auth

import (
	"time"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

type authService struct {
	log   util.Logger
	timer util.Timer
	repo  AuthRepo
}

type refreshToken struct {
	string
	issuedAt int64
	userId   int64
	nearEOL  bool
}

func NewAuthService(cfg cfg.Config, repo AuthRepo) *authService {
	return &authService{
		log:   cfg.NewLogger("auth"),
		timer: cfg.Timer(),
		repo:  repo,
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
	if err := svc.repo.storeRefreshToken(token); err != nil {
		return "", err
	}
	return token.string, nil
}

func (svc *authService) ValidateRefreshToken(token *refreshToken) (err error) {
	if err = svc.checkRefreshTokenExpiry(token); err != nil {
		return
	}
	if err = svc.repo.validateRefreshToken(token); err != nil {
		return
	}
	return
}

// sets EOL flag for token; returns error if token is expired
func (svc *authService) checkRefreshTokenExpiry(token *refreshToken) error {
	if timeLeft := refreshTokenTimeLimit - (svc.timer.Now().Unix() - token.issuedAt); timeLeft < ImminentExpirationWindow {
		token.nearEOL = true
		if timeLeft <= 0 {
			return resp.ErrRefreshTokenIsExpired
		}
	}
	return nil
}

func RefreshTokenFromTokenString(tokenStr string) (*refreshToken, error) {
	claims, err := util.ParseWithClaims(tokenStr)
	if err != nil {
		return nil, resp.ErrParseRefreshToken
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