package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/stretchr/testify/assert"
)

func Test_Login(t *testing.T) {
	tests := map[string]struct {
		body             io.Reader
		method           string
		wantedBody       string
		wantedStatusCode int
		authService      *TestAuthService
		userService      *TestUserService
	}{
		"valid": {
			body:             strings.NewReader(`{"user":"ben","password":"SomePassword123"}`),
			wantedBody:       fmt.Sprintln(`{"access_token":"test.access.token","refresh_token":"test.refresh.token"}`),
			wantedStatusCode: http.StatusOK,
			method:           http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserUser: &model.User{},
			},
			authService: &TestAuthService{
				GenerateRefreshTokenString: "test.refresh.token",
				GenerateAccessTokenString:  "test.access.token",
			},
		},
		"invalid, auth service couldn't store a refresh token ": {
			body:             strings.NewReader(`{"user":"ben","password":"SomePassword123"}`),
			wantedBody:       resp.ErrCouldNotStoreRefreshToken.Message,
			wantedStatusCode: resp.ErrCouldNotStoreRefreshToken.Code,
			method:           http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserUser: &model.User{},
			},
			authService: &TestAuthService{
				GenerateRefreshTokenError: resp.ErrCouldNotStoreRefreshToken,
			},
		},
		"invalid, user service couldn't find user": {
			body:             strings.NewReader(`{"user":"*234olj2kx.s","password":"SomePassword123"}`),
			wantedBody:       resp.ErrCouldNotDetermineUserType.Message,
			wantedStatusCode: resp.ErrCouldNotDetermineUserType.Code,
			method:           http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserError: resp.ErrCouldNotDetermineUserType,
			},
		},
		"invalid, missing password": {
			body:             strings.NewReader(`{"user":"ben"}`),
			wantedBody:       resp.ErrNoPassword.Message,
			wantedStatusCode: resp.ErrNoPassword.Code,
			method:           http.MethodPost,
		},
		"invalid, missing user": {
			body:             strings.NewReader(`{"password":"SomePassword123"}`),
			wantedBody:       resp.ErrNoUser.Message,
			wantedStatusCode: resp.ErrNoUser.Code,
			method:           http.MethodPost,
		},
		"invalid, bad body": {
			body:             strings.NewReader(`howdy!`),
			wantedBody:       resp.ErrParseUser.Message,
			wantedStatusCode: resp.ErrParseUser.Code,
			method:           http.MethodPost,
		},
		"invalid, GET http method": {
			wantedBody:       resp.ErrInvalidHttpMethod.Message,
			wantedStatusCode: resp.ErrInvalidHttpMethod.Code,
			method:           http.MethodGet,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			req, err := http.NewRequest(tcase.method, "/login", tcase.body)
			if err != nil {
				panic(err)
			}
			testApp := NewTestApp()
			testApp.auth = tcase.authService
			testApp.users = tcase.userService
			w := &TestWriter{}
			testApp.Login(w, req)
			assert.Equal(t, tcase.wantedBody, w.respBody)
			assert.Equal(t, tcase.wantedStatusCode, w.statusCode)
		})
	}
}
