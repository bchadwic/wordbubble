package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
)

func Test_Login(t *testing.T) {
	tests := map[string]TestCase{
		"valid": {
			reqBody:        `{"user":"ben","password":"SomePassword123"}`,
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
			reqBody:        `{"user":"ben","password":"SomePassword123"}`,
			respBody:       structToJson(resp.ErrCouldNotStoreRefreshToken),
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
			reqBody:        `{"user":"*234olj2kx.s","password":"SomePassword123"}`,
			respBody:       structToJson(resp.ErrCouldNotDetermineUserType),
			respStatusCode: resp.ErrCouldNotDetermineUserType.Code,
			reqMethod:      http.MethodPost,
			userService: &TestUserService{
				RetrieveAuthenticatedUserError: resp.ErrCouldNotDetermineUserType,
			},
		},
		"invalid, missing password": {
			reqBody:        `{"user":"ben"}`,
			respBody:       structToJson(resp.ErrNoPassword),
			respStatusCode: resp.ErrNoPassword.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, missing user": {
			reqBody:        `{"password":"SomePassword123"}`,
			respBody:       structToJson(resp.ErrNoUser),
			respStatusCode: resp.ErrNoUser.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, bad body": {
			reqBody:        `howdy!`,
			respBody:       structToJson(resp.ErrParseUser),
			respStatusCode: resp.ErrParseUser.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, GET http method": {
			respBody:       structToJson(resp.ErrInvalidHttpMethod),
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
