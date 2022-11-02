package app

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
)

func Test_Login(t *testing.T) {
	tests := map[string]TestCase{
		"valid": {
			reqBody:        strings.NewReader(`{"user":"ben","password":"SomePassword123"}`),
			respBody:       fmt.Sprintln(`{"access_token":"test.access.token","refresh_token":"test.refresh.token"}`),
			respStatusCode: http.StatusOK,
			reqMethod:      http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserUser: &model.User{},
			},
			authService: &TestAuthService{
				GenerateRefreshTokenString: "test.refresh.token",
				GenerateAccessTokenString:  "test.access.token",
			},
		},
		"invalid, auth service couldn't store a refresh token ": {
			reqBody:        strings.NewReader(`{"user":"ben","password":"SomePassword123"}`),
			respBody:       resp.ErrCouldNotStoreRefreshToken.Message,
			respStatusCode: resp.ErrCouldNotStoreRefreshToken.Code,
			reqMethod:      http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserUser: &model.User{},
			},
			authService: &TestAuthService{
				GenerateRefreshTokenError: resp.ErrCouldNotStoreRefreshToken,
			},
		},
		"invalid, user service couldn't find user": {
			reqBody:        strings.NewReader(`{"user":"*234olj2kx.s","password":"SomePassword123"}`),
			respBody:       resp.ErrCouldNotDetermineUserType.Message,
			respStatusCode: resp.ErrCouldNotDetermineUserType.Code,
			reqMethod:      http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserError: resp.ErrCouldNotDetermineUserType,
			},
		},
		"invalid, missing password": {
			reqBody:        strings.NewReader(`{"user":"ben"}`),
			respBody:       resp.ErrNoPassword.Message,
			respStatusCode: resp.ErrNoPassword.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, missing user": {
			reqBody:        strings.NewReader(`{"password":"SomePassword123"}`),
			respBody:       resp.ErrNoUser.Message,
			respStatusCode: resp.ErrNoUser.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, bad body": {
			reqBody:        strings.NewReader(`howdy!`),
			respBody:       resp.ErrParseUser.Message,
			respStatusCode: resp.ErrParseUser.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, GET http method": {
			respBody:       resp.ErrInvalidHttpMethod.Message,
			respStatusCode: resp.ErrInvalidHttpMethod.Code,
			reqMethod:      http.MethodGet,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			tcase.testApp = NewTestApp()
			tcase.operation = tcase.testApp.Login
			tcase.HttpRequestTest(t)
		})
	}
}
