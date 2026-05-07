package async

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
)

const DefaultSelectOrderIDsLimit = 100

type DBOrderIDsProducer interface {
	Produce(ctx context.Context, size uint) <-chan string
}

func NewDBOrderIDsProducer(
	userOrderProvider provider.UserOrderProvider,
	logger *logrus.Logger,
	selectLimit *uint,
) DBOrderIDsProducer {
	return &dbOrderIDsProducer{
		userOrderProvider: userOrderProvider,
		logger:            logger,
		selectLimit:       selectLimit,
	}
}

type dbOrderIDsProducer struct {
	userOrderProvider provider.UserOrderProvider
	logger            *logrus.Logger
	selectLimit       *uint
}

func (p *dbOrderIDsProducer) Produce(ctx context.Context, size uint) <-chan string {
	startTime := time.Now()

	p.logger.Infof("DB producer started at %v", startTime)

	ch := make(chan string, size)

	go func() {
		defer close(ch)
		defer func() {
			p.logger.Infof("DB producer finished at %v (%v)", time.Now(), time.Since(startTime))
		}()

		for {
			select {
			case <-ctx.Done():
				p.logger.Info("DB producer closed: context closed")
				return
			default:
				orderIDs, err := p.selectOrderIDs(ctx, startTime)
				if err != nil {
					p.logger.WithError(err).Errorf("DB producer failed to selectOrderIDs: %v", err)
				}
				if len(orderIDs) == 0 {
					return
				}
				for _, id := range orderIDs {
					ch <- id
				}
			}
		}
	}()

	return ch
}

func (p *dbOrderIDsProducer) selectOrderIDs(ctx context.Context, executeAt time.Time) ([]string, error) {
	spec := model.FindUserOrderSpecification{
		Statuses:        []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing},
		BeforeExecuteAt: executeAt,
		Limit:           p.selectLimit,
	}
	orderIDs, err := p.userOrderProvider.FindOrderIDs(ctx, spec)
	if err != nil {
		return nil, err
	}
	return orderIDs, nil
}
