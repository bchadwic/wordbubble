package main

import (
	"errors"
	"net/mail"
	"unicode"
)

const (
	minPasswordLength = 6
	maxUsernameLength = 40
	maxEmailLength    = 320
)

func ValidUser(user *User) error {
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
		return errors.New("email in request is not a valid email")
	}
	if len(email) > maxEmailLength {
		return errors.New("no one should have an email this long")
	}
	return nil
}

// validate username, shorter than maxUsernameLength, longer than "", only letters, numbers or '_'s
func ValidUsername(username string) error {
	if len(username) > maxUsernameLength {
		return errors.New("no one should have a username this long")
	} else if len(username) == 0 {
		return errors.New("a username is required")
	}
	for _, c := range username {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_' {
			continue
		}
		return errors.New("username must only consist of letters, numbers, or '_'")
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
		return errors.New(errStr + last)
	}
	return errors.New(errStr + "and " + last)
}
