package util

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/stretchr/testify/assert"
)

func Test_ValidEmail(t *testing.T) {
	tests := map[string]struct {
		email       string
		expectedErr error
	}{
		"valid": {
			email: "benchadwick87@gmail.com",
		},
		"invalid, no domain ext": {
			email:       "benchadwick87@",
			expectedErr: resp.ErrEmailIsNotValid,
		},
		"invalid, no username": {
			email:       "@gmail.com",
			expectedErr: resp.ErrEmailIsNotValid,
		},
		"invalid, too long": {
			email:       strings.Repeat("a", maxEmailLength) + "@gmail.com",
			expectedErr: resp.ErrEmailIsTooLong,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			err := ValidEmail(tcase.email)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), tcase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_ValidUsername(t *testing.T) {
	tests := map[string]struct {
		username    string
		expectedErr error
	}{
		"valid": {
			username: "ben",
		},
		"valid, complex": {
			username: "a_1",
		},
		"invalid, too long": {
			username:    strings.Repeat("a", maxUsernameLength+1),
			expectedErr: resp.ErrUsernameIsTooLong,
		},
		"invalid, not long enough": {
			username:    "",
			expectedErr: resp.ErrUsernameIsMissing,
		},
		"invalid, spaces": {
			username:    "b e n",
			expectedErr: resp.ErrUsernameInvalidChars,
		},
		"invalid, special characters": {
			username:    "*",
			expectedErr: resp.ErrUsernameInvalidChars,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			err := ValidUsername(tcase.username)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), tcase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_ValidPassword(t *testing.T) {
	tests := map[string]struct {
		password    string
		expectedErr error
	}{
		"valid": {
			password: "Password123" + strings.Repeat("a", minPasswordLength),
		},
		"invalid, empty": {
			password:    "",
			expectedErr: errors.New("password must contain at least 6 characters, one uppercase character, one lowercase character, and one number"),
		},
		"invalid, has lowercase but not long enough": {
			password:    strings.Repeat("a", minPasswordLength),
			expectedErr: fmt.Errorf("password must contain at least %d characters, one uppercase character, and one number", minPasswordLength),
		},
		"invalid, has lowercase is long enough, but no uppercase or number": {
			password:    strings.Repeat("a", minPasswordLength+1),
			expectedErr: errors.New("password must contain one uppercase character, and one number"),
		},
		"invalid, has uppercase is long enough, but no lowercase or number": {
			password:    strings.Repeat("A", minPasswordLength+1),
			expectedErr: errors.New("password must contain one lowercase character, and one number"),
		},
		"invalid, has uppercase but not long enough, but no lowercase or number": {
			password:    strings.Repeat("A", minPasswordLength),
			expectedErr: fmt.Errorf("password must contain at least %d characters, one lowercase character, and one number", minPasswordLength),
		},
		"invalid, numbers only": {
			password:    strings.Repeat("1", minPasswordLength+1),
			expectedErr: errors.New("password must contain one uppercase character, and one lowercase character"),
		},
		"invalid, everything but a number": {
			password:    "Password" + strings.Repeat("a", minPasswordLength),
			expectedErr: errors.New("password must contain one number"),
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			err := ValidPassword(tcase.password)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), tcase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_ValidWordbubble(t *testing.T) {
	tests := map[string]struct {
		wordbubble  *req.WordbubbleRequest
		expectedErr error
	}{
		"valid": {
			wordbubble: &req.WordbubbleRequest{
				Text: "hi",
			},
		},
		"invalid, empty": {
			wordbubble: &req.WordbubbleRequest{
				Text: "",
			},
			expectedErr: resp.BadRequest(
				fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", MinWordbubbleLength, MaxWordbubbleLength, 0),
			),
		},
		"invalid, too long": {
			wordbubble: &req.WordbubbleRequest{
				Text: strings.Repeat("a", MaxWordbubbleLength+1),
			},
			expectedErr: resp.BadRequest(
				fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", MinWordbubbleLength, MaxWordbubbleLength, MaxWordbubbleLength+1),
			),
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			err := ValidWordbubble(tcase.wordbubble)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), tcase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_ValidUser(t *testing.T) {
	tests := map[string]struct {
		user        *model.User
		expectedErr error
	}{
		"valid": {
			user: &model.User{
				Username: "ben",
				Email:    "benchadwick87@gmail.com",
				Password: "Password123" + strings.Repeat("a", minPasswordLength),
			},
		},
		"invalid, invalid username": {
			user: &model.User{
				Username: "ben*",
				Email:    "benchadwick87@gmail.com",
				Password: "Password123" + strings.Repeat("a", minPasswordLength),
			},
			expectedErr: resp.ErrUsernameInvalidChars,
		},
		"invalid, invalid email": {
			user: &model.User{
				Username: "ben",
				Email:    "benchadwick87@gmail.com" + strings.Repeat("a", maxEmailLength),
				Password: "Password123" + strings.Repeat("a", minPasswordLength),
			},
			expectedErr: resp.ErrEmailIsTooLong,
		},
		"invalid, invalid password": {
			user: &model.User{
				Username: "ben",
				Email:    "benchadwick87@gmail.com",
				Password: "Password" + strings.Repeat("a", minPasswordLength),
			},
			expectedErr: errors.New("password must contain one number"),
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			err := ValidUser(tcase.user)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), tcase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
