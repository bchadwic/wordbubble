package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/bchadwic/wordbubble/util"
	"github.com/golang-jwt/jwt"
)

const (
	refreshTokenTimeLimit    = 60
	accessTokenTimeLimit     = 10 * time.Second // change me to something quicker
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .2)
)

var cleanupExpiredRefreshTokensStatement = `DELETE FROM tokens WHERE issued_at < ?`

type AuthService interface {
	GenerateAccessToken(userId int64) (string, error)
	GenerateRefreshToken(userId int64) (string, error)
	GetUserIdFromTokenString(tokenStr string) (int64, error)
	VerifyTokenAgainstAuthSource(userId int64, tokenStr string) (int64, error)
	GetOrCreateLatestRefreshToken(userId int64) string
}

type authService struct {
	repo       AuthRepo
	log        util.Logger
	signingKey string
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}

type refreshToken struct {
	string
	issuedAt int64
	userId   int64
}

func NewAuthService(repo AuthRepo, logger util.Logger, signingKey string) *authService {
	return &authService{repo, logger, signingKey}
}

// TODO combine GenerateAccessToken and GenerateRefreshToken?
func (svc *authService) GenerateAccessToken(userId int64) (string, error) {
	iat := time.Now().Unix()
	exp := time.Now().Add(accessTokenTimeLimit).Unix()

	fmt.Printf("iat: %d\nexp: %d\n", iat, exp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(accessTokenTimeLimit).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})
	signedToken, err := token.SignedString([]byte(svc.signingKey))
	if err != nil {
		svc.log.Error("failed to create access token for user: %d, error: %s", userId, err)
		return "", errors.New("failed to sign and generate an access token")
	}
	return signedToken, nil
}

func (svc *authService) GenerateRefreshToken(userId int64) (string, error) {
	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(refreshTokenTimeLimit * time.Second).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	}).SignedString([]byte(svc.signingKey))
	if err != nil {
		svc.log.Error("failed to create access token for user: %d, error: %s", userId, err)
		return "", errors.New("failed to sign and generate a refresh token")
	}
	token := &refreshToken{
		string:   signedToken,
		userId:   userId,
		issuedAt: time.Now().Unix(),
	}
	if err := svc.repo.StoreRefreshToken(token); err != nil {
		return "", err
	}
	return signedToken, nil
}

func (svc *authService) GetUserIdFromTokenString(tokenStr string) (int64, error) {
	tokenClaims := &tokenClaims{} // TODO come back to this mapping
	token, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(svc.signingKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			svc.log.Error("signature is invalid, error: %s", err)
			return 0, errors.New("token signature was found to be invalid")
		}
		svc.log.Error("an error occurred while parsing token, defaulting to expiration: error %s", err)
		return 0, errors.New("access token is expired")
	}
	if !token.Valid { // only applicable to access tokens
		svc.log.Error("token is expired for user: %d, error: %s", tokenClaims.UserId, err)
		return 0, errors.New("access token is expired")
	}
	return tokenClaims.UserId, nil
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
	timeDiff := time.Now().Unix() - issuedAt
	if timeDiff >= refreshTokenTimeLimit {
		svc.log.Error("token was found to be expired for user: %d", userId)
		return 0, errors.New("refresh token is expired, please login again")
	}
	return refreshTokenTimeLimit - timeDiff, nil
}

func (svc *authService) GetOrCreateLatestRefreshToken(userId int64) string {
	token := svc.repo.GetLatestRefreshToken(userId)
	if token != nil { // if there is a token that isn't close to dying, use that
		if timeRemaining := time.Now().Unix() - token.issuedAt; timeRemaining > ImminentExpirationWindow {
			return token.string
		}
	} // otherwise create a new one
	newRefreshToken, _ := svc.GenerateRefreshToken(userId)
	return newRefreshToken
}
