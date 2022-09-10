package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	minPasswordLength    = 6
	maxUsernameLength    = 40
	maxEmailLength       = 320
	refreshTokenDayLimit = 10
)

var cleanupExpiredRefreshTokensStatement = fmt.Sprintf(`DELETE FROM tokens WHERE created_timestamp < date('now', '-%d days')`, refreshTokenDayLimit)

type AuthSource interface {
	StoreRefreshToken(logger Logger, userId int64, refreshToken string) error
	ValidateRefreshToken(logger Logger, userId int64, refreshToken string) error
	RemoveRefreshToken(logger Logger, userId int64, refreshToken string) error
	CleanupExpiredRefreshTokens(logger Logger)
}
type Auth interface {
	GenerateAccessToken(logger Logger, user *User) (string, error)
	ValidateAccessTokenAndReceiveId(logger Logger, tokenStr string) (int64, error)
}

type authSource struct {
	db *sql.DB
}

type auth struct {
	source     AuthSource
	signingKey string
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
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
		token_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,  
		refresh_token TEXT NOT NULL,
		created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`)
	panicker(err)
	return &authSource{db}
}

func NewAuth(source AuthSource, signingKey string) *auth {
	return &auth{source, signingKey}
}

func (auth *auth) GenerateAccessToken(logger Logger, user *User) (string, error) {
	logger.Info("auth.GenerateAccessToken: generating access token for user: %d", user.UserId)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.UserId,
		user.Username,
	})
	signedToken, err := token.SignedString([]byte(auth.signingKey))
	if err != nil {
		logger.Error("auth.GenerateAccessToken: couldn't sign the newly created token for user: %d, error: %s", user.UserId, err)
		return "", errors.New("failed to sign and generate access token")
	}
	logger.Info("auth.GenerateAccessToken: successfully generated a token for user: %d", user.UserId)
	return signedToken, nil
}

func (auth *auth) GenerateRefreshToken(logger Logger, user *User) (string, error) {
	logger.Info("auth.GenerateRefreshToken: generating refresh token for user: %d", user.UserId)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(24 * time.Hour * refreshTokenDayLimit).Unix(),
		IssuedAt:  time.Now().Unix(),
	})
	signedToken, err := token.SignedString([]byte(auth.signingKey))
	if err != nil {
		logger.Error("auth.GenerateRefreshToken: failed to sign the new refresh token for user: %d, error: %s", user.UserId, err)
		return "", errors.New("failed to sign and generate a new refresh token")
	}
	if err := auth.source.StoreRefreshToken(logger, user.UserId, signedToken); err != nil {
		return "", err
	}
	logger.Info("auth.GenerateRefreshToken: successfully generated a new refresh token for user: %d", user.UserId)
	return signedToken, nil
}

func (auth *auth) ValidateAccessTokenAndReceiveId(logger Logger, tokenStr string) (int64, error) {
	logger.Info("auth.ValidateAccessTokenAndReceiveId: parsing token claims and receiving userId")

	tokenClaims := &tokenClaims{} // TODO come back to this mapping
	token, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(auth.signingKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			logger.Error("auth.ValidateAccessTokenAndReceiveId: signature is invalid, error: %s", err)
			return 0, errors.New("access token signature was found to be invalid")
		}
		logger.Error("auth.ValidateAccessTokenAndReceiveId: could not parse token: error %s", err)
		return 0, errors.New("could not parse the access token")
	}
	if !token.Valid {
		return 0, errors.New("access token is expired")
	}
	logger.Info("auth.ValidateAccessTokenAndReceiveId: successfully grabbed userId from access token, user: %d", tokenClaims.UserId)
	return tokenClaims.UserId, nil
}

func (source *authSource) StoreRefreshToken(logger Logger, userId int64, refreshToken string) error {
	logger.Info("auth.StoreRefreshToken: storing new token for user: %d", userId)

	stmt, err := source.db.Prepare(`INSERT INTO tokens (user_id, refresh_token) VALUES (?, ?)`)
	if err != nil {
		logger.Error("auth.StoreRefreshToken: could not store token for user: %d, error: %s", userId, err)
		return errors.New("could not successfully prepare refresh token for storage on server")
	}
	_, err = stmt.Exec(userId, refreshToken)
	if err != nil {
		logger.Error("auth.StoreRefreshToken: could not execute statement for user: %d, error: %s", err)
		return errors.New("could not successfully store refresh token on server")
	}
	logger.Info("auth.StoreRefreshToken: successfully stored new refresh token for user: %d", userId)
	return nil
}

func (source *authSource) ValidateRefreshToken(logger Logger, userId int64, refreshToken string) error {
	logger.Info("auth.ValidateRefreshToken: validating refresh token for user: %d", userId)

	stmt, err := source.db.Prepare(`SELECT refresh_token FROM tokens WHERE user_id = ? AND refresh_token = ?`)
	if err != nil {
		logger.Error("auth.ValidateRefreshToken: could not validate token for user: %d:, error: %s", userId, err)
		return errors.New("could not prepare token for validation against server")
	}
	rows, err := stmt.Query(userId, refreshToken)
	if err != nil {
		logger.Error("auth.ValidateRefreshToken: could not query for refresh token for user: %d, error: %s", userId, err)
		return errors.New("could not retrieve server refresh tokens for user")
	}
	defer rows.Close()
	if !rows.Next() {
		logger.Error("auth.ValidateRefreshToken: no tokens matched what was passed in for user: %d", userId)
		return errors.New("could not validate refresh token, please login again")
	}
	logger.Info("auth.ValidateRefreshToken: successfully matched refresh token sent with a token found in db for user: %d", userId)
	return nil
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

func (source *authSource) CleanupExpiredRefreshTokens(logger Logger) {
	logger.Info("auth.CleanupExpiredRefreshTokens: cleaning up tokens older than %d days", refreshTokenDayLimit)

	rs, err := source.db.Exec(cleanupExpiredRefreshTokensStatement)
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
