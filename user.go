package main

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int64
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Users interface {
	// add a new user
	AddUser(user *User) error
	// retrieve everything about a user, except sensitive info, using a string that could be a username or an email
	RetrieveUserByString(userStr string) *User
	// retrieve everything about a user using by a string that could be a username or an email and the user's unencrypted password
	RetrieveAuthenticatedUserByString(userStr, password string) *User
}

type users struct {
	source DataSource
	logger Logger
}

func NewUsersService(source DataSource, logger Logger) *users {
	return &users{source, logger}
}

func (users *users) AddUser(user *User) error {


	// need to check if this user already exists before inserting



	var passwordBytes = []byte(user.Password)
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		users.logger.Error("bcrypt error, could not add user %s", err)
		return err // bcrypt err'd out, can't continue
	}
	user.Password = string(hashedPasswordBytes)
	id, err := users.source.AddUser(user)
	if err != nil {
		users.logger.Error("could not add user %s", err)
		return err
	}
	user.Id = id
	return nil
}

func (users *users) RetrieveUserByString(userStr string) *User {
	user := users.source.RetrieveUserByString(userStr)
	user.Password = ""
	return user
}

func (users *users) RetrieveAuthenticatedUserByString(userStr, password string) *User {
	user := users.source.RetrieveUserByString(userStr)
	if user == nil {
		users.logger.Error("couldn't retrieve user by string, user: %s", userStr)
		return nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		users.logger.Error("password did not match hashed password %s", err)
		return nil // db password and the password passed did not match
	}
	return user // successfully authenticated
}
