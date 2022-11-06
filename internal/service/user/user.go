package user

import "github.com/bchadwic/wordbubble/model"

const (
	AddUser                = `INSERT INTO users(username, email, password) VALUES ($1, $2, $3) RETURNING user_id;`
	RetrieveUserByEmail    = `SELECT user_id, username, email, password FROM users WHERE email = $1`
	RetrieveUserByUsername = `SELECT user_id, username, email, password FROM users WHERE username = $1`
)

// UserService is the interface that the application
// uses to interact with user information
type UserService interface {
	// AddUser verifies uniqueness, then adds a new user to the database after hashing the password.
	// error can be (500) resp.ErrCouldNotBeHashPassword, (500) resp.ErrCouldNotAddUser,
	// (400) resp.ErrUserWithUsernameAlreadyExists, (400) resp.ErrCouldNotDetermineUserExistence,
	// (400) resp.ErrUserWithEmailAlreadyExists, (400) resp.ErrCouldNotDetermineUserExistence or nil.
	AddUser(user *model.User) error
	// RetrieveUnauthenticatedUser retrieve everything about a user by a user string (email or username), without a password.
	// *model.User is the unauthenticated user found, can be nil.
	// error can be (500) resp.ErrSQLMappingError, (400) resp.ErrUnknownUser, (400) resp.ErrCouldNotDetermineUserType or nil.
	RetrieveUnauthenticatedUser(userStr string) (*model.User, error)
	// RetrieveAuthenticatedUser retrieve everything about a user by user string (email or username), without a password
	// this func validates that the password matches what's in the database; an error is returned if it does not
	// *model.User is the authenticated user found, can be nil.
	// error can be (500) resp.ErrSQLMappingError, (400) resp.ErrUnknownUser,
	// (400) resp.ErrCouldNotDetermineUserType, (401) resp.ErrInvalidCredentials or nil.
	RetrieveAuthenticatedUser(userStr, password string) (*model.User, error)
}

// UserRepo is the interface that the service layer
// interacts with to access user information
type UserRepo interface {
	// addUser adds a new user to the database.
	// int64 is the user id of the user added, could be 0 for no user.
	// error can be (500) resp.ErrCouldNotAddUser or nil.
	addUser(user *model.User) (int64, error)
	// retrieveUserByEmail retrieves user details by email.
	// *model.User is the user retrieved from the email, could be nil.
	// error can be (400) resp.ErrUnknownUser, (500) resp.ErrSQLMappingError or nil.
	retrieveUserByEmail(email string) (*model.User, error)
	// retrieveUserByUsername retrieves user details by username.
	// *model.User is the user retrieved from the username, could be nil.
	// error can be (400) resp.ErrUnknownUser, (500) resp.ErrSQLMappingError or nil.
	retrieveUserByUsername(username string) (*model.User, error)
}
