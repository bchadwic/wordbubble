package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bchadwic/wordbubble/model/resp"
)

func Test_ValidSignup(t *testing.T) {
	tests := map[string]TestCase{
		"valid": {
			reqBody:        `{"email":"benchadwick87@gmail.com","username":"user1","password":"Password123!"}`,
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
			reqBody:        `{"email":"benchadwick87@gmail.com","username":"user1","password":"Password123!"}`,
			respBody:       structToJson(resp.ErrCouldNotStoreRefreshToken),
			respStatusCode: resp.ErrCouldNotStoreRefreshToken.Code,
			reqMethod:      http.MethodPost,
			userService:    &TestUserService{},
			authService: &TestAuthService{
				GenerateRefreshTokenError: resp.ErrCouldNotStoreRefreshToken,
			},
		},
		"invalid, error adding user": {
			reqBody:        `{"email":"benchadwick87@gmail.com","username":"user1","password":"Password123!"}`,
			respBody:       structToJson(resp.ErrCouldNotAddUser),
			respStatusCode: resp.ErrCouldNotAddUser.Code,
			reqMethod:      http.MethodPost,
			userService: &TestUserService{
				AddUserError: resp.ErrCouldNotAddUser,
			},
		},
		"invalid, no password": {
			reqBody:        `{"email":"benchadwick87@gmail.com","username":"user1"}`,
			respBody:       structToJson(resp.BadRequest(`password must contain at least 6 characters, one uppercase character, one lowercase character, and one number`)),
			respStatusCode: http.StatusBadRequest,
			reqMethod:      http.MethodPost,
		},
		"invalid, no username": {
			reqBody:        `{"email":"benchadwick87@gmail.com","password":"Password123!"}`,
			respBody:       structToJson(resp.ErrUsernameIsMissing),
			respStatusCode: resp.ErrUsernameIsMissing.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, no email": {
			reqBody:        `{"username":"ben","password":"Password123!"}`,
			respBody:       structToJson(resp.ErrEmailIsNotValid),
			respStatusCode: resp.ErrEmailIsNotValid.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, no body": {
			reqBody:        ``,
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
			tcase.operation = tcase.testApp.Signup
			tcase.HttpRequestTest(t)
		})
	}
}
