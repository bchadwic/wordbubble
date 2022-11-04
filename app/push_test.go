package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

func Test_Push(t *testing.T) {
	util.SigningKey = func() []byte {
		return []byte("test signing key")
	}
	tests := map[string]TestCase{
		"valid": {
			reqBody:           `{"text":"hello"}`,
			respBody:          fmt.Sprintln(`{"message":"thank you!"}`),
			respStatusCode:    http.StatusCreated,
			reqMethod:         http.MethodPost,
			reqHeader:         http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}},
			wordbubbleService: &TestWordbubbleService{},
		},
		"invalid, user maxed out the amount of wordbubbles": {
			reqBody:        `{"text":"hello"}`,
			respBody:       structToJson(resp.ErrMaxAmountOfWordbubblesReached),
			respStatusCode: resp.ErrMaxAmountOfWordbubblesReached.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}},
			wordbubbleService: &TestWordbubbleService{
				AddNewWordbubbleError: resp.ErrMaxAmountOfWordbubblesReached,
			},
		},
		"invalid, request body format": {
			reqBody:        `hello`,
			respBody:       structToJson(resp.ErrParseWordbubble),
			respStatusCode: resp.ErrParseWordbubble.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}},
		},
		"invalid, token is expired": {
			respBody:       structToJson(resp.ErrTokenIsExpired),
			respStatusCode: resp.ErrTokenIsExpired.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMCwidXNlcl9pZCI6Mn0.TQLM9cWE3eSFmvnD8ipFpRtXcozDMl3sTY_qekH24SI"}},
		},
		"invalid, bad bearer token": {
			respBody:       structToJson(resp.ErrInvalidTokenSignature),
			respStatusCode: resp.ErrInvalidTokenSignature.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer kladsjfkasjd;fkljasd"}},
		},
		"invalid, bad token": {
			respBody:       structToJson(resp.ErrUnauthorized),
			respStatusCode: resp.ErrUnauthorized.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"not a real token"}},
		},
		"invalid, no token": {
			respBody:       structToJson(resp.ErrUnauthorized),
			respStatusCode: resp.ErrUnauthorized.Code,
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
			tcase.operation = tcase.testApp.Push
			tcase.HttpRequestTest(t)
		})
	}
}
