package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/handler"
	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, user model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
}

type UserProvider interface {
	FindByLogin(ctx context.Context, login string) (*model.User, error)
}

func NewUserService(
	uowFactory UnitOfWorkFactory,
	passwordService PasswordService,
	userProvider UserProvider,
) *UserService {
	return &UserService{
		uowFactory:      uowFactory,
		passwordService: passwordService,
		userProvider:    userProvider,
	}
}

type UserService struct {
	uowFactory      UnitOfWorkFactory
	passwordService PasswordService
	userProvider    UserProvider
}

func (s *UserService) Create(ctx context.Context, input handler.CreateUserInput) (model.User, error) {
	validator := createUserInputValidator{Input: input}
	err := validator.Validate()
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

	err = s.uowFactory.ExecuteWithUnitOfWork(ctx, lockNames, func(provider RepositoryProvider) error {
		userRepository := provider.UserRepository()

		var userPtr *model.User
		userPtr, err = userRepository.FindByLogin(ctx, input.Login)
		if err != nil {
			return err
		}
		if userPtr != nil {
			return handler.ErrUserLoginExists
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

func (s *UserService) Login(ctx context.Context, input handler.LoginUserInput) (model.User, error) {
	validator := loginUserInputValidator{Input: input}
	err := validator.Validate()
	if err != nil {
		return model.User{}, fmt.Errorf("invalid login user input: %w", err)
	}

	user, err := s.userProvider.FindByLogin(ctx, input.Login)
	if err != nil {
		return model.User{}, err
	}
	if user == nil {
		return model.User{}, fmt.Errorf("not found user by login '%s': %w", input.Login, handler.ErrUserNotFound)
	}

	ok, err := s.passwordService.CheckHash(ctx, input.Password, user.PassHash)
	if err != nil {
		return model.User{}, err
	}

	if !ok {
		return model.User{}, handler.ErrNotValidUserLoginPassword
	}

	return *user, nil
}
