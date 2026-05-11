package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	ilogger "github.com/liebeSonne/gophermart/internal/logger"
	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

var ErrUnknownAccrualOrderStatus = errors.New("unknown accrual order status")

type UserOrderProvider interface {
	FindUserIDByOrderID(ctx context.Context, orderID string) (*uuid.UUID, error)
}

// NewOrderIDResultWorker - обработчик результатов проверки заказов
// retryDelay - интервал для повторной попытки обработки заявки (при не финальном статусе и при ошибке)
// minTooManyRequestsRetryDelay - интервал через который произойдет повторная попытка при ответе от внешнего сервиса о слишком большом числе запросов, при получении RetryAfter заголовка, интервал будет не меньше минимального
// maxTooManyRequestsRetryDelay - максимальный интервал через который пройдет повторная попытка, при получении RetryAfter заголовка которое больше максимального
// retryProducer - поставщик канала повторных попыток, принимающий запросы на отложенный запуск обработки заявок
func NewOrderIDResultWorker(
	ctx context.Context,
	name string,
	retryDelay time.Duration,
	minTooManyRequestsRetryDelay time.Duration,
	maxTooManyRequestsRetryDelay time.Duration,
	retryProducer async.Producer[string],
	uowFactory UnitOfWorkFactory,
	userOrderProvider UserOrderProvider,
	logger ilogger.Logger,
) async.Worker[OrderIDWorkerResult, struct{}] {
	return &orderIDResultWorker{
		ctx:                          ctx,
		name:                         name,
		retryDelay:                   retryDelay,
		minTooManyRequestsRetryDelay: minTooManyRequestsRetryDelay,
		maxTooManyRequestsRetryDelay: maxTooManyRequestsRetryDelay,
		retryProducer:                retryProducer,
		uowFactory:                   uowFactory,
		userOrderProvider:            userOrderProvider,
		logger:                       logger,
	}
}

type orderIDResultWorker struct {
	ctx                          context.Context
	name                         string
	retryDelay                   time.Duration
	minTooManyRequestsRetryDelay time.Duration
	maxTooManyRequestsRetryDelay time.Duration
	retryProducer                async.Producer[string]
	uowFactory                   UnitOfWorkFactory
	userOrderProvider            UserOrderProvider
	logger                       ilogger.Logger
}

func (w *orderIDResultWorker) Handle(result OrderIDWorkerResult) struct{} {
	doRetry := false
	executeAtDelay := w.calculateExecuteAtDelay(result.Err)

	defer func() {
		if doRetry {
			w.logger.Debugf("'%s' worker schedule retry order (%v) delay (%v)", w.name, result.OrderID, executeAtDelay)
			_ = w.retryProducer.Schedule(result.OrderID, executeAtDelay)
		}
	}()

	if result.Err != nil {
		w.logger.Debugf("'%s' worker do retry order (%v) on result error (%v)", w.name, result.OrderID, result.Err)
		doRetry = true
		return struct{}{}
	}

	var newOrderStatusPtr *model.UserOrderStatus
	newOrderStatusPtr, err := w.calculateNewOrderStatus(result.OrderID, result.Status)
	if err != nil {
		w.logger.Debugf("'%s' worker do retry order (%v) on new status error (%v)", w.name, result.OrderID, err)
		executeAtDelay = w.calculateExecuteAtDelay(err)
		doRetry = true
		return struct{}{}
	}

	userIDPtr, err := w.userOrderProvider.FindUserIDByOrderID(w.ctx, result.OrderID)
	if err != nil {
		doRetry = true
		executeAtDelay = w.calculateExecuteAtDelay(err)
		w.logger.Errorf("'%s' worker failed to find user by order (%v): %v", w.name, result.OrderID, err)
		w.logger.Debugf("'%s' worker do retry order (%v) on found user by order error (%v)", w.name, result.OrderID, err)
		return struct{}{}
	}
	if userIDPtr == nil {
		w.logger.Warnf("'%s' worker not found user by order (%v)", w.name, result.OrderID)
		return struct{}{}
	}

	lockNames := []string{
		MakeUserOrderLockName(result.OrderID),
		MakeUserBalanceLockName(*userIDPtr),
	}

	err = w.uowFactory.ExecuteWithUnitOfWork(w.ctx, lockNames, func(repositoryProvider RepositoryProvider) error {
		doRetry, executeAtDelay, err = w.updateUserOrder(result.OrderID, newOrderStatusPtr, result.Accrual, repositoryProvider)
		return err
	})
	if err != nil {
		doRetry = true
		executeAtDelay = w.calculateExecuteAtDelay(err)
		w.logger.Errorf("'%s' worker failed to handle user order (%v): %v", w.name, result.OrderID, err)
		w.logger.Debugf("'%s' worker do retry order (%v) on execute error (%v)", w.name, result.OrderID, err)
		return struct{}{}
	}

	return struct{}{}
}

func (w *orderIDResultWorker) SleepingHandle(_ struct{}) (bool, time.Duration) {
	return false, 0
}

func (w *orderIDResultWorker) calculateExecuteAtDelay(err error) time.Duration {
	if err == nil {
		return w.retryDelay
	}

	var retryErr *ErrTooManyRequestsRetryAfter
	if errors.As(err, &retryErr) && retryErr.RetryAfter > 0 {
		return min(max(w.minTooManyRequestsRetryDelay, retryErr.RetryAfter), w.maxTooManyRequestsRetryDelay)
	}

	if errors.Is(err, ErrTooManyRequests) {
		return w.minTooManyRequestsRetryDelay
	}

	if errors.Is(err, ErrUnknownAccrualOrderStatus) {
		return w.retryDelay
	}

	return w.retryDelay
}

func (w *orderIDResultWorker) calculateNewOrderStatus(orderID string, status *OrderStatus) (*model.UserOrderStatus, error) {
	var newOrderStatusPtr *model.UserOrderStatus
	if status != nil {
		newStatus, err := w.convertOrderStatus(*status)
		if err != nil {
			return nil, fmt.Errorf("'%s' worker failed to convert result order (%v) status (%v): %w", w.name, orderID, status, err)
		}
		newOrderStatusPtr = &newStatus
	}
	return newOrderStatusPtr, nil
}

func (w *orderIDResultWorker) convertOrderStatus(status OrderStatus) (model.UserOrderStatus, error) {
	switch status {
	case OrderStatusRegistered:
		return model.UserOrderStatusNew, nil
	case OrderStatusProcessing:
		return model.UserOrderStatusProcessing, nil
	case OrderStatusInvalid:
		return model.UserOrderStatusInvalid, nil
	case OrderStatusProcessed:
		return model.UserOrderStatusProcessed, nil
	default:
		return -1, fmt.Errorf("%w: %d", ErrUnknownAccrualOrderStatus, status)
	}
}

func (w *orderIDResultWorker) isFinalStatus(status model.UserOrderStatus) bool {
	switch status {
	case model.UserOrderStatusInvalid, model.UserOrderStatusProcessed:
		return true
	default:
		return false
	}
}

func (w *orderIDResultWorker) updateUserOrder(
	orderID string,
	newStatus *model.UserOrderStatus,
	newAccrual *decimal.Decimal,
	repositoryProvider RepositoryProvider,
) (
	doRetry bool,
	executeAtDelay time.Duration,
	err error,
) {
	orderRepository := repositoryProvider.UserOrderRepository()
	balanceRepository := repositoryProvider.UserBalanceRepository()

	var userOrderPtr *model.UserOrder
	userOrderPtr, err = orderRepository.FindByOrderID(w.ctx, orderID)
	if err != nil {
		return doRetry, executeAtDelay, err
	}
	// Пропускаем обработку если запись не найдена
	if userOrderPtr == nil {
		w.logger.Warnf("'%s' worker not found order (%v)", w.name, orderID)
		return doRetry, executeAtDelay, err
	}

	userOrder := *userOrderPtr
	// Не изменяем запись, если у неё уже финальный статус = обработанный
	if userOrder.Status == model.UserOrderStatusProcessed {
		w.logger.Warnf("'%s' worker not found user order (%v)", w.name, orderID)
		return doRetry, executeAtDelay, err
	}

	var changeBalance *decimal.Decimal
	if newStatus != nil {
		userOrder.Status = *newStatus
		// Изменяем начисление только при новом финальном статусе = обработанный
		if *newStatus == model.UserOrderStatusProcessed {
			userOrder.Accrual = newAccrual
			// Начисление в баланс только при переходе в завершенный статус = обработанный, с суммой начисления больше ноля
			if userOrder.Accrual != nil && userOrder.Accrual.GreaterThan(decimal.Zero) {
				changeBalance = userOrder.Accrual
			}
		}
	}
	isFinalStatus := w.isFinalStatus(userOrder.Status)
	// Повторная попытка обработать если статус не финальный
	if !isFinalStatus {
		executeAtDelay = w.calculateExecuteAtDelay(nil)
		userOrder.ExecuteAt = time.Now().Add(executeAtDelay)
		doRetry = true
		w.logger.Debugf("'%s' worker do retry order (%v) on not final status (%v)", w.name, orderID, userOrder.Status)
	}
	userOrder.Retries++

	err = orderRepository.Store(w.ctx, []model.UserOrder{userOrder})
	if err != nil {
		err = fmt.Errorf("'%s' worker failed to store result user order (%v): %w", w.name, orderID, err)
		return doRetry, executeAtDelay, err
	}

	// Начисление в баланс
	if changeBalance != nil {
		var userBalance model.UserBalance
		userBalance, err = balanceRepository.GetByUserID(w.ctx, userOrder.UserID)
		if err != nil {
			err = fmt.Errorf("'%s' worker failed to get user  (%v) balance: %w", w.name, userOrder.UserID, err)
			return doRetry, executeAtDelay, err
		}
		userBalance.Balance = userBalance.Balance.Add(*changeBalance)
		err = balanceRepository.Store(w.ctx, userBalance)
		if err != nil {
			err = fmt.Errorf("'%s' worker failed to store user (%v) balance: %w", w.name, userOrder.UserID, err)
			return doRetry, executeAtDelay, err
		}
	}

	return doRetry, executeAtDelay, err
}
