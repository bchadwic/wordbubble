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

func Test_Pop(t *testing.T) {
	tests := map[string]struct {
		body              io.Reader
		method            string
		wantedBody        string
		wantedStatusCode  int
		userService       *TestUserService
		wordbubbleService *TestWordbubbleService
	}{
		"valid": {
			body:             strings.NewReader(`{"user":"ben"}`),
			wantedBody:       fmt.Sprintln(`{"text":"hello world"}`),
			wantedStatusCode: http.StatusOK,
			method:           http.MethodDelete,
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
			body:             strings.NewReader(`{"user":"ben"}`),
			wantedBody:       resp.ErrNoWordbubble.Message,
			wantedStatusCode: resp.ErrNoWordbubble.Code,
			method:           http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserUser: &model.User{},
			},
			wordbubbleService: &TestWordbubbleService{},
		},
		"invalid, couldn't find the user": {
			body:             strings.NewReader(`{"user":"ben"}`),
			wantedBody:       resp.ErrUnknownUser.Message,
			wantedStatusCode: resp.ErrUnknownUser.Code,
			method:           http.MethodDelete,
			userService: &TestUserService{
				RetrieveUnauthenticatedUserError: resp.ErrUnknownUser,
			},
		},
		"invalid, no user": {
			body:             strings.NewReader(`{}`),
			wantedBody:       resp.ErrNoUser.Message,
			wantedStatusCode: resp.ErrNoUser.Code,
			method:           http.MethodDelete,
		},
		"invalid, could not parse body": {
			body:             strings.NewReader(`what's goin' on here?`),
			wantedBody:       resp.ErrParseUser.Message,
			wantedStatusCode: resp.ErrParseUser.Code,
			method:           http.MethodDelete,
		},
		"invalid, GET http method": {
			wantedBody:       resp.ErrInvalidHttpMethod.Message,
			wantedStatusCode: resp.ErrInvalidHttpMethod.Code,
			method:           http.MethodGet,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			req, err := http.NewRequest(tcase.method, "/pop", tcase.body)
			if err != nil {
				panic(err)
			}
			testApp := NewTestApp()
			testApp.users = tcase.userService
			testApp.wordbubbles = tcase.wordbubbleService
			w := &TestWriter{}
			testApp.Pop(w, req)
			assert.Equal(t, tcase.wantedBody, w.respBody)
			assert.Equal(t, tcase.wantedStatusCode, w.statusCode)
		})
	}
}
