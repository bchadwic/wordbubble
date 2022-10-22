package resp

import (
	"net/http"
)

var (
	Unknown                           = []byte("sorry, it looks like an unknown error occurred")
	ErrNoWordbubble                   = NoContent("could not find a wordbubble for this user")
	ErrParseWordbubble                = BadRequest("could not parse wordbubble from request body")
	ErrParseUser                      = BadRequest("could not parse user from request body")
	ErrParseRefreshToken              = BadRequest("could not parse refresh token from request body")
	ErrCouldNotDetermineUserType      = BadRequest("could not determine if user passed is a username or an email")
	ErrUnknownUser                    = BadRequest("could not find user")
	ErrUserWithUsernameAlreadyExists  = BadRequest("a user already exists with this username")
	ErrUserWithEmailAlreadyExists     = BadRequest("a user already exists with this email")
	ErrEmailIsNotValid                = BadRequest("email in request is not a valid email")
	ErrEmailIsTooLong                 = BadRequest("no one should have an email this long")
	ErrUsernameIsTooLong              = BadRequest("no one should have a username this long")
	ErrUsernameIsNotLongEnough        = BadRequest("a username is required")
	ErrUsernameInvalidChars           = BadRequest("username must only consist of letters, numbers, or '_'")
	ErrUnauthorized                   = Unauthorized("bearer token authorization is required for this operation")
	ErrInvalidCredentials             = Unauthorized("could not authenticate using credentials passed")
	ErrCouldNotValidateRefreshToken   = Unauthorized("could not validate the refresh token, please login again")
	ErrRefreshTokenIsExpired          = Unauthorized("refresh token is expired, please login again")
	ErrInvalidTokenSignature          = Unauthorized("token signature was found to be invalid")
	ErrInvalidHttpMethod              = MethodNotAllowed("invalid http method")
	ErrMaxAmountOfWordbubblesReached  = Conflict("the max amount of wordbubbles has been created for this user")
	ErrCouldNotStoreRefreshToken      = InternalServerError("could not successfully store refresh token")
	ErrCouldNotDetermineUserExistence = InternalServerError("could not determine if user exists")
	ErrCouldNotBeHashPassword         = InternalServerError("an error occurred storing password")
	ErrCouldNotCleanupTokens          = InternalServerError("an error occurred cleaning up old refresh tokens")
	ErrCouldNotAddUser                = InternalServerError("an error occurred adding user to database")
	ErrSQLMappingError                = InternalServerError("an error occurred mapping data from the database")
)

type StatusNoContent struct {
	Code    int    `example:"201"`
	Message string `example:"could not find a wordbubble for this user"`
}

type StatusBadRequest struct {
	Code    int    `example:"400"`
	Message string `example:"could not determine if user passed is a username or an email"`
}

type StatusUnauthorized struct {
	Code    int    `example:"401"`
	Message string `example:"could not validate the refresh token, please login again"`
}

type StatusMethodNotAllowed struct {
	Code    int    `example:"405"`
	Message string `example:"invalid http method"`
}

type StatusConflict struct {
	Code    int    `example:"409"`
	Message string `example:"the max amount of wordbubbles has been created for this user"`
}

type StatusInternalServerError struct {
	Code    int    `example:"500"`
	Message string `example:"an error occurred mapping data from the database"`
}

func NoContent(message string) *StatusNoContent {
	return &StatusNoContent{http.StatusNoContent, message}
}

func BadRequest(message string) *StatusBadRequest {
	return &StatusBadRequest{http.StatusBadRequest, message}
}

func Unauthorized(message string) *StatusUnauthorized {
	return &StatusUnauthorized{http.StatusUnauthorized, message}
}

func MethodNotAllowed(message string) *StatusMethodNotAllowed {
	return &StatusMethodNotAllowed{http.StatusMethodNotAllowed, message}
}

func Conflict(message string) *StatusConflict {
	return &StatusConflict{http.StatusConflict, message}
}

func InternalServerError(message string) *StatusInternalServerError {
	return &StatusInternalServerError{http.StatusInternalServerError, message}
}

func (err *StatusNoContent) Error() string {
	return err.Message
}

func (err *StatusBadRequest) Error() string {
	return err.Message
}

func (err *StatusUnauthorized) Error() string {
	return err.Message
}

func (err *StatusMethodNotAllowed) Error() string {
	return err.Message
}

func (err *StatusConflict) Error() string {
	return err.Message
}

func (err *StatusInternalServerError) Error() string {
	return err.Message
}
