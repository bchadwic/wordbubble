package user

import (
	"database/sql"
	"errors"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

type userRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewUserRepo(cfg cfg.Config) *userRepo {
	return &userRepo{
		log: cfg.NewLogger("users_repo"),
		db:  cfg.DB(),
	}
}

func (repo *userRepo) addUser(user *model.User) (int64, error) {
	row := repo.db.QueryRow(AddUser, user.Username, user.Email, user.Password)
	var lastInsertedId int64
	if err := row.Scan(&lastInsertedId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, resp.ErrCouldNotAddUser
		}
		return 0, resp.ErrSQLMappingError
	}
	return lastInsertedId, nil
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
