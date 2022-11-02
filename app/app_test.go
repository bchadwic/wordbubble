package app

import (
	"io"
	"net/http"
	"testing"

	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func (tcase *TestCase) HttpRequestTest(t *testing.T) {
	req, err := http.NewRequest(tcase.reqMethod, tcase.reqPath, tcase.reqBody)
	req.Header = tcase.reqHeader
	if err != nil {
		panic(err)
	}
	tcase.testApp.wordbubbles = tcase.wordbubbleService
	tcase.testApp.users = tcase.userService
	tcase.testApp.auth = tcase.authService
	w := &TestWriter{}
	tcase.operation(w, req)
	assert.Equal(t, tcase.respBody, w.respBody)
	assert.Equal(t, tcase.respStatusCode, w.statusCode)
}

func NewTestApp() *app {
	return &app{
		log: util.TestLogger(),
	}
}

type TestCase struct {
	testApp   *app
	operation func(w http.ResponseWriter, r *http.Request)

	reqPath   string
	reqBody   io.Reader
	reqMethod string
	reqHeader http.Header

	respBody       string
	respStatusCode int

	userService       *TestUserService
	wordbubbleService *TestWordbubbleService
	authService       *TestAuthService
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
	TokenIsNearEOL             bool
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
	token.NearEOL = tas.TokenIsNearEOL
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

type TestWordbubbleService struct {
	AddNewWordbubbleError                              error
	RemoveAndReturnLatestWordbubbleForUserIdWordbubble *resp.WordbubbleResponse
}

func (tws *TestWordbubbleService) AddNewWordbubble(userId int64, wb *req.WordbubbleRequest) error {
	return tws.AddNewWordbubbleError
}

func (tws *TestWordbubbleService) RemoveAndReturnLatestWordbubbleForUserId(userId int64) *resp.WordbubbleResponse {
	return tws.RemoveAndReturnLatestWordbubbleForUserIdWordbubble
}
