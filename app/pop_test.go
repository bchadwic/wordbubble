package app

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
)

func Test_Pop(t *testing.T) {
	tests := map[string]TestCase{
		"valid": {
			reqBody:        strings.NewReader(`{"user":"ben"}`),
			respBody:       fmt.Sprintln(`{"text":"hello world"}`),
			respStatusCode: http.StatusOK,
			reqMethod:      http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserUser: &model.User{},
			},
			wordbubbleService: &TestWordbubbleService{
				RemoveAndReturnLatestWordbubbleForUserIdWordbubble: &resp.WordbubbleResponse{
					Text: "hello world",
				},
			},
		},
		"invalid, no wordbubble found": {
			reqBody:        strings.NewReader(`{"user":"ben"}`),
			respBody:       resp.ErrNoWordbubble.Message,
			respStatusCode: resp.ErrNoWordbubble.Code,
			reqMethod:      http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserUser: &model.User{},
			},
			wordbubbleService: &TestWordbubbleService{},
		},
		"invalid, couldn't find the user": {
			reqBody:        strings.NewReader(`{"user":"ben"}`),
			respBody:       resp.ErrUnknownUser.Message,
			respStatusCode: resp.ErrUnknownUser.Code,
			reqMethod:      http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserError: resp.ErrUnknownUser,
			},
		},
		"invalid, no user": {
			reqBody:        strings.NewReader(`{}`),
			respBody:       resp.ErrNoUser.Message,
			respStatusCode: resp.ErrNoUser.Code,
			reqMethod:      http.MethodDelete,
		},
		"invalid, could not parse body": {
			reqBody:        strings.NewReader(`what's goin' on here?`),
			respBody:       resp.ErrParseUser.Message,
			respStatusCode: resp.ErrParseUser.Code,
			reqMethod:      http.MethodDelete,
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
			tcase.operation = tcase.testApp.Pop
			tcase.HttpRequestTest(t)
		})
	}
}
