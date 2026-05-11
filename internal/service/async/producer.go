package async

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CancelFunc func() bool

// Producer - поставщик канала данных с типом T
type Producer[T any] interface {
	Start(ctx context.Context)
	Stop() bool

	Produce() <-chan T

	Add(value T)
	Schedule(value T, delay time.Duration) CancelFunc
	Setup(inputCh <-chan T) CancelFunc
}

// NewProducer - поставщик данных через канал
// name - название поставщика (для логирования)
// channelSize - размер канала
func NewProducer[T any](
	name string,
	channelSize uint,
	logger *logrus.Logger,
) Producer[T] {
	return &producer[T]{
		name:        name,
		channelSize: channelSize,
		ch:          nil,
		closed:      true,
		started:     false,
		logger:      logger,
	}
}

type producer[T any] struct {
	ctx         context.Context
	name        string
	channelSize uint
	logger      *logrus.Logger
	ch          chan T
	closed      bool
	started     bool
	cancel      CancelFunc
	mu          sync.RWMutex
}

func (p *producer[T]) Start(ctx context.Context) {
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

	p.ch = make(chan T, p.channelSize)
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

		for {
			select {
			case <-p.ctx.Done():
				p.logger.Infof("'%s' producer finished on context closed (%v)", p.name, p.ctx.Err())
				return
			case <-doneCh:
				p.logger.Infof("'%s' producer finished on cancel", p.name)
				return
			}
		}
	}()
}

func (p *producer[T]) Stop() bool {
	if p.cancel != nil {
		return p.cancel()
	}
	return false
}

func (p *producer[T]) Produce() <-chan T {
	return p.ch
}

func (p *producer[T]) Add(value T) {
	p.mu.RLock()
	isClosed := p.closed
	ch := p.ch
	p.mu.RUnlock()

	if isClosed || ch == nil {
		return
	}

	select {
	case ch <- value:
		p.logger.Debugf("'%s' producer add value (%v)", p.name, value)
	case <-p.ctx.Done():
		p.logger.Debugf("'%s' producer context closed (%v) on add value (%v)", p.name, p.ctx.Err(), value)
		return
	}
}

func (p *producer[T]) Schedule(value T, delay time.Duration) CancelFunc {
	p.logger.Debugf("'%s' producer run schedule add value (%v) delay (%v)", p.name, value, delay)

	timer := time.AfterFunc(delay, func() {
		select {
		case <-p.ctx.Done():
			p.logger.Debugf("'%s' producer context closed (%v) on schedule value", p.name, p.ctx.Err())
			return
		default:
			p.Add(value)
		}
	})

	return func() bool {
		return timer.Stop()
	}
}

func (p *producer[T]) Setup(inputCh <-chan T) CancelFunc {
	p.mu.RLock()
	isStarted := p.started
	p.mu.RUnlock()

	if !isStarted {
		return nil
	}

	startTime := time.Now()

	p.logger.Infof("'%s' producer setup started at %v", p.name, startTime)

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

	go func() {
		defer func() {
			p.logger.Infof("'%s' producer setup finished at %v (%v)", p.name, startTime, time.Since(startTime))
		}()

		for {
			select {
			case <-p.ctx.Done():
				p.logger.Infof("'%s' producer setup finished on context closed (%v)", p.name, p.ctx.Err())
				return
			case <-doneCh:
				p.logger.Infof("'%s' producer setup finished on cancel", p.name)
				return
			case value, ok := <-inputCh:
				if !ok {
					p.logger.Infof("'%s' producer setup finished on closed input channel", p.name)
					return
				}
				p.Add(value)
			}
		}
	}()

	return cancel
}
