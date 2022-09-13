package user

import (
	"database/sql"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type UserRepo interface {
	AddUser(user *model.User) (int64, error)
	RetrieveUserByString(userStr string) *model.User
}

type userRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewUserRepo(logger util.Logger) *userRepo {
	panicker := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	db, err := sql.Open("sqlite3", "./wb-database.db")
	panicker(err)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`)
	panicker(err)
	return &userRepo{db, logger}
}

func (repo *userRepo) AddUser(user *model.User) (int64, error) {
	stmt, err := repo.db.Prepare(`INSERT INTO users(username, email, password) VALUES (?, ?, ?);`)
	if err != nil {
		repo.log.Error("prepared statement error for adding user: %s, error: %s", user.Username, err)
		return 0, err
	}
	res, err := stmt.Exec(user.Username, user.Email, user.Password)
	if err != nil {
		repo.log.Error("executing error for adding user: %s, error: %s", user.Username, err)
		return 0, err
	}
	return res.LastInsertId()
}

func (repo *userRepo) RetrieveUserByString(userStr string) *model.User {
	// TODO log the error once we get it back
	var stmt *sql.Stmt
	var err error
	switch {
	case util.ValidEmail(userStr) == nil:
		stmt, err = repo.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE email = ?`)
	case util.ValidUsername(userStr) == nil:
		stmt, err = repo.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE username = ?`)
	default:
		repo.log.Info("couldn't determine if the string passed is a username or an email")
		return nil
	}
	if err != nil {
		repo.log.Error("prepared statement error for retrieving user by string, user: %s, error: %s", userStr, err)
		return nil
	}
	row, err := stmt.Query(userStr)
	if err != nil {
		repo.log.Error("querying error for retrieving user by string: %s, error: %s", userStr, err)
		return nil
	}
	defer row.Close()
	if !row.Next() {
		repo.log.Info("could not find user in database, user: %s", userStr)
		return nil
	}
	var dbUser model.User
	if err := row.Scan(&dbUser.Id, &dbUser.Username, &dbUser.Email, &dbUser.Password); err != nil {
		repo.log.Error("could not map db user to user struct, user: %s, error: %s", userStr, err)
		return nil
	}
	return &dbUser
}
