package user

import (
	"database/sql"
	"errors"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
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
	return &userRepo{
		db:  db,
		log: logger,
	}
}

func (repo *userRepo) addUser(user *model.User) (int64, error) {
	res, err := repo.db.Exec(AddUser, user.Username, user.Email, user.Password)
	if err != nil {
		return 0, resp.ErrCouldNotAddUser
	}
	return res.LastInsertId() // sqlite3 supports last id, error is nil
}

func (repo *userRepo) retrieveUserByEmail(email string) (*model.User, error) {
	return repo.mapUserRow(repo.db.QueryRow(RetrieveUserByEmail, email))
}

func (repo *userRepo) retrieveUserByUsername(username string) (*model.User, error) {
	return repo.mapUserRow(repo.db.QueryRow(RetrieveUserByUsername, username))
}

func (repo *userRepo) mapUserRow(row *sql.Row) (*model.User, error) {
	var dbUser model.User
	if err := row.Scan(&dbUser.Id, &dbUser.Username, &dbUser.Email, &dbUser.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, resp.ErrUnknownUser
		}
		return nil, resp.ErrSQLMappingError
	}
	return &dbUser, nil
}
