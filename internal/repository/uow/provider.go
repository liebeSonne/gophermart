package uow

import (
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type RepositoryProvider interface {
	UserRepository() repository.UserRepository
	UserOrderRepository() repository.UserOrderRepository
	UserBalanceRepository() repository.UserBalanceRepository
	UserBalanceWithdrawnRepository() repository.UserBalanceWithdrawnRepository
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

func (u *repositoryProvider) UserBalanceRepository() repository.UserBalanceRepository {
	return repository.NewUserBalanceRepository(u.client)
}

func (u *repositoryProvider) UserBalanceWithdrawnRepository() repository.UserBalanceWithdrawnRepository {
	return repository.NewUserBalanceWithdrawnRepository(u.client)
}
