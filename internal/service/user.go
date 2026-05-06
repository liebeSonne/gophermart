package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

var ErrUserLoginExists = errors.New("user login exists")
var ErrUserNotFound = errors.New("user not found")
var ErrNotValidUserLoginPassword = errors.New("not valid user login password")

type UserService interface {
	Create(ctx context.Context, input CreateUserInput) (model.User, error)
	Login(ctx context.Context, input LoginUserInput) (model.User, error)
}

func NewUserService(
	uowFactory uow.UnitOfWorkFactory,
	passwordService PasswordService,
	userProvider provider.UserProvider,
) UserService {
	return &userService{
		uowFactory:      uowFactory,
		passwordService: passwordService,
		userProvider:    userProvider,
	}
}

type userService struct {
	uowFactory      uow.UnitOfWorkFactory
	passwordService PasswordService
	userProvider    provider.UserProvider
}

func (s *userService) Create(ctx context.Context, input CreateUserInput) (model.User, error) {
	err := input.Validate()
	if err != nil {
		return model.User{}, fmt.Errorf("invalid create user input: %w", err)
	}

	passHash, err := s.passwordService.CreateHash(ctx, input.Password)
	if err != nil {
		return model.User{}, err
	}

	var newUser model.User

	lockName := MakeUsersLockName()
	lockNames := []string{lockName}

	err = s.uowFactory.ExecuteWithUnitOfWork(ctx, lockNames, func(provider uow.RepositoryProvider) error {
		userRepository := provider.UserRepository()

		var userPtr *model.User
		userPtr, err = userRepository.FindByLogin(ctx, input.Login)
		if err != nil {
			return err
		}
		if userPtr != nil {
			return ErrUserLoginExists
		}

		userID := userRepository.NextID(ctx)

		newUser = model.User{
			ID:       userID,
			Login:    input.Login,
			PassHash: passHash,
		}

		return userRepository.Store(ctx, newUser)
	})
	if err != nil {
		return model.User{}, err
	}

	return newUser, nil
}

func (s *userService) Login(ctx context.Context, input LoginUserInput) (model.User, error) {
	err := input.Validate()
	if err != nil {
		return model.User{}, fmt.Errorf("invalid login user input: %w", err)
	}

	user, err := s.userProvider.FindByLogin(ctx, input.Login)
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
