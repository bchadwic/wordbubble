package resp

import (
	"net/http"
)

type ErrorResponse struct {
	Message []byte
	Code    int
}

var (
	// Unknown Error
	Unknown = []byte("sorry, it looks like an unknown error occurred")

	// 204
	ErrNoWordBubble = NewErrorResp("could not find a wordbubble for this user", http.StatusNoContent)

	// 400 - Bad Request
	ErrParseWordBubble               = NewErrorResp("could not parse wordbubble from request body", http.StatusBadRequest)
	ErrParseUser                     = NewErrorResp("could not parse user from request body", http.StatusBadRequest)
	ErrParseRefreshToken             = NewErrorResp("could not parse refresh token from request body", http.StatusBadRequest)
	ErrCouldNotDetermineUserType     = NewErrorResp("could not determine if user passed is a username or an email", http.StatusBadRequest)
	ErrUnknownUser                   = NewErrorResp("could not find user", http.StatusBadRequest)
	ErrUserWithUsernameAlreadyExists = NewErrorResp("a user already exists with this username", http.StatusBadRequest)
	ErrUserWithEmailAlreadyExists    = NewErrorResp("a user already exists with this email", http.StatusBadRequest)
	ErrEmailIsNotValid               = NewErrorResp("email in request is not a valid email", http.StatusBadRequest)
	ErrEmailIsTooLong                = NewErrorResp("no one should have an email this long", http.StatusBadRequest)
	ErrUsernameIsTooLong             = NewErrorResp("no one should have a username this long", http.StatusBadRequest)
	ErrUsernameIsNotLongEnough       = NewErrorResp("a username is required", http.StatusBadRequest)
	ErrUsernameInvalidChars          = NewErrorResp("username must only consist of letters, numbers, or '_'", http.StatusBadRequest)

	// 401 - Unauthorized
	ErrUnauthorized                 = NewErrorResp("bearer token authorization is required for this operation", http.StatusUnauthorized)
	ErrInvalidCredentials           = NewErrorResp("could not authenticate using credentials passed", http.StatusUnauthorized)
	ErrCouldNotValidateRefreshToken = NewErrorResp("could not validate the refresh token, please login again", http.StatusUnauthorized)
	ErrRefreshTokenIsExpired        = NewErrorResp("refresh token is expired, please login again", http.StatusUnauthorized)
	ErrInvalidTokenSignature        = NewErrorResp("token signature was found to be invalid", http.StatusUnauthorized)

	// 405 - Method Not Allowed
	ErrInvalidHttpMethod = NewErrorResp("invalid http method", http.StatusMethodNotAllowed)

	// 409 - Conflict
	ErrMaxAmountOfWordBubblesReached = NewErrorResp("the max amount of wordbubbles has been created for this user", http.StatusConflict)

	// 500 - InternalServerError
	ErrCouldNotStoreRefreshToken      = NewErrorResp("could not successfully store refresh token", http.StatusInternalServerError)
	ErrCouldNotDetermineUserExistence = NewErrorResp("could not determine if user exists", http.StatusInternalServerError)
	ErrCouldNotBeHashPassword         = NewErrorResp("an error occurred storing password", http.StatusInternalServerError)
	ErrCouldNotCleanupTokens          = NewErrorResp("an error occurred cleaning up old refresh tokens", http.StatusInternalServerError)
	ErrCouldNotAddUser                = NewErrorResp("an error occurred adding user to database", http.StatusInternalServerError)
	ErrSQLMappingError                = NewErrorResp("an error occurred mapping data from the database", http.StatusInternalServerError)
)

func NewErrorResp(message string, statusCode int) *ErrorResponse {
	return &ErrorResponse{
		Message: []byte(message),
		Code:    statusCode,
	}
}

func (err *ErrorResponse) Error() string {
	return string(err.Message)
}
