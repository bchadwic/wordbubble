package user

import "github.com/bchadwic/wordbubble/model"

const (
	AddUser                = `INSERT INTO users(username, email, password) VALUES (?, ?, ?);`
	RetrieveUserByEmail    = `SELECT user_id, username, email, password FROM users WHERE email = ?`
	RetrieveUserByUsername = `SELECT user_id, username, email, password FROM users WHERE username = ?`
)

// UserService is the interface that the application
// uses to interact with user information
type UserService interface {
	// adds a new user
	AddUser(user *model.User) error
	// retrieve everything about a user by a user string (email or username), without a password
	RetrieveUnauthenticatedUser(userStr string) (*model.User, error)
	// retrieve everything about a user by user string (email or username), without a password
	// this func validates that the password matches what's in the database
	// an error is returned if it does not
	RetrieveAuthenticatedUser(userStr, password string) (*model.User, error)
}

// UserRepo is the interface that the service layer
// interacts with to access user information
type UserRepo interface {
	// adds a new user to the database
	addUser(user *model.User) (int64, error)
	// retrieve user details by email
	retrieveUserByEmail(email string) (*model.User, error)
	// retrieve user details by username
	retrieveUserByUsername(username string) (*model.User, error)
}
