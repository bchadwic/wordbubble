package auth

import "time"

const (
	refreshTokenTimeLimit    = 60
	accessTokenTimeLimit     = 10 * time.Second // change me to something quicker
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .2) // TODO make better?

	CleanupExpiredRefreshTokens = `DELETE FROM tokens WHERE issued_at < ?`
	StoreRefreshToken           = `INSERT INTO tokens (user_id, refresh_token, issued_at) VALUES (?, ?, ?)`
	ValidateRefreshToken        = `SELECT issued_at FROM tokens WHERE user_id = ? AND refresh_token = ?`
	GetLatestRefreshToken       = `SELECT refresh_token, issued_at FROM tokens WHERE user_id = ? ORDER BY issued_at DESC LIMIT 1`
)

// AuthService is the interface that the application
// uses to interact with access and refresh tokens
type AuthService interface {
	// generates an access token
	GenerateAccessToken(userId int64) string
	// generates a refresh token, an error is generated when the token couldn't be successfully saved to the database
	GenerateRefreshToken(userId int64) (string, error)
	// validates the refresh token string passed using the signing key and by checking the auth datasource
	ValidateRefreshToken(token *refreshToken) error
}

// AuthRepo is the interface that the service layer
// uses to interact with refresh tokens in the database
type AuthRepo interface {
	// store a refresh token in the database
	storeRefreshToken(token *refreshToken) error
	// validate a refresh token against database
	validateRefreshToken(token *refreshToken) error
	// find and return the latest refresh token for a user in the database
	getLatestRefreshToken(userId int64) *refreshToken
}

// AuthCleaner is the interface that the application
// uses to clean up expired refresh tokens
type AuthCleaner interface {
	// remove any refresh tokens from the database that are expired
	CleanupExpiredRefreshTokens(since int64) error
}
