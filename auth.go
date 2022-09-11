package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	minPasswordLength        = 6
	maxUsernameLength        = 40
	maxEmailLength           = 320
	refreshTokenTimeLimit    = 60
	accessTokenTimeLimit     = 30 * time.Second
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .8)
)

var cleanupExpiredRefreshTokensStatement = `DELETE FROM tokens WHERE issued_at < ?`

type Auth interface {
	GenerateAccessToken(logger Logger, userId int64) (string, error)
	GenerateRefreshToken(logger Logger, userId int64) (string, error)
	GetUserIdFromTokenString(logger Logger, tokenStr string) (int64, error)
	VerifyTokenAgainstAuthSource(logger Logger, userId int64, tokenStr string) (int64, error)
	GetOrCreateLatestRefreshToken(logger Logger, userId int64) string
}

type AuthSource interface {
	StoreRefreshToken(logger Logger, userId int64, refreshToken string) error
	ValidateRefreshToken(logger Logger, userId int64, refreshToken string) (int64, error)
	RemoveRefreshToken(logger Logger, userId int64, refreshToken string) error
	CleanupExpiredRefreshTokens(logger Logger)
	GetLatestRefreshToken(logger Logger, userId int64) (string, int64)
}
type auth struct {
	source     AuthSource
	signingKey string
}
type authSource struct {
	db *sql.DB
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}

func NewAuthSource() *authSource {
	panicker := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	db, err := sql.Open("sqlite3", "./wb-database.db")
	panicker(err)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tokens (
		user_id INTEGER NOT NULL,  
		refresh_token TEXT NOT NULL,
		issued_at INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`)
	panicker(err)
	return &authSource{db}
}

func NewAuth(source AuthSource, signingKey string) *auth {
	return &auth{source, signingKey}
}

func (auth *auth) GenerateAccessToken(logger Logger, userId int64) (string, error) {
	logger.Info("auth.GenerateAccessToken: generating access token for user: %d", userId)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(accessTokenTimeLimit).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})
	signedToken, err := token.SignedString([]byte(auth.signingKey))
	if err != nil {
		logger.Error("auth.GenerateAccessToken: couldn't sign the newly created token for user: %d, error: %s", userId, err)
		return "", errors.New("failed to sign and generate access token")
	}
	logger.Info("auth.GenerateAccessToken: successfully generated a token for user: %d", userId)
	return signedToken, nil
}

func (auth *auth) GenerateRefreshToken(logger Logger, userId int64) (string, error) {
	logger.Info("auth.GenerateRefreshToken: generating refresh token for user: %d", userId)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{},
		userId,
	})
	signedToken, err := token.SignedString([]byte(auth.signingKey))
	if err != nil {
		logger.Error("auth.GenerateRefreshToken: failed to sign the new refresh token for user: %d, error: %s", userId, err)
		return "", errors.New("failed to sign and generate a new refresh token")
	}
	if err := auth.source.StoreRefreshToken(logger, userId, signedToken); err != nil {
		return "", err
	}
	logger.Info("auth.GenerateRefreshToken: successfully generated a new refresh token for user: %d", userId)
	return signedToken, nil
}

func (auth *auth) GetUserIdFromTokenString(logger Logger, tokenStr string) (int64, error) {
	logger.Info("auth.GetUserIdFromTokenString: parsing token claims and receiving userId")

	tokenClaims := &tokenClaims{} // TODO come back to this mapping
	token, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(auth.signingKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			logger.Error("auth.GetUserIdFromTokenString: signature is invalid, error: %s", err)
			return 0, errors.New("token signature was found to be invalid")
		}
		logger.Error("auth.GetUserIdFromTokenString: an error occurred while parsing token, defaulting to expiration: error %s", err)
		return 0, errors.New("access token is expired")
	}
	if !token.Valid { // only applicable to access tokens
		logger.Warn("auth.GetUserIdFromTokenString: token is expired for user: %d, error: %s", tokenClaims.UserId, err)
		return 0, errors.New("access token is expired")
	}
	logger.Info("auth.GetUserIdFromTokenString: successfully grabbed userId from access token, user: %d", tokenClaims.UserId)
	return tokenClaims.UserId, nil
}

func (auth *auth) VerifyTokenAgainstAuthSource(logger Logger, userId int64, tokenStr string) (int64, error) {
	issuedAt, err := auth.source.ValidateRefreshToken(logger, userId, tokenStr)
	if err != nil {
		return 0, err
	}
	timeDiff := time.Now().Unix() - issuedAt
	if timeDiff >= refreshTokenTimeLimit {
		logger.Error("auth.VerifyTokenAgainstAuthSource: token was found to be expired for user: %d", userId)
		return 0, errors.New("refresh token is expired, please login again")
	}
	return refreshTokenTimeLimit - timeDiff, nil
}

func (auth *auth) GetOrCreateLatestRefreshToken(logger Logger, userId int64) string {
	logger.Info("GetOrCreateLatestRefreshToken: grabbing the latest token in the database for user: %d", userId)
	latestRefreshToken, issuedAt := auth.source.GetLatestRefreshToken(logger, userId)
	if latestRefreshToken != "" { // if there is a token that isn't close to dying, use that
		if refreshTokenTimeLimit-timeSinceIssued(issuedAt) > ImminentExpirationWindow {
			return latestRefreshToken
		}
	}
	// otherwise create a new one
	logger.Info("GetOrCreateLatestRefreshToken: generating a new refresh token for user: %d", userId)
	newRefreshToken, _ := auth.GenerateRefreshToken(logger, userId)
	return newRefreshToken
}

func timeSinceIssued(issuedAt int64) int64 {
	return time.Now().Unix() - issuedAt
}

func (source *authSource) StoreRefreshToken(logger Logger, userId int64, refreshToken string) error {
	logger.Info("auth.StoreRefreshToken: storing new token for user: %d", userId)

	stmt, err := source.db.Prepare(`INSERT INTO tokens (user_id, refresh_token, issued_at) VALUES (?, ?, ?)`)
	if err != nil {
		logger.Error("auth.StoreRefreshToken: could not store token for user: %d, error: %s", userId, err)
		return errors.New("could not successfully prepare refresh token for storage on server")
	}
	_, err = stmt.Exec(userId, refreshToken, time.Now().Unix())
	if err != nil {
		logger.Error("auth.StoreRefreshToken: could not execute statement for user: %d, error: %s", err)
		return errors.New("could not successfully store refresh token on server")
	}
	logger.Info("auth.StoreRefreshToken: successfully stored new refresh token for user: %d", userId)
	return nil
}

func (source *authSource) ValidateRefreshToken(logger Logger, userId int64, refreshToken string) (int64, error) {
	logger.Info("auth.ValidateRefreshToken: validating refresh token for user: %d", userId)

	stmt, err := source.db.Prepare(`SELECT issued_at FROM tokens WHERE user_id = ? AND refresh_token = ?`)
	if err != nil {
		logger.Error("auth.ValidateRefreshToken: could not validate token for user: %d:, error: %s", userId, err)
		return 0, errors.New("could not prepare token for validation against server")
	}
	row, err := stmt.Query(userId, refreshToken)
	if err != nil {
		logger.Error("auth.ValidateRefreshToken: could not query for refresh token for user: %d, error: %s", userId, err)
		return 0, errors.New("could not retrieve server refresh tokens for user")
	}
	defer row.Close()
	if !row.Next() {
		logger.Error("auth.ValidateRefreshToken: no tokens matched what was passed in for user: %d", userId)
		return 0, errors.New("could not validate refresh token, please login again")
	}
	var issuedAt int64
	if err := row.Scan(&issuedAt); err != nil {
		logger.Error("auth.ValidateRefreshToken: could not map rows for user: %d, error: %s", userId, err)
		return 0, errors.New("could not validate issued time of refresh token, please login again")
	}
	logger.Info("auth.ValidateRefreshToken: successfully matched refresh token sent with a token found in db for user: %d", userId)
	return issuedAt, nil
}

func (source *authSource) RemoveRefreshToken(logger Logger, userId int64, refreshToken string) error {
	logger.Info("auth.RemoveRefreshToken: removing refresh token for user: %d", userId)

	stmt, err := source.db.Prepare(`DELETE FROM tokens WHERE user_id = ? AND refresh_token = ?`)
	if err != nil {
		logger.Error("auth.RemoveRefreshToken: could not remove token for user: %d:, error: %s", userId, err)
		return errors.New("could not prepare task to remove existing refresh token")
	}
	rs, err := stmt.Exec(userId, refreshToken)
	if err != nil {
		logger.Error("auth.RemoveRefreshToken: could not execute refresh token deletion for user: %d, error: %s", userId, err)
		return errors.New("could not execute the removal of existing refresh token")
	}
	amt, err := rs.RowsAffected()
	if err != nil {
		logger.Error("auth.RemoveRefreshToken: could not successfully determine if the token was deleted for user: %d, error: %s", userId, err)
		return errors.New("could not determine if refresh token was deleted successfully")
	}
	if amt < 1 {
		logger.Error("auth.RemoveRefreshToken: no records were deleted %d", userId)
		return errors.New("could not successfully delete refresh token")
	}
	logger.Info("auth.RemoveRefreshToken: successfully deleted refresh token for user: %d", userId)
	return nil
}

// TODO fix me
func (source *authSource) CleanupExpiredRefreshTokens(logger Logger) {
	logger.Info("auth.CleanupExpiredRefreshTokens: cleaning up tokens than the refreshTokenTimeLimit: %d", refreshTokenTimeLimit)
	stmt, err := source.db.Prepare(cleanupExpiredRefreshTokensStatement)
	if err != nil {
		logger.Error("auth.CleanupExpiredRefreshTokens: could not execute delete tokens: error: %s", err)
		return
	}
	rs, err := stmt.Exec(time.Now().Unix() - refreshTokenTimeLimit)
	if err != nil {
		logger.Error("auth.CleanupExpiredRefreshTokens: could not execute delete tokens: error: %s", err)
		return
	}
	amt, err := rs.RowsAffected()
	if err != nil {
		logger.Error("auth.CleanupExpiredRefreshTokens: could not successfully determine the amount of tokens that were deleted, error: %s", err)
		return
	}
	logger.Info("auth.CleanupExpiredRefreshTokens: successfully deleted %d tokens", amt)
}

func (source *authSource) GetLatestRefreshToken(logger Logger, userId int64) (string, int64) {
	logger.Info("auth.GetLatestRefreshToken: getting the latest token for user: %d", userId)

	stmt, err := source.db.Prepare(`SELECT refresh_token, issued_at FROM tokens WHERE user_id = ? ORDER BY issued_at DESC LIMIT 1`)
	if err != nil {
		logger.Error("auth.GetLatestRefreshToken: could not get latest refresh token for user: %d:, error: %s", userId, err)
		return "", 0
	}
	row, err := stmt.Query(userId)
	if err != nil {
		logger.Error("auth.GetLatestRefreshToken: could not query for refresh token for user: %d, error: %s", userId, err)
		return "", 0
	}
	defer row.Close()
	if !row.Next() {
		logger.Error("auth.GetLatestRefreshToken: no refresh tokens were found for user: %d", userId)
		return "", 0
	}
	var issuedAt int64
	var latestRefreshToken string
	if err := row.Scan(&latestRefreshToken, &issuedAt); err != nil {
		logger.Error("auth.GetLatestRefreshToken: could not map the latest refresh token for user: %d, error: %s", userId, err)
		return "", 0
	}
	logger.Info("auth.GetLatestRefreshToken: successfully matched refresh token sent with a token found in db for user: %d", userId)
	return latestRefreshToken, issuedAt
}
