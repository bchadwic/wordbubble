package auth

import (
	"errors"
	"time"

	"github.com/bchadwic/wordbubble/util"
)

const (
	refreshTokenTimeLimit    = 60
	accessTokenTimeLimit     = 10 * time.Second // change me to something quicker
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .2)
)

var cleanupExpiredRefreshTokensStatement = `DELETE FROM tokens WHERE issued_at < ?`

type AuthService interface {
	GenerateAccessToken(userId int64) string
	// Generates a refresh token, it is possible to get a refresh token back with an error.
	// An error is generated when the token couldn't be successfully saved to the database
	GenerateRefreshToken(userId int64) (string, error)
	VerifyTokenAgainstAuthSource(userId int64, tokenStr string) (int64, error)
	GetOrCreateLatestRefreshToken(userId int64) string
}

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
}

func NewAuthService(log util.Logger, repo AuthRepo, timer util.Timer, signingKey string) *authService {
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
	token := &refreshToken{
		string:   util.GenerateSignedToken(now.Unix(), now.Add(refreshTokenTimeLimit*time.Second).Unix(), userId),
		userId:   userId,
		issuedAt: now.Unix(),
	}
	return token.string, svc.repo.StoreRefreshToken(token)
}

func (svc *authService) VerifyTokenAgainstAuthSource(userId int64, tokenStr string) (int64, error) {
	token := &refreshToken{
		string: tokenStr,
		userId: userId,
	}
	issuedAt, err := svc.repo.ValidateRefreshToken(token)
	if err != nil {
		return 0, err
	}
	timeDiff := svc.timer.Now().Unix() - issuedAt
	if timeDiff >= refreshTokenTimeLimit {
		svc.log.Error("token was found to be expired for user: %d", userId)
		return 0, errors.New("refresh token is expired, please login again")
	}
	return refreshTokenTimeLimit - timeDiff, nil
}

func (svc *authService) GetOrCreateLatestRefreshToken(userId int64) string {
	token := svc.repo.GetLatestRefreshToken(userId)
	if token != nil { // if there is a token that isn't close to dying, use that
		if timeRemaining := svc.timer.Now().Unix() - token.issuedAt; timeRemaining > ImminentExpirationWindow {
			return token.string
		}
	} // otherwise create a new one
	newRefreshToken, _ := svc.GenerateRefreshToken(userId)
	return newRefreshToken
}
