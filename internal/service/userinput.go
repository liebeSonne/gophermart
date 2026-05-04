package service

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var ErrInvalidUserLogin = errors.New("invalid user login")
var ErrInvalidUserPassword = errors.New("invalid user password")

const MaxUserLoginLength = 255
const MaxUserPasswordLength = 255

type CreateUserInput struct {
	Login    string
	Password string `json:"-"`
}

type LoginUserInput struct {
	Login    string
	Password string `json:"-"`
}

func (i *CreateUserInput) Validate() error {
	if i.Login == "" {
		return fmt.Errorf("empty user login: %w", ErrInvalidUserLogin)
	}

	loginLength := utf8.RuneCountInString(i.Login)
	if loginLength > MaxUserLoginLength {
		return fmt.Errorf("invalid user login '%s' length '%d': %w", i.Login, loginLength, ErrInvalidUserLogin)
	}

	if i.Password == "" {
		return fmt.Errorf("empty user password: %w", ErrInvalidUserPassword)
	}

	passwordLength := utf8.RuneCountInString(i.Password)
	if passwordLength > MaxUserPasswordLength {
		return fmt.Errorf("invalid user password length '%d': %w", passwordLength, ErrInvalidUserLogin)
	}

	return nil
}

func (i *LoginUserInput) Validate() error {
	if i.Login == "" {
		return fmt.Errorf("empty user login: %w", ErrInvalidUserLogin)
	}

	if i.Password == "" {
		return fmt.Errorf("empty user password: %w", ErrInvalidUserPassword)
	}

	return nil
}
