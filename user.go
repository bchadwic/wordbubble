package main

import (
	"time"
	"unicode"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
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
	ValidPassword(password string) bool
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
func (users *users) ValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
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
	return hasUpper && hasLower && hasNumber && hasSpecial
}
