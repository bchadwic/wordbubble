package resp

import (
	"net/http"
)

type ErrorResponse struct {
	Message []byte
	Code    int
}

var (
	Unknown                          = []byte("sorry, it looks like an unknown error occurred")
	ErrInvalidMethod                 = NewErrorResp("invalid http method", http.StatusMethodNotAllowed)
	ErrUnauthorized                  = NewErrorResp("bearer token authorization is required for this operation", http.StatusUnauthorized)
	ErrParseWordBubble               = NewErrorResp("could not parse wordbubble from request body", http.StatusBadRequest)
	ErrParseUser                     = NewErrorResp("could not parse user from request body", http.StatusBadRequest)
	ErrParseRefreshToken             = NewErrorResp("could not parse refresh token from request body", http.StatusBadRequest)
	ErrUnknownUser                   = NewErrorResp("could not find user", http.StatusBadRequest)
	ErrInvalidCredentials            = NewErrorResp("could not authenticate using credentials passed", http.StatusUnauthorized)
	ErrNoWordBubble                  = NewErrorResp("could not find a wordbubble for this user", http.StatusNoContent)
	ErrMaxAmountOfWordBubblesReached = NewErrorResp("the max amount of wordbubbles has been created for this user", http.StatusConflict)
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
