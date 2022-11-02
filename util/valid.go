package util

import (
	"fmt"
	"net/mail"
	"unicode"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
)

const (
	minPasswordLength   = 6
	maxUsernameLength   = 40
	maxEmailLength      = 100
	MinWordbubbleLength = 1
	MaxWordbubbleLength = 255
)

// validate all the fields of a user
func ValidUser(user *model.User) error {
	if err := ValidEmail(user.Email); err != nil {
		return err
	}
	if err := ValidUsername(user.Username); err != nil {
		return err
	}
	if err := ValidPassword(user.Password); err != nil {
		return err
	}
	return nil
}

// ValidEmail validates that the string passed in is an email, and that it's not longer than maxEmailLength
func ValidEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return resp.ErrEmailIsNotValid
	}
	if len(email) > maxEmailLength {
		return resp.ErrEmailIsTooLong
	}
	return nil
}

// ValidUsername validates username, shorter than maxUsernameLength, longer than "", only letters, numbers or '_'s
func ValidUsername(username string) error {
	if len(username) > maxUsernameLength {
		return resp.ErrUsernameIsTooLong
	} else if len(username) == 0 {
		return resp.ErrUsernameIsMissing
	}
	for _, c := range username {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_' {
			continue
		}
		return resp.ErrUsernameInvalidChars
	}
	return nil
}

// ValidPassword validate password based on the 6 characters, 1 upper, 1 lower, 1 number, 1 special character
func ValidPassword(password string) error {
	var hasMinLen, hasUpper, hasLower, hasNumber bool
	if len(password) > minPasswordLength {
		hasMinLen = true
	}
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		}
	}
	if hasUpper && hasLower && hasNumber {
		return nil
	}
	errStr := "password must contain "
	count := 0
	last := ""
	incrementAndAppendLast := func() {
		count++
		if last != "" {
			errStr += last + ", "
		}
	}
	if !hasMinLen {
		incrementAndAppendLast()
		last = fmt.Sprintf("at least %d characters", minPasswordLength)
	}
	if !hasUpper {
		incrementAndAppendLast()
		last = "one uppercase character"
	}
	if !hasLower {
		incrementAndAppendLast()
		last = "one lowercase character"
	}
	if !hasNumber {
		incrementAndAppendLast()
		last = "one number"
	}
	if count == 1 {
		return resp.BadRequest(errStr + last)
	} // InvalidPassword
	return resp.BadRequest(errStr + "and " + last)
}

// ValidWordbubble validates a wordbubble making sure it meets size constraints
func ValidWordbubble(wb *req.WordbubbleRequest) error {
	len := len(wb.Text)
	if len < MinWordbubbleLength || len > MaxWordbubbleLength {
		return resp.BadRequest( // InvalidWordbubble
			fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", MinWordbubbleLength, MaxWordbubbleLength, len),
		)
	}
	return nil
}
