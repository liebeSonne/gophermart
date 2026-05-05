package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository"
)

var ErrUserLoginExists = errors.New("user login exists")
var ErrUserNotFound = errors.New("user not found")
var ErrNotValidUserLoginPassword = errors.New("not valid user login password")

type UserService interface {
	Create(ctx context.Context, input CreateUserInput) (model.User, error)
	Login(ctx context.Context, input LoginUserInput) (model.User, error)
}

func NewUserService(
	userRepository repository.UserRepository,
	passwordService PasswordService,
) UserService {
	return &userService{
		userRepository:  userRepository,
		passwordService: passwordService,
	}
}

type userService struct {
	userRepository  repository.UserRepository
	passwordService PasswordService
}

func (s *userService) Create(ctx context.Context, input CreateUserInput) (model.User, error) {
	err := input.Validate()
	if err != nil {
		return model.User{}, err
	}

	user, err := s.userRepository.FindByLogin(ctx, input.Login)
	if err != nil {
		return model.User{}, err
	}
	if user != nil {
		return model.User{}, ErrUserLoginExists
	}

	passHash, err := s.passwordService.CreateHash(ctx, input.Password)
	if err != nil {
		return model.User{}, err
	}

	userID := s.userRepository.NextID(ctx)

	newUser := model.User{
		ID:       userID,
		Login:    input.Login,
		PassHash: passHash,
	}

	err = s.userRepository.Store(ctx, newUser)
	if err != nil {
		return model.User{}, err
	}

	return newUser, nil
}

func (s *userService) Login(ctx context.Context, input LoginUserInput) (model.User, error) {
	err := input.Validate()
	if err != nil {
		return model.User{}, err
	}

	user, err := s.userRepository.FindByLogin(ctx, input.Login)
	if err != nil {
		return model.User{}, err
	}
	if user == nil {
		return model.User{}, fmt.Errorf("not found user by login '%s': %w", input.Login, ErrUserNotFound)
	}

	ok, err := s.passwordService.CheckHash(ctx, input.Password, user.PassHash)
	if err != nil {
		return model.User{}, err
	}

	if !ok {
		return model.User{}, ErrNotValidUserLoginPassword
	}

	return *user, nil
}
