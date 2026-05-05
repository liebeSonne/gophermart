package uow

import (
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type RepositoryProvider interface {
	UserRepository() repository.UserRepository
	UserOrderRepository() repository.UserOrderRepository
}

func NewRepositoryProvider(
	client database.ContextClient,
) RepositoryProvider {
	return &repositoryProvider{
		client: client,
	}
}

type repositoryProvider struct {
	client database.ContextClient
}

func (u *repositoryProvider) UserRepository() repository.UserRepository {
	return repository.NewUserRepository(u.client)
}

func (u *repositoryProvider) UserOrderRepository() repository.UserOrderRepository {
	return repository.NewUserOrderRepository(u.client)
}
