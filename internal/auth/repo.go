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

func NewAuthRepo(logger util.Logger) *authRepo {
	panicker := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	db, err := sql.Open("sqlite3", "./wordbubble.db")
	panicker(err)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tokens (
		user_id INTEGER NOT NULL,  
		refresh_token TEXT NOT NULL,
		issued_at INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`)
	panicker(err)
	return &authRepo{db, logger}
}

func (repo *authRepo) StoreRefreshToken(token *refreshToken) error {
	stmt, err := repo.db.Prepare(`INSERT INTO tokens (user_id, refresh_token, issued_at) VALUES (?, ?, ?)`)
	if err != nil {
		repo.log.Error("could not store token for user: %d, error: %s", token.UserId(), err)
		return errors.New("could not successfully prepare refresh token for storage on server")
	}
	_, err = stmt.Exec(token.UserId(), token.string, token.issuedAt)
	if err != nil {
		repo.log.Error("could not execute statement for user: %d, error: %s", err)
		return errors.New("could not successfully store refresh token on server")
	}
	return nil
}

func (repo *authRepo) ValidateRefreshToken(token *refreshToken) error {
	stmt, err := repo.db.Prepare(`SELECT issued_at FROM tokens WHERE user_id = ? AND refresh_token = ?`)
	if err != nil {
		repo.log.Error("could not validate token for user: %d:, error: %s", token.UserId(), err)
		return errors.New("could not prepare token for validation against server")
	}
	row, err := stmt.Query(token.UserId(), token.string)
	if err != nil {
		repo.log.Error("could not query for refresh token for user: %d, error: %s", token.UserId(), err)
		return errors.New("could not retrieve server refresh tokens for user")
	}
	defer row.Close()
	if !row.Next() {
		repo.log.Error("no tokens matched what was passed in for user: %d", token.UserId())
		return errors.New("could not validate refresh token, please login again")
	}
	var issuedAt int64
	if err := row.Scan(&issuedAt); err != nil {
		repo.log.Error("could not map rows for user: %d, error: %s", token.UserId(), err)
		return errors.New("could not validate issued time of refresh token, please login again")
	}
	token.issuedAt = issuedAt
	return nil
}

func (repo *authRepo) GetLatestRefreshToken(userId int64) *refreshToken {
	stmt, err := repo.db.Prepare(`SELECT refresh_token, issued_at FROM tokens WHERE user_id = ? ORDER BY issued_at DESC LIMIT 1`)
	if err != nil {
		repo.log.Error("could not get latest refresh token for user: %d:, error: %s", userId, err)
		return nil
	}
	row, err := stmt.Query(userId)
	if err != nil {
		repo.log.Error("could not query for refresh token for user: %d, error: %s", userId, err)
		return nil
	}
	defer row.Close()
	if !row.Next() {
		repo.log.Error("no refresh tokens were found for user: %d", userId)
		return nil
	}
	var issuedAt int64
	var val string
	if err := row.Scan(&val, &issuedAt); err != nil {
		repo.log.Error("could not map the latest refresh token for user: %d, error: %s", userId, err)
		return nil
	}
	return &refreshToken{
		string: val, issuedAt: issuedAt, userId: userId}
}

type AuthCleaner interface {
	CleanupExpiredRefreshTokens(since int64)
}

func (repo *authRepo) CleanupExpiredRefreshTokens(since int64) {
	stmt, err := repo.db.Prepare(cleanupExpiredRefreshTokensStatement)
	if err != nil {
		repo.log.Error("could not execute delete tokens: error: %s", err)
		return
	}
	rs, err := stmt.Exec(since)
	if err != nil {
		repo.log.Error("could not execute delete tokens: error: %s", err)
		return
	}
	amt, err := rs.RowsAffected()
	if err != nil {
		repo.log.Error("could not successfully determine the amount of tokens that were deleted, error: %s", err)
		return
	}
	repo.log.Info("refresh token cleaner deleted: %d tokens", amt)
}
