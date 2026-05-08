package async

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
)

type OrderIDsProducer interface {
	Start()
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
	ctx context.Context,
	name string,
	channelSize uint,
	selectLimit *uint,
	limitRetriesOnError uint,
	waitingOnError time.Duration,
	userOrderProvider provider.UserOrderProvider,
	logger *logrus.Logger,
) OrderIDsProducer {
	return &orderIDsProducer{
		ctx:                 ctx,
		name:                name,
		channelSize:         channelSize,
		selectLimit:         selectLimit,
		waitingOnError:      waitingOnError,
		limitRetriesOnError: limitRetriesOnError,
		userOrderProvider:   userOrderProvider,
		logger:              logger,
		ch:                  nil,
		closed:              true,
		started:             false,
		cancel:              nil,
	}
}

type orderIDsProducer struct {
	ctx                 context.Context
	name                string
	channelSize         uint
	selectLimit         *uint
	limitRetriesOnError uint
	waitingOnError      time.Duration
	userOrderProvider   provider.UserOrderProvider
	logger              *logrus.Logger
	ch                  chan string
	closed              bool
	started             bool
	cancel              CancelFunc
	mu                  sync.RWMutex
}

func (p *orderIDsProducer) Produce() <-chan string {
	return p.ch
}

func (p *orderIDsProducer) Start() {
	p.mu.RLock()
	isStarted := p.started
	p.mu.RUnlock()

	if isStarted {
		return
	}

	p.mu.Lock()

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
				values, err := p.selectOrderIDs(p.ctx, startTime, limit, &offset)
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
					p.logger.Debugf("'%s' producer finish on count values %d", p.name, len(values))
					return
				}

				for _, v := range values {
					p.ch <- v
				}

				if limit == nil {
					p.logger.Debugf("'%s' producer finish on limit %v", p.name, limit)
					return
				}

				offset += uint(len(values))
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

func (p *orderIDsProducer) selectOrderIDs(ctx context.Context, executeAt time.Time, limit, offset *uint) ([]string, error) {
	spec := model.FindUserOrderSpecification{
		Statuses:        []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing},
		BeforeExecuteAt: executeAt,
		Limit:           limit,
		Offset:          offset,
	}
	orderIDs, err := p.userOrderProvider.FindOrderIDs(ctx, spec)
	if err != nil {
		return nil, err
	}
	return orderIDs, nil
}
