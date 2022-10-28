package app

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func Test_Push(t *testing.T) {
	util.SigningKey = func() []byte {
		return []byte("test signing key")
	}
	tests := map[string]TestCase{
		"valid": {
			reqBody:           strings.NewReader(`{"text":"hello"}`),
			respBody:          fmt.Sprintln(`{"message":"thank you!"}`),
			respStatusCode:    http.StatusCreated,
			reqMethod:         http.MethodPost,
			reqHeader:         http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}},
			wordbubbleService: &TestWordbubbleService{},
		},
		"invalid, user maxed out the amount of wordbubbles": {
			reqBody:        strings.NewReader(`{"text":"hello"}`),
			respBody:       resp.ErrMaxAmountOfWordbubblesReached.Message,
			respStatusCode: resp.ErrMaxAmountOfWordbubblesReached.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}},
			wordbubbleService: &TestWordbubbleService{
				AddNewWordbubbleError: resp.ErrMaxAmountOfWordbubblesReached,
			},
		},
		"invalid, request body format": {
			reqBody:        strings.NewReader(`hello`),
			respBody:       resp.ErrParseWordbubble.Message,
			respStatusCode: resp.ErrParseWordbubble.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjkyMjMzNzIwMzY4NTQ3NzU4MDcsInVzZXJfaWQiOjJ9.PdKS3GZkc1LDMyJhZMP0CIkTUDUIXxcvQP0jwOkygO8"}},
		},
		"invalid, token is expired": {
			respBody:       resp.ErrTokenIsExpired.Message,
			respStatusCode: resp.ErrTokenIsExpired.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMCwidXNlcl9pZCI6Mn0.TQLM9cWE3eSFmvnD8ipFpRtXcozDMl3sTY_qekH24SI"}},
		},
		"invalid, bad bearer token": {
			respBody:       resp.ErrInvalidTokenSignature.Message,
			respStatusCode: resp.ErrInvalidTokenSignature.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"Bearer kladsjfkasjd;fkljasd"}},
		},
		"invalid, bad token": {
			respBody:       resp.ErrUnauthorized.Message,
			respStatusCode: resp.ErrUnauthorized.Code,
			reqMethod:      http.MethodPost,
			reqHeader:      http.Header{"Authorization": []string{"not a real token"}},
		},
		"invalid, no token": {
			respBody:       resp.ErrUnauthorized.Message,
			respStatusCode: resp.ErrUnauthorized.Code,
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
			req, err := http.NewRequest(tcase.reqMethod, "/v1/push", tcase.reqBody)
			req.Header = tcase.reqHeader
			if err != nil {
				panic(err)
			}
			testApp := NewTestApp()
			testApp.users = tcase.userService
			testApp.wordbubbles = tcase.wordbubbleService
			w := &TestWriter{}
			testApp.Push(w, req)
			assert.Equal(t, tcase.respBody, w.respBody)
			assert.Equal(t, tcase.respStatusCode, w.statusCode)
		})
	}
}
