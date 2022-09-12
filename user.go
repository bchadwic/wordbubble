package main

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int64
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Users interface {
	AddUser(user *User) error
	AuthenticateUser(user *User) error
	// gets a user from database from either a username or an email
	GetUserFromUserString(userStr string) *User
	// validate user based on whether the user exists with either the valid username or valid email
	ValidUser(user *User) error
}

type users struct {
	source DataSource
	logger Logger
}

func NewUsersService(source DataSource, logger Logger) *users {
	return &users{source, logger}
}

func (users *users) AddUser(user *User) error {
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

func (users *users) AuthenticateUser(user *User) error {
	dbUser, err := users.source.GetAuthenticatedUserFromUsername(user)
	if err != nil {
		users.logger.Error("could not retrieve user from database %s", err)
		return err // could not find the user by username
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		users.logger.Error("password did not match hashed password %s", err)
		return err // db password and the password passed did not match
	}
	user.Id = dbUser.Id
	return nil // successfully authenticated
}

func (users *users) GetUserFromUserString(userStr string) *User {
	if err := ValidEmail(userStr); err == nil {
		user, _ := users.source.GetUserFromEmail(userStr)
		return user
	}
	if err := ValidUsername(userStr); err == nil {
		user, _ := users.source.GetUserFromUsername(userStr)
		return user
	}
	return nil
}

func (users *users) ValidUser(user *User) error {
	username, email := user.Username, user.Email

	if err := ValidEmail(email); err != nil {
		return err
	}

	if err := ValidUsername(username); err != nil {
		return err
	}

	// lookups
	if _, err := users.source.GetUserFromEmail(email); err == nil {
		return errors.New("a user already exists with this email")
	}
	if _, err := users.source.GetUserFromUsername(username); err == nil {
		return fmt.Errorf("the user '%s' already exists", username)
	}
	return nil
}
