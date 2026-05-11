package handler

import (
	"context"
	"errors"

	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrUserLoginExists = errors.New("user login exists")
var ErrUserNotFound = errors.New("user not found")
var ErrNotValidUserLoginPassword = errors.New("not valid user login password")
var ErrInvalidUserLogin = errors.New("invalid user login")
var ErrInvalidUserPassword = errors.New("invalid user password")

type CreateUserInput struct {
	Login    string
	Password string `json:"-"`
}

type LoginUserInput struct {
	Login    string
	Password string `json:"-"`
}

type UserService interface {
	Create(ctx context.Context, input CreateUserInput) (model.User, error)
	Login(ctx context.Context, input LoginUserInput) (model.User, error)
}
