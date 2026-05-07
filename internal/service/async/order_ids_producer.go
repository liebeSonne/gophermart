package async

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
)

type OrderIDsProducer interface {
	Produce(size uint) <-chan string
}

func NewOrderIDsProducer(
	ctx context.Context,
	name string,
	executeAt time.Time,
	selectLimit *uint,
	limitRetriesOnError uint,
	waitingOnError time.Duration,
	userOrderProvider provider.UserOrderProvider,
	logger *logrus.Logger,
) OrderIDsProducer {
	return &orderIDsProducer{
		ctx:                 ctx,
		name:                name,
		selectLimit:         selectLimit,
		executeAt:           executeAt,
		waitingOnError:      waitingOnError,
		limitRetriesOnError: limitRetriesOnError,
		userOrderProvider:   userOrderProvider,
		logger:              logger,
	}
}

type orderIDsProducer struct {
	ctx                 context.Context
	name                string
	executeAt           time.Time
	selectLimit         *uint
	limitRetriesOnError uint
	waitingOnError      time.Duration
	userOrderProvider   provider.UserOrderProvider
	logger              *logrus.Logger
}

func (p *orderIDsProducer) Produce(size uint) <-chan string {
	startTime := time.Now()

	p.logger.Infof("'%s' producer started at %v", p.name, startTime)

	ch := make(chan string, size)

	go func() {
		defer close(ch)
		defer func() {
			p.logger.Infof("'%s' producer finished at %v (%v)", p.name, time.Now(), time.Since(startTime))
		}()

		offset := uint(0)
		limit := p.selectLimit
		retries := uint(0)

		for {
			select {
			case <-p.ctx.Done():
				p.logger.Infof("'%s' producer context closed (%v)", p.name, p.ctx.Err())
				return
			default:
				p.logger.Debugf("'%s' producer select (limit: %v, offset: %v)", p.name, limit, offset)
				values, err := p.selectOrderIDs(p.ctx, limit, &offset)
				if err != nil {
					p.logger.WithError(err).Errorf("'%s' producer error on select", p.name)
					retries++
					if retries > p.limitRetriesOnError {
						p.logger.Debugf("'%s' producer finish on error (retrice %d/%d)", p.name, retries, p.limitRetriesOnError)
						return
					}
					time.Sleep(p.waitingOnError)
					continue
				}

				if len(values) == 0 {
					p.logger.Debugf("'%s' producer finish on size %d", p.name, size)
					return
				}

				for _, v := range values {
					ch <- v
				}

				if limit == nil {
					p.logger.Debugf("'%s' producer finish on limit %v", p.name, limit)
					return
				}

				offset += uint(len(values))
			}
		}
	}()

	return ch
}

func (p *orderIDsProducer) selectOrderIDs(ctx context.Context, limit, offset *uint) ([]string, error) {
	spec := model.FindUserOrderSpecification{
		Statuses:        []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing},
		BeforeExecuteAt: p.executeAt,
		Limit:           limit,
		Offset:          offset,
	}
	orderIDs, err := p.userOrderProvider.FindOrderIDs(ctx, spec)
	if err != nil {
		return nil, err
	}
	return orderIDs, nil
}
