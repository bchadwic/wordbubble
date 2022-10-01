package auth

import (
	"database/sql"
	"errors"

	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type authRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewAuthRepo(logger util.Logger, db *sql.DB) *authRepo {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			user_id INTEGER NOT NULL,  
			refresh_token TEXT NOT NULL,
			issued_at INTEGER NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(user_id)
		);
	`)
	return &authRepo{
		log: logger,
		db:  db,
	}
}

func (repo *authRepo) storeRefreshToken(token *refreshToken) error {
	_, err := repo.db.Exec(StoreRefreshToken, token.UserId(), token.string, token.issuedAt)
	if err != nil {
		repo.log.Error("could not execute statement for user: %d, error: %s", err)
		return errors.New("could not successfully store refresh token on server")
	}
	return nil
}

func (repo *authRepo) validateRefreshToken(token *refreshToken) error {
	row := repo.db.QueryRow(ValidateRefreshToken, token.UserId(), token.string)
	var issuedAt int64
	if err := row.Scan(&issuedAt); err != nil {
		repo.log.Error("could not map rows for user: %d, error: %s", token.UserId(), err)
		return errors.New("could not validate issued time of refresh token, please login again")
	}
	token.issuedAt = issuedAt
	return nil
}

func (repo *authRepo) getLatestRefreshToken(userId int64) *refreshToken {
	row := repo.db.QueryRow(GetLatestRefreshToken, userId)
	var issuedAt int64
	var val string
	if err := row.Scan(&val, &issuedAt); err != nil {
		repo.log.Error("could not map the latest refresh token for user: %d, error: %s", userId, err)
		return nil
	}
	return &refreshToken{string: val, issuedAt: issuedAt, userId: userId}
}

func (repo *authRepo) CleanupExpiredRefreshTokens(since int64) error {
	rs, err := repo.db.Exec(CleanupExpiredRefreshTokens, since)
	if err != nil {
		repo.log.Error("could not delete tokens: error: %s", err)
		return errors.New("an error occurred cleaning up expired refresh tokens")
	}
	amt, _ := rs.RowsAffected()
	repo.log.Info("refresh token cleaner deleted: %d tokens", amt)
	return nil
}
