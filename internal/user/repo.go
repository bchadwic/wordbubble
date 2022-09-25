package user

import (
	"database/sql"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type userRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewUserRepo(logger util.Logger, db *sql.DB) *userRepo {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
	`)
	return &userRepo{db, logger}
}

func (repo *userRepo) AddUser(user *model.User) (int64, error) {
	res, err := repo.db.Exec(AddUser, user.Username, user.Email, user.Password)
	if err != nil {
		repo.log.Error("executing error for adding user: %s, error: %s", user.Username, err)
		return 0, err
	}
	return res.LastInsertId()
}

func (repo *userRepo) RetrieveUserByString(userStr string) *model.User {
	var row *sql.Row
	switch {
	case util.ValidEmail(userStr) == nil:
		row = repo.db.QueryRow(RetrieveUserByEmail, userStr)
	case util.ValidUsername(userStr) == nil:
		row = repo.db.QueryRow(RetrieveUserByUsername, userStr)
	default:
		repo.log.Info("couldn't determine if the string passed is a username or an email")
		return nil
	}
	var dbUser model.User
	if err := row.Scan(&dbUser.Id, &dbUser.Username, &dbUser.Email, &dbUser.Password); err != nil {
		repo.log.Error("could not map db user to user struct, user: %s, error: %s", userStr, err)
		return nil
	}
	return &dbUser
}
