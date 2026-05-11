package service

import (
	"fmt"
	"unicode/utf8"

	"github.com/liebeSonne/gophermart/internal/handler"
)

const MaxUserLoginLength = 255
const MaxUserPasswordLength = 255

type createUserInputValidator struct {
	Input handler.CreateUserInput
}

func (v *createUserInputValidator) Validate() error {
	if v.Input.Login == "" {
		return fmt.Errorf("empty user login: %w", handler.ErrInvalidUserLogin)
	}

	loginLength := utf8.RuneCountInString(v.Input.Login)
	if loginLength > MaxUserLoginLength {
		return fmt.Errorf("invalid user login '%s' length '%d': %w", v.Input.Login, loginLength, handler.ErrInvalidUserLogin)
	}

	if v.Input.Password == "" {
		return fmt.Errorf("empty user password: %w", handler.ErrInvalidUserPassword)
	}

	passwordLength := utf8.RuneCountInString(v.Input.Password)
	if passwordLength > MaxUserPasswordLength {
		return fmt.Errorf("invalid user password length '%d': %w", passwordLength, handler.ErrInvalidUserLogin)
	}

	return nil
}

type loginUserInputValidator struct {
	Input handler.LoginUserInput
}

func (v *loginUserInputValidator) Validate() error {
	if v.Input.Login == "" {
		return fmt.Errorf("empty user login: %w", handler.ErrInvalidUserLogin)
	}

	if v.Input.Password == "" {
		return fmt.Errorf("empty user password: %w", handler.ErrInvalidUserPassword)
	}

	return nil
}
