package app

import (
	"net/http"

	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
)

func NewTestApp() *app {
	return &app{
		log: util.TestLogger(),
	}
}

type TestWriter struct {
	header     map[string][]string
	respBody   string
	writeErr   error
	statusCode int
}

func (tr *TestWriter) Header() http.Header {
	return tr.header
}

func (tr *TestWriter) Write(b []byte) (int, error) {
	tr.respBody = string(b)
	return len(b), tr.writeErr
}

func (tr *TestWriter) WriteHeader(statusCode int) {
	tr.statusCode = statusCode
}

type TestAuthService struct {
	GenerateAccessTokenString  string
	GenerateRefreshTokenString string
	GenerateRefreshTokenError  error
	ValidateRefreshTokenError  error
}

func (tas *TestAuthService) GenerateAccessToken(userId int64) string {
	return tas.GenerateAccessTokenString
}

func (tas *TestAuthService) GenerateRefreshToken(userId int64) (string, error) {
	return tas.GenerateRefreshTokenString, tas.GenerateRefreshTokenError
}

func (tas *TestAuthService) ValidateRefreshToken(token *auth.RefreshToken) error {
	return tas.ValidateRefreshTokenError
}

type TestUserService struct {
	AddUserError                     error
	RetrieveUnauthenticatedUserUser  *model.User
	RetrieveUnauthenticatedUserError error
	RetrieveAuthenticatedUserUser    *model.User
	RetrieveAuthenticatedUserError   error
}

func (tus *TestUserService) AddUser(user *model.User) error {
	return tus.AddUserError
}

func (tus *TestUserService) RetrieveUnauthenticatedUser(userStr string) (*model.User, error) {
	return tus.RetrieveUnauthenticatedUserUser, tus.RetrieveUnauthenticatedUserError
}

func (tus *TestUserService) RetrieveAuthenticatedUser(userStr, password string) (*model.User, error) {
	return tus.RetrieveAuthenticatedUserUser, tus.RetrieveAuthenticatedUserError
}
