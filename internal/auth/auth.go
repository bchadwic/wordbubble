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

type AuthService interface {
	// Generates an access token
	GenerateAccessToken(userId int64) string
	// Generates a refresh token, an error is generated when the token couldn't be successfully saved to the database
	GenerateRefreshToken(userId int64) (string, error)
	// Validates the refresh token string passed using the signing key and by checking the auth datasource
	ValidateRefreshToken(token *refreshToken) error
}

type AuthRepo interface {
	// Store a refresh token in the backend datasource
	StoreRefreshToken(token *refreshToken) error
	// Validate a refresh token against the backend datasource
	ValidateRefreshToken(token *refreshToken) error
	// Find and return the latest refresh token for a user
	GetLatestRefreshToken(userId int64) *refreshToken
}

type AuthCleaner interface {
	// Remove any refresh tokens from the database that are expired
	CleanupExpiredRefreshTokens(since int64) error
}
