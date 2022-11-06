package auth

import "time"

const (
	refreshTokenTimeLimit    = 36000 // 10 minutes
	accessTokenTimeLimit     = 30 * time.Second
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .2) // TODO make better?

	CleanupExpiredRefreshTokens = `DELETE FROM tokens WHERE issued_at < $1`
	StoreRefreshToken           = `INSERT INTO tokens (user_id, refresh_token, issued_at) VALUES ($1, $2, $3)`
	ValidateRefreshToken        = `SELECT issued_at FROM tokens WHERE user_id = $1 AND refresh_token = $2`
	GetLatestRefreshToken       = `SELECT refresh_token, issued_at FROM tokens WHERE user_id = $1 ORDER BY issued_at DESC LIMIT 1`
)

// AuthService is the interface that the application
// uses to interact with access and refresh tokens
type AuthService interface {
	// GenerateAccessToken generates an access token
	// string is the access token's string
	GenerateAccessToken(userId int64) string
	// GenerateRefreshToken generates a refresh token, an error is generated when the token couldn't be successfully saved to the database
	// string is the refresh token's string, or empty string.
	// error could be (500) resp.ErrCouldNotStoreRefreshToken or nil.
	GenerateRefreshToken(userId int64) (string, error)
	// ValidateRefreshToken validates the refresh token string passed using the signing key and by checking the auth datasource
	// error could be (401) ErrTokenIsExpired, (401) resp.ErrCouldNotValidateRefreshToken or nil.
	ValidateRefreshToken(token *RefreshToken) error
}

// AuthRepo is the interface that the service layer
// uses to interact with refresh tokens in the database
type AuthRepo interface {
	// storeRefreshToken stores a refresh token in the database.
	// error can be (500) resp.ErrCouldNotStoreRefreshToken or nil.
	storeRefreshToken(token *RefreshToken) error
	// validateRefreshToken validates a refresh token against database.
	// error can be (401) resp.ErrCouldNotValidateRefreshToken or nil.
	validateRefreshToken(token *RefreshToken) error
	// getLatestRefreshToken find and return the latest refresh token for a user in the database.
	// refreshToken can be nil if there is no latest refresh token
	getLatestRefreshToken(userId int64) *RefreshToken
}

// AuthCleaner is the interface that the application
// uses to clean up expired refresh tokens
type AuthCleaner interface {
	// CleanupExpiredRefreshTokens remove any refresh tokens from the database that are expired.
	// error can be (500) resp.ErrCouldNotCleanupTokens or nil
	CleanupExpiredRefreshTokens(since int64) error
}
