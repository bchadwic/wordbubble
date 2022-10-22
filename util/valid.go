package util

import (
	"fmt"
	"net/mail"
	"unicode"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
)

const (
	minPasswordLength   = 6
	maxUsernameLength   = 40
	maxEmailLength      = 320
	MinWordBubbleLength = 1
	MaxWordBubbleLength = 255
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

// validate that the string passed in is an email, and that it's not longer than maxEmailLength
func ValidEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return resp.ErrEmailIsNotValid
	}
	if len(email) > maxEmailLength {
		return resp.ErrEmailIsTooLong
	}
	return nil
}

// validate username, shorter than maxUsernameLength, longer than "", only letters, numbers or '_'s
func ValidUsername(username string) error {
	if len(username) > maxUsernameLength {
		return resp.ErrUsernameIsTooLong
	} else if len(username) == 0 {
		return resp.ErrUsernameIsNotLongEnough
	}
	for _, c := range username {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_' {
			continue
		}
		return resp.ErrUsernameInvalidChars
	}
	return nil
}

// validate password based on the 6 characters, 1 upper, 1 lower, 1 number, 1 special character
func ValidPassword(password string) error {
	var hasMinLen, hasUpper, hasLower, hasNumber, hasSpecial bool
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
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	if hasUpper && hasLower && hasNumber && hasSpecial {
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
		last = "at least 6 characters"
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
	if !hasSpecial {
		incrementAndAppendLast()
		last = "one special character"
	}
	if count == 1 {
		return resp.BadRequest(errStr + last)
	} // InvalidPassword
	return resp.BadRequest(errStr + "and" + last)
}

func ValidWordBubble(wb *model.WordBubble) error {
	len := len(wb.Text)
	if len < MinWordBubbleLength || len > MaxWordBubbleLength {
		return resp.BadRequest( // InvalidWordBubble
			fmt.Sprintf("wordbubble sent is invalid, must be inbetween %d-%d characters, received a length of %d", MinWordBubbleLength, MaxWordBubbleLength, len),
		)
	}
	return nil
}
