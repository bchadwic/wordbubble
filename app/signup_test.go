package app

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model/resp"
)

func Test_ValidSignup(t *testing.T) {
	tests := map[string]TestCase{
		"valid": {
			reqBody:        strings.NewReader(`{"email":"benchadwick87@gmail.com","username":"user1","password":"Password123!"}`),
			respBody:       fmt.Sprintln(`{"access_token":"some.access.token","refresh_token":"some.refresh.token"}`),
			respStatusCode: http.StatusCreated,
			reqMethod:      http.MethodPost,
			userService:    &TestUserService{},
			authService: &TestAuthService{
				GenerateRefreshTokenString: "some.refresh.token",
				GenerateAccessTokenString:  "some.access.token",
			},
		},
		"invalid, error generating the refresh token": {
			reqBody:        strings.NewReader(`{"email":"benchadwick87@gmail.com","username":"user1","password":"Password123!"}`),
			respBody:       resp.ErrCouldNotStoreRefreshToken.Message,
			respStatusCode: resp.ErrCouldNotStoreRefreshToken.Code,
			reqMethod:      http.MethodPost,
			userService:    &TestUserService{},
			authService: &TestAuthService{
				GenerateRefreshTokenError: resp.ErrCouldNotStoreRefreshToken,
			},
		},
		"invalid, error adding user": {
			reqBody:        strings.NewReader(`{"email":"benchadwick87@gmail.com","username":"user1","password":"Password123!"}`),
			respBody:       resp.ErrCouldNotAddUser.Message,
			respStatusCode: resp.ErrCouldNotAddUser.Code,
			reqMethod:      http.MethodPost,
			userService: &TestUserService{
				AddUserError: resp.ErrCouldNotAddUser,
			},
		},
		"invalid, no password": {
			reqBody:        strings.NewReader(`{"email":"benchadwick87@gmail.com","username":"user1"}`),
			respBody:       `password must contain at least 6 characters, one uppercase character, one lowercase character, and one number`,
			respStatusCode: http.StatusBadRequest,
			reqMethod:      http.MethodPost,
		},
		"invalid, no username": {
			reqBody:        strings.NewReader(`{"email":"benchadwick87@gmail.com","password":"Password123!"}`), // no body
			respBody:       resp.ErrUsernameIsMissing.Message,
			respStatusCode: resp.ErrUsernameIsMissing.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, no email": {
			reqBody:        strings.NewReader(`{"username":"ben","password":"Password123!"}`),
			respBody:       resp.ErrEmailIsNotValid.Message,
			respStatusCode: resp.ErrEmailIsNotValid.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, no body": {
			reqBody:        strings.NewReader(``),
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
			tcase.operation = tcase.testApp.Signup
			tcase.HttpRequestTest(t)
		})
	}
}
