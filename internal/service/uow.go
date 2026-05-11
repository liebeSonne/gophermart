package service

import "context"

type RepositoryProvider interface {
	UserRepository() UserRepository
	UserOrderRepository() UserOrderRepository
	UserBalanceRepository() UserBalanceRepository
	UserBalanceWithdrawnRepository() UserBalanceWithdrawnRepository
}

type UnitOfWorkFactory interface {
	ExecuteWithUnitOfWork(ctx context.Context, lockNames []string, fn func(provider RepositoryProvider) error) error
}
