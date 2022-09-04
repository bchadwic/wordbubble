package main

import (
	"fmt"
	"net/mail"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 6
	maxUsernameLength = 40
	maxEmailLength    = 320
)

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Users interface {
	AddUser(logger Logger, user *User) error
	GenerateToken(logger Logger, user *User) (string, error)
	AuthenticateUser(logger Logger, user *User) bool
	ValidPassword(logger Logger, password string) error
	ValidUser(logger Logger, user *User) error
}

type users struct {
	db         DB
	signingKey string
}

type tokenClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Email    string `json:"email"`
}

func NewUsersService(signingKey string) *users {
	return &users{
		db:         NewDB(),
		signingKey: signingKey,
	}
}

func (users *users) AddUser(logger Logger, user *User) error {
	logger.Info("users.AddUser: inserting new user %s", user.Username)

	var passwordBytes = []byte(user.Password)
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		logger.Error("users.AddUser: bcrypt error, could not add user %s", err)
		return err // bcrypt err'd out, can't continue
	}

	user.Password = string(hashedPasswordBytes)
	logger.Info("users.AddUser: successfully hashed password")
	return users.db.AddUser(logger, user)
}

func (users *users) GenerateToken(logger Logger, user *User) (string, error) {
	logger.Info("users.GenerateToken: generating token for %s", user.Username)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.Username,
		user.Email,
	}) // constructing payload of the jwt token before signing

	logger.Info("users.GenerateToken: successfully generated token for %s", user.Username)
	return token.SignedString([]byte(users.signingKey))
}

func (users *users) AuthenticateUser(logger Logger, user *User) bool {
	logger.Info("users.AuthenticateUser: verifying %s login credentials", user.Username)

	dbUser, err := users.db.GetUserFromUsername(logger, user.Username)
	if err != nil {
		logger.Error("users.AuthenticateUser: could not retrieve user from database %s", err)
		return false // could not find the user by username
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		logger.Error("users.AuthenticateUser: password did not match hashed password %s", err)
		return false // db password and the password passed did not match
	}

	logger.Info("users.AuthenticateUser: user %s is verified to be who they say they are", user.Username)
	return true // successfully authenticated
}

// validate password based on the 6 characters, 1 upper, 1 lower, 1 number, 1 special character
// error is safe to return to consumer as a response message
func (users *users) ValidPassword(logger Logger, password string) error {
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

// validate user based on whether the user exists with either the username or email
// also be sure to check that they are both valid inputs
// error is safe to return to consumer as a response message
func (users *users) ValidUser(logger Logger, user *User) error {
	username, email := user.Username, user.Email
	logger.Info("username is %s", username)

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

	// lookups
	if _, err := users.db.GetUserFromEmail(logger, email); err == nil {
		return fmt.Errorf("a user already exists with this email")
	}
	if _, err := users.db.GetUserFromUsername(logger, username); err == nil {
		return fmt.Errorf("the user '%s' already exists", username)
	}
	return nil
}
