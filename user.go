package main

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserId   int64
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Users interface {
	AddUser(user *User) error
	AuthenticateUser(user *User) error
	ResolveUserIdFromValue(userStr string) (int64, error)
	// validate password based on the 6 characters, 1 upper, 1 lower, 1 number, 1 special character
	// error is safe to return to consumer as a response message
	ValidPassword(password string) error
	// validate user based on whether the user exists with either the username or email
	// also be sure to check that they are both valid inputs
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
	user.UserId = id
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
	user.UserId = dbUser.UserId
	return nil // successfully authenticated
}

// resolves a userId from either a username or an email
func (users *users) ResolveUserIdFromValue(userStr string) (int64, error) {
	if strings.ContainsRune(userStr, '@') {
		// TODO validate that this is a valid email before reaching out to datasource
		return users.source.ResolveUserIdFromEmail(userStr)
	}
	// TODO validate that this is a valid username before reaching out to datasource
	return users.source.ResolveUserIdFromUsername(userStr)
}

func (users *users) ValidPassword(password string) error {
	var hasMinLen, hasUpper, hasLower, hasNumber, hasSpecial bool
	if len(password) > minPasswordLength {
		hasMinLen = true
	}
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	if hasUpper && hasLower && hasNumber && hasSpecial {
		return nil
	}
	errStr := "password must contain "
	count := 0
	last := ""
	incrementAndAppendLast := func() {
		count++
		if last != "" {
			errStr += last + ", "
		}
	}
	if !hasMinLen {
		incrementAndAppendLast()
		last = "at least 6 characters"
	}
	if !hasUpper {
		incrementAndAppendLast()
		last = "one uppercase character"
	}
	if !hasLower {
		incrementAndAppendLast()
		last = "one lowercase character"
	}
	if !hasNumber {
		incrementAndAppendLast()
		last = "one number"
	}
	if !hasSpecial {
		incrementAndAppendLast()
		last = "one special character"
	}
	if count == 1 {
		return fmt.Errorf(errStr + last)
	}
	return fmt.Errorf(errStr + "and " + last)
}

func (users *users) ValidUser(user *User) error {
	username, email := user.Username, user.Email

	// validation
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("email passed is not a valid email")
	}
	if len(email) > maxEmailLength {
		return fmt.Errorf("no one should have an email this long")
	}
	if len(username) > maxUsernameLength {
		return fmt.Errorf("no one needs a username this long")
	} else if len(username) == 0 {
		return fmt.Errorf("a username is required")
	}
	for _, c := range username {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_' {
			continue
		}
		return fmt.Errorf("username must only consist of letters, numbers, or '_'")
	}

	// lookups
	if _, err := users.source.GetUserFromEmail(email); err == nil {
		return fmt.Errorf("a user already exists with this email")
	}
	if _, err := users.source.GetUserFromUsername(username); err == nil {
		return fmt.Errorf("the user '%s' already exists", username)
	}
	return nil
}
