package uow

import (
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/service"
)

func NewRepositoryProvider(
	client repository.ContextClient,
) service.RepositoryProvider {
	return &repositoryProvider{
		client: client,
	}
}

type repositoryProvider struct {
	client repository.ContextClient
}

func (u *repositoryProvider) UserRepository() service.UserRepository {
	return repository.NewUserRepository(u.client)
}

func (u *repositoryProvider) UserOrderRepository() service.UserOrderRepository {
	return repository.NewUserOrderRepository(u.client)
}

func (u *repositoryProvider) UserBalanceRepository() service.UserBalanceRepository {
	return repository.NewUserBalanceRepository(u.client)
}

func (u *repositoryProvider) UserBalanceWithdrawnRepository() service.UserBalanceWithdrawnRepository {
	return repository.NewUserBalanceWithdrawnRepository(u.client)
}
