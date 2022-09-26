package user

import "github.com/bchadwic/wordbubble/model"

const (
	AddUser                = `INSERT INTO users(username, email, password) VALUES (?, ?, ?);`
	RetrieveUserByEmail    = `SELECT user_id, username, email, password FROM users WHERE email = ?`
	RetrieveUserByUsername = `SELECT user_id, username, email, password FROM users WHERE username = ?`
)

type UserService interface {
	// add a new user
	AddUser(user *model.User) error
	// retrieve everything about a user, except sensitive info, using a string that could be a username or an email
	RetrieveUnauthenticatedUser(userStr string) (*model.User, error)
	// retrieve everything about a user using by a string that could be a username or an email and the user's unencrypted password
	RetrieveAuthenticatedUser(userStr, password string) (*model.User, error)
}

type UserRepo interface {
	AddUser(user *model.User) (int64, error)
	RetrieveUserByEmail(userStr string) (*model.User, error)
	RetrieveUserByUsername(userStr string) (*model.User, error)
}
