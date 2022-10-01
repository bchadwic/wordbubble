package auth

import (
	"database/sql"

	"github.com/bchadwic/wordbubble/resp"
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
		return resp.ErrCouldNotStoreRefreshToken
	}
	return nil
}

func (repo *authRepo) validateRefreshToken(token *refreshToken) error {
	row := repo.db.QueryRow(ValidateRefreshToken, token.UserId(), token.string)
	var issuedAt int64
	if err := row.Scan(&issuedAt); err != nil {
		return resp.ErrCouldNotValidateRefreshToken
	}
	token.issuedAt = issuedAt
	return nil
}

func (repo *authRepo) getLatestRefreshToken(userId int64) *refreshToken {
	row := repo.db.QueryRow(GetLatestRefreshToken, userId)
	var issuedAt int64
	var val string
	if err := row.Scan(&val, &issuedAt); err != nil {
		return nil
	}
	return &refreshToken{
		string:   val,
		issuedAt: issuedAt,
		userId:   userId,
	}
}

func (repo *authRepo) CleanupExpiredRefreshTokens(since int64) error {
	rs, err := repo.db.Exec(CleanupExpiredRefreshTokens, since)
	if err != nil {
		return resp.ErrCouldNotCleanupTokens
	}
	amt, _ := rs.RowsAffected()
	repo.log.Info("refresh token cleaner deleted: %d tokens", amt)
	return nil
}
