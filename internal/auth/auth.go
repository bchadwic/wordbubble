package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/bchadwic/wordbubble/util"
	"github.com/golang-jwt/jwt"
)

const (
	refreshTokenTimeLimit    = 60
	accessTokenTimeLimit     = 30 * time.Second
	RefreshTokenCleanerRate  = 30 * time.Second
	ImminentExpirationWindow = int64(float64(refreshTokenTimeLimit) * .2)
)

var cleanupExpiredRefreshTokensStatement = `DELETE FROM tokens WHERE issued_at < ?`

type Auth interface {
	GenerateAccessToken(userId int64) (string, error)
	GenerateRefreshToken(userId int64) (string, error)
	GetUserIdFromTokenString(tokenStr string) (int64, error)
	VerifyTokenAgainstAuthSource(userId int64, tokenStr string) (int64, error)
	GetOrCreateLatestRefreshToken(userId int64) string
}

type AuthSource interface {
	StoreRefreshToken(userId int64, refreshToken string) error
	ValidateRefreshToken(userId int64, refreshToken string) (int64, error)
	RemoveRefreshToken(userId int64, refreshToken string) error
	GetLatestRefreshToken(userId int64) *refreshToken
	CleanupExpiredRefreshTokens()
}
type auth struct {
	source     AuthSource
	log        util.Logger
	signingKey string
}
type authSource struct {
	db  *sql.DB
	log util.Logger
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}

type token string

type refreshToken struct {
	string
	issuedAt int64
}

func NewAuthSource(logger util.Logger) *authSource {
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
	return &authSource{db, logger}
}

func NewAuth(source AuthSource, logger util.Logger, signingKey string) *auth {
	return &auth{source, logger, signingKey}
}

// TODO combine GenerateAccessToken and GenerateRefreshToken?
func (auth *auth) GenerateAccessToken(userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(accessTokenTimeLimit).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})
	signedToken, err := token.SignedString([]byte(auth.signingKey))
	if err != nil {
		auth.log.Error("failed to create access token for user: %d, error: %s", userId, err)
		return "", errors.New("failed to sign and generate an access token")
	}
	return signedToken, nil
}

func (auth *auth) GenerateRefreshToken(userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{},
		userId,
	})
	signedToken, err := token.SignedString([]byte(auth.signingKey))
	if err != nil {
		auth.log.Error("failed to create access token for user: %d, error: %s", userId, err)
		return "", errors.New("failed to sign and generate a refresh token")
	}
	if err := auth.source.StoreRefreshToken(userId, signedToken); err != nil {
		return "", err
	}
	return signedToken, nil
}

func (auth *auth) GetUserIdFromTokenString(tokenStr string) (int64, error) {
	tokenClaims := &tokenClaims{} // TODO come back to this mapping
	token, err := jwt.ParseWithClaims(tokenStr, tokenClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(auth.signingKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			auth.log.Error("signature is invalid, error: %s", err)
			return 0, errors.New("token signature was found to be invalid")
		}
		auth.log.Error("an error occurred while parsing token, defaulting to expiration: error %s", err)
		return 0, errors.New("access token is expired")
	}
	if !token.Valid { // only applicable to access tokens
		auth.log.Error("token is expired for user: %d, error: %s", tokenClaims.UserId, err)
		return 0, errors.New("access token is expired")
	}
	return tokenClaims.UserId, nil
}

func (auth *auth) VerifyTokenAgainstAuthSource(userId int64, tokenStr string) (int64, error) {
	issuedAt, err := auth.source.ValidateRefreshToken(userId, tokenStr)
	if err != nil {
		return 0, err
	}
	timeDiff := time.Now().Unix() - issuedAt
	if timeDiff >= refreshTokenTimeLimit {
		auth.log.Error("token was found to be expired for user: %d", userId)
		return 0, errors.New("refresh token is expired, please login again")
	}
	return refreshTokenTimeLimit - timeDiff, nil
}

func (auth *auth) GetOrCreateLatestRefreshToken(userId int64) string {
	token := auth.source.GetLatestRefreshToken(userId)
	if token != nil { // if there is a token that isn't close to dying, use that
		if timeRemaining := time.Now().Unix() - token.issuedAt; timeRemaining > ImminentExpirationWindow {
			return token.string
		}
	} // otherwise create a new one
	newRefreshToken, _ := auth.GenerateRefreshToken(userId)
	return newRefreshToken
}

func (source *authSource) StoreRefreshToken(userId int64, refreshToken string) error {
	stmt, err := source.db.Prepare(`INSERT INTO tokens (user_id, refresh_token, issued_at) VALUES (?, ?, ?)`)
	if err != nil {
		source.log.Error("could not store token for user: %d, error: %s", userId, err)
		return errors.New("could not successfully prepare refresh token for storage on server")
	}
	_, err = stmt.Exec(userId, refreshToken, time.Now().Unix())
	if err != nil {
		source.log.Error("could not execute statement for user: %d, error: %s", err)
		return errors.New("could not successfully store refresh token on server")
	}
	return nil
}

func (source *authSource) ValidateRefreshToken(userId int64, refreshToken string) (int64, error) {
	stmt, err := source.db.Prepare(`SELECT issued_at FROM tokens WHERE user_id = ? AND refresh_token = ?`)
	if err != nil {
		source.log.Error("could not validate token for user: %d:, error: %s", userId, err)
		return 0, errors.New("could not prepare token for validation against server")
	}
	row, err := stmt.Query(userId, refreshToken)
	if err != nil {
		source.log.Error("could not query for refresh token for user: %d, error: %s", userId, err)
		return 0, errors.New("could not retrieve server refresh tokens for user")
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("no tokens matched what was passed in for user: %d", userId)
		return 0, errors.New("could not validate refresh token, please login again")
	}
	var issuedAt int64
	if err := row.Scan(&issuedAt); err != nil {
		source.log.Error("could not map rows for user: %d, error: %s", userId, err)
		return 0, errors.New("could not validate issued time of refresh token, please login again")
	}
	return issuedAt, nil
}

func (source *authSource) RemoveRefreshToken(userId int64, refreshToken string) error {
	stmt, err := source.db.Prepare(`DELETE FROM tokens WHERE user_id = ? AND refresh_token = ?`)
	if err != nil {
		source.log.Error("could not remove token for user: %d:, error: %s", userId, err)
		return errors.New("could not prepare task to remove existing refresh token")
	}
	rs, err := stmt.Exec(userId, refreshToken)
	if err != nil {
		source.log.Error("could not execute refresh token deletion for user: %d, error: %s", userId, err)
		return errors.New("could not execute the removal of existing refresh token")
	}
	amt, err := rs.RowsAffected()
	if err != nil {
		source.log.Error("could not successfully determine if the token was deleted for user: %d, error: %s", userId, err)
		return errors.New("could not determine if refresh token was deleted successfully")
	}
	if amt < 1 {
		source.log.Error("no records were deleted %d", userId)
		return errors.New("could not successfully delete refresh token")
	}
	return nil
}

func (source *authSource) CleanupExpiredRefreshTokens() {
	stmt, err := source.db.Prepare(cleanupExpiredRefreshTokensStatement)
	if err != nil {
		source.log.Error("could not execute delete tokens: error: %s", err)
		return
	}
	rs, err := stmt.Exec(time.Now().Unix() - refreshTokenTimeLimit)
	if err != nil {
		source.log.Error("could not execute delete tokens: error: %s", err)
		return
	}
	amt, err := rs.RowsAffected()
	if err != nil {
		source.log.Error("could not successfully determine the amount of tokens that were deleted, error: %s", err)
		return
	}
	source.log.Info("expired refresh token cleaner interval: %gs, deleted: %d tokens", RefreshTokenCleanerRate.Seconds(), amt)
}

func (source *authSource) GetLatestRefreshToken(userId int64) *refreshToken {
	stmt, err := source.db.Prepare(`SELECT refresh_token, issued_at FROM tokens WHERE user_id = ? ORDER BY issued_at DESC LIMIT 1`)
	if err != nil {
		source.log.Error("could not get latest refresh token for user: %d:, error: %s", userId, err)
		return nil
	}
	row, err := stmt.Query(userId)
	if err != nil {
		source.log.Error("could not query for refresh token for user: %d, error: %s", userId, err)
		return nil
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("no refresh tokens were found for user: %d", userId)
		return nil
	}
	var issuedAt int64
	var val string
	if err := row.Scan(&val, &issuedAt); err != nil {
		source.log.Error("could not map the latest refresh token for user: %d, error: %s", userId, err)
		return nil
	}
	return &refreshToken{
		val,
		issuedAt,
	}
}
