package auth

import "time"

const (
	refreshTokenTimeLimit    = 60
	accessTokenTimeLimit     = 10 * time.Second // change me to something quicker
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .2) // TODO make better?
	// SQL statements
	cleanupExpiredRefreshTokensStatement = `DELETE FROM tokens WHERE issued_at < ?`
)

type AuthService interface {
	// Generates an access token
	GenerateAccessToken(userId int64) string
	// Generates a refresh token, an error is generated when the token couldn't be successfully saved to the database
	GenerateRefreshToken(userId int64) (string, error)
	// Validates the refresh token string passed using the signing key and by checking the auth datasource
	ValidateRefreshToken(token *refreshToken) error
}

type AuthRepo interface {
	StoreRefreshToken(token *refreshToken) error
	ValidateRefreshToken(token *refreshToken) error
	GetLatestRefreshToken(userId int64) *refreshToken
}
