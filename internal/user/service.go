package user

import (
	"fmt"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	// add a new user
	AddUser(user *model.User) error
	// retrieve everything about a user, except sensitive info, using a string that could be a username or an email
	RetrieveUserByString(userStr string) *model.User
	// retrieve everything about a user using by a string that could be a username or an email and the user's unencrypted password
	RetrieveAuthenticatedUserByString(userStr, password string) *model.User
}

type userService struct {
	repo   UserRepo
	logger util.Logger
}

func NewUserService(source UserRepo, logger util.Logger) *userService {
	return &userService{source, logger}
}

func (svc *userService) AddUser(user *model.User) error {
	// super inefficient to do two calls into the database to check existence, then another to insert,
	// but this doesn't get called often. Might come back
	if svc.repo.RetrieveUserByString(user.Email) != nil {
		return fmt.Errorf("user with the email %s already exists", user.Email)
	}
	if svc.repo.RetrieveUserByString(user.Username) != nil {
		return fmt.Errorf("%s already exists", user.Username)
	}
	var passwordBytes = []byte(user.Password)
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		svc.logger.Error("bcrypt error, could not add user %s", err)
		return err // bcrypt err'd out, can't continue
	}
	user.Password = string(hashedPasswordBytes)
	id, err := svc.repo.AddUser(user)
	if err != nil {
		svc.logger.Error("could not add user %s", err)
		return err
	}
	user.Id = id
	return nil
}

func (svc *userService) RetrieveUserByString(userStr string) *model.User {
	user := svc.repo.RetrieveUserByString(userStr)
	user.Password = ""
	return user
}

func (svc *userService) RetrieveAuthenticatedUserByString(userStr, password string) *model.User {
	user := svc.repo.RetrieveUserByString(userStr)
	if user == nil {
		svc.logger.Error("couldn't retrieve user by string, user: %s", userStr)
		return nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		svc.logger.Error("password did not match hashed password %s", err)
		return nil // db password and the password passed did not match
	}
	return user // successfully authenticated
}
