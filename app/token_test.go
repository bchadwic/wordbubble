package app

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

func Test_Token(t *testing.T) {
	util.SigningKey = func() []byte {
		return []byte("test signing key")
	}
	tests := map[string]TestCase{
		"valid, refresh token is near end of life": {
			reqBody:        strings.NewReader(`{"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}`),
			respBody:       fmt.Sprintln(`{"access_token":"aaa.bbb.ccc","refresh_token":"ddd.eee.fff"}`),
			respStatusCode: http.StatusOK,
			reqMethod:      http.MethodPost,
			authService: &TestAuthService{
				GenerateAccessTokenString:  "aaa.bbb.ccc",
				GenerateRefreshTokenString: "ddd.eee.fff",
				TokenIsNearEOL:             true,
			},
		},
		"valid, refresh token is not at end of life": {
			reqBody:        strings.NewReader(`{"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}`),
			respBody:       fmt.Sprintln(`{"access_token":"aaa.bbb.ccc"}`),
			respStatusCode: http.StatusOK,
			reqMethod:      http.MethodPost,
			authService: &TestAuthService{
				GenerateAccessTokenString: "aaa.bbb.ccc",
			},
		},
		"invalid, database couldn't validate the token": {
			reqBody:        strings.NewReader(`{"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}`),
			respBody:       resp.ErrCouldNotValidateRefreshToken.Message,
			respStatusCode: resp.ErrCouldNotValidateRefreshToken.Code,
			reqMethod:      http.MethodPost,
			authService: &TestAuthService{
				ValidateRefreshTokenError: resp.ErrCouldNotValidateRefreshToken,
			},
		},
		"invalid, empty token": {
			reqBody:        strings.NewReader(`{"refresh_token":""}`),
			respBody:       resp.ErrParseRefreshToken.Message,
			respStatusCode: resp.ErrParseRefreshToken.Code,
			reqMethod:      http.MethodPost,
		},
		"invalid, no body": {
			reqBody:        strings.NewReader(``),
			respBody:       resp.ErrParseRefreshToken.Message,
			respStatusCode: resp.ErrParseRefreshToken.Code,
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
			tcase.operation = tcase.testApp.Token
			tcase.HttpRequestTest(t)

		})
	}
}
