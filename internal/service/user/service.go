package user

import (
	"errors"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo UserRepo
	log  util.Logger
}

func NewUserService(cfg cfg.Config, repo UserRepo) *userService {
	return &userService{
		log:  cfg.NewLogger("users"),
		repo: repo,
	}
}

func (svc *userService) AddUser(user *model.User) error {
	if err := svc.verifyUserUniqueness(user); err != nil {
		return err
	}
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return resp.ErrCouldNotBeHashPassword
	}
	user.Password = string(hashedPasswordBytes)
	id, err := svc.repo.addUser(user)
	if err != nil {
		return err
	}
	user.Id = id
	return nil
}

func (svc *userService) RetrieveUnauthenticatedUser(userStr string) (*model.User, error) {
	user, err := svc.retrieveUserByString(userStr)
	if err != nil {
		return nil, err
	}
	user.Password = "" // sanitize
	return user, nil
}

func (svc *userService) RetrieveAuthenticatedUser(userStr, password string) (*model.User, error) {
	user, err := svc.retrieveUserByString(userStr)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, resp.ErrInvalidCredentials
	}
	user.Password = "" // no reason to pass the password around
	return user, nil   // successfully authenticated
}

// verify the uniqueness of a user against the database
// soon to be deprecated. A stored procedure should be made to
// give a code on which uniqueness constraint has been violated
func (svc *userService) verifyUserUniqueness(uniqueUser *model.User) error {
	var exists bool
	user, err := svc.repo.retrieveUserByUsername(uniqueUser.Username)
	if exists = user != nil; exists || !errors.Is(err, resp.ErrUnknownUser) {
		if exists {
			return resp.ErrUserWithUsernameAlreadyExists
		} // if the error from repo is not a mapping error, we can't determine if the user exists
		return resp.ErrCouldNotDetermineUserExistence
	}
	user, err = svc.repo.retrieveUserByEmail(uniqueUser.Email)
	if exists = user != nil; exists || !errors.Is(err, resp.ErrUnknownUser) {
		if exists {
			return resp.ErrUserWithEmailAlreadyExists
		}
		return resp.ErrCouldNotDetermineUserExistence
	}
	return nil
}

func (svc *userService) retrieveUserByString(userStr string) (*model.User, error) {
	switch {
	case util.ValidEmail(userStr) == nil:
		return svc.repo.retrieveUserByEmail(userStr)
	case util.ValidUsername(userStr) == nil:
		return svc.repo.retrieveUserByUsername(userStr)
	default:
		return nil, resp.ErrCouldNotDetermineUserType
	}
}
