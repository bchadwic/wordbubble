package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
)

func Test_Pop(t *testing.T) {
	tests := map[string]TestCase{
		"valid": {
			reqBody:        `{"user":"ben"}`,
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
			reqBody:        `{"user":"ben"}`,
			respBody:       structToJson(resp.ErrNoWordbubble),
			respStatusCode: resp.ErrNoWordbubble.Code,
			reqMethod:      http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserUser: &model.User{},
			},
			wordbubbleService: &TestWordbubbleService{},
		},
		"invalid, couldn't find the user": {
			reqBody:        `{"user":"ben"}`,
			respBody:       structToJson(resp.ErrUnknownUser),
			respStatusCode: resp.ErrUnknownUser.Code,
			reqMethod:      http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserError: resp.ErrUnknownUser,
			},
		},
		"invalid, no user": {
			reqBody:        `{}`,
			respBody:       structToJson(resp.ErrNoUser),
			respStatusCode: resp.ErrNoUser.Code,
			reqMethod:      http.MethodDelete,
		},
		"invalid, could not parse body": {
			reqBody:        `what's goin' on here?`,
			respBody:       structToJson(resp.ErrParseUser),
			respStatusCode: resp.ErrParseUser.Code,
			reqMethod:      http.MethodDelete,
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
			tcase.operation = tcase.testApp.Pop
			tcase.HttpRequestTest(t)
		})
	}
}
