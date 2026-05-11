package service

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

type FindUserOrderSpecification struct {
	Statuses        []model.UserOrderStatus
	BeforeUpdatedAt time.Time
	Limit           *uint
	Offset          *uint
}

type ExecutedUserOrderProvider interface {
	FindOrderIDToExecuteAtMap(ctx context.Context, spec FindUserOrderSpecification) (map[string]time.Time, error)
}

type OrderIDsProducer interface {
	Start(ctx context.Context)
	Stop() bool
	Produce() <-chan string
}

// NewOrderIDsProducer - поставщик канала с заявками на обработку хранящимися в БД.
// channelSize - размер канала
// selectLimit - лимит записей в выборке за один раз
// limitRetriesOnError - количество повторных попыток выборки из БД при получении ошибки
// waitingOnError - время ожидания после получения ошибки перед следующей попыткой
// будет выбирать записи из БД порциями в selectLimit и помещать их в канал, до тех пор пока в БД будут записи
func NewOrderIDsProducer(
	name string,
	channelSize uint,
	selectLimit *uint,
	limitRetriesOnError uint,
	waitingOnError time.Duration,
	executedUserOrderProvider ExecutedUserOrderProvider,
	retryProducer async.Producer[string],
	logger *logrus.Logger,
) OrderIDsProducer {
	return &orderIDsProducer{
		name:                      name,
		channelSize:               channelSize,
		selectLimit:               selectLimit,
		waitingOnError:            waitingOnError,
		limitRetriesOnError:       limitRetriesOnError,
		executedUserOrderProvider: executedUserOrderProvider,
		retryProducer:             retryProducer,
		logger:                    logger,
		ch:                        nil,
		closed:                    true,
		started:                   false,
		cancel:                    nil,
	}
}

type orderIDsProducer struct {
	ctx                       context.Context
	name                      string
	channelSize               uint
	selectLimit               *uint
	limitRetriesOnError       uint
	waitingOnError            time.Duration
	executedUserOrderProvider ExecutedUserOrderProvider
	retryProducer             async.Producer[string]
	logger                    *logrus.Logger
	ch                        chan string
	closed                    bool
	started                   bool
	cancel                    async.CancelFunc
	mu                        sync.RWMutex
}

func (p *orderIDsProducer) Produce() <-chan string {
	return p.ch
}

//nolint:gocognit
func (p *orderIDsProducer) Start(ctx context.Context) {
	p.mu.RLock()
	isStarted := p.started
	p.mu.RUnlock()

	if isStarted {
		return
	}

	p.mu.Lock()

	p.ctx = ctx

	startTime := time.Now()
	p.logger.Infof("'%s' producer started at %v", p.name, startTime)

	p.ch = make(chan string, p.channelSize)
	p.closed = false

	doneCh := make(chan struct{})
	var once sync.Once

	cancel := func() bool {
		stopped := false
		once.Do(func() {
			close(doneCh)
			stopped = true
		})
		return stopped
	}

	p.cancel = cancel
	p.started = true

	p.mu.Unlock()

	go func() {
		defer func() {
			p.mu.Lock()
			close(p.ch)
			p.closed = true
			p.started = false
			p.cancel = nil
			p.mu.Unlock()
		}()
		defer func() {
			p.logger.Infof("'%s' producer finished at %v (%v)", p.name, time.Now(), time.Since(startTime))
		}()

		offset := uint(0)
		limit := p.selectLimit
		retries := uint(0)

		for {
			select {
			case <-p.ctx.Done():
				p.logger.Infof("'%s' producer finished on context closed (%v)", p.name, p.ctx.Err())
				return
			case <-doneCh:
				p.logger.Infof("'%s' producer finished on cancel", p.name)
				return
			default:
				p.logger.Debugf("'%s' producer select (limit: %v, offset: %v)", p.name, limit, offset)
				orderIDToExecuteAtMap, err := p.selectOrderIDs(p.ctx, startTime, limit, &offset)
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

				if len(orderIDToExecuteAtMap) == 0 {
					p.logger.Debugf("'%s' producer finish on count values %d", p.name, len(orderIDToExecuteAtMap))
					return
				}

				for orderID, executeAt := range orderIDToExecuteAtMap {
					if executeAt.Before(time.Now()) {
						p.ch <- orderID
					} else {
						_ = p.retryProducer.Schedule(orderID, time.Until(executeAt))
					}
				}

				if limit == nil {
					p.logger.Debugf("'%s' producer finish on limit %v", p.name, limit)
					return
				}

				offset += uint(len(orderIDToExecuteAtMap))
			}
		}
	}()
}

func (p *orderIDsProducer) Stop() bool {
	if p.cancel != nil {
		return p.cancel()
	}
	return false
}

func (p *orderIDsProducer) selectOrderIDs(ctx context.Context, updatedAt time.Time, limit, offset *uint) (map[string]time.Time, error) {
	spec := FindUserOrderSpecification{
		Statuses:        []model.UserOrderStatus{model.UserOrderStatusNew, model.UserOrderStatusProcessing},
		BeforeUpdatedAt: updatedAt,
		Limit:           limit,
		Offset:          offset,
	}
	orderIDToExecuteAtMap, err := p.executedUserOrderProvider.FindOrderIDToExecuteAtMap(ctx, spec)
	if err != nil {
		return nil, err
	}
	return orderIDToExecuteAtMap, nil
}
