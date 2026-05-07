package async

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type RequestProducer interface {
	Produce() <-chan string
	Add(value string)
}

type RetryProducer interface {
	Produce() <-chan string
	Schedule(value string, delay time.Duration) *time.Timer
}

type CancelFunc func() bool

type Producer interface {
	Produce() <-chan string
	Add(value string)
	Schedule(value string, delay time.Duration) CancelFunc
}

func NewProducer(
	ctx context.Context,
	name string,
	channelSize uint,
	logger *logrus.Logger,
) Producer {
	p := &producer{
		ctx:    ctx,
		name:   name,
		ch:     make(chan string, channelSize),
		logger: logger,
	}

	startTime := time.Now()
	p.logger.Infof("'%s' producer started at %v", p.name, startTime)

	go func() {
		<-ctx.Done()
		p.logger.Infof("'%s' producer closed: context closed", p.name)

		p.mu.Lock()
		close(p.ch)
		p.closed = true
		p.mu.Unlock()

		p.logger.Infof("'%s' producer finished at %v (%v)", p.name, time.Now(), time.Since(startTime))
	}()

	return p
}

type producer struct {
	ctx    context.Context
	name   string
	logger *logrus.Logger
	ch     chan string
	closed bool
	mu     sync.RWMutex
}

func (p *producer) Produce() <-chan string {
	return p.ch
}

func (p *producer) Add(value string) {
	p.mu.RLock()
	isClosed := p.closed
	ch := p.ch
	p.mu.RUnlock()

	if isClosed || ch == nil {
		return
	}

	select {
	case ch <- value:
		p.logger.Debugf("'%s' producer add value (%s)", p.name, value)
	case <-p.ctx.Done():
		p.logger.Debugf("'%s' producer context closed (%v) on add value (%v)", p.name, p.ctx.Err(), value)
		return
	}
}

func (p *producer) Schedule(value string, delay time.Duration) CancelFunc {
	p.logger.Debugf("'%s' producer run schedule add value (%s) delay (%v)", p.name, value, delay)

	timer := time.AfterFunc(delay, func() {
		select {
		case <-p.ctx.Done():
			return
		default:
			p.Add(value)
		}
	})

	return func() bool {
		return timer.Stop()
	}
}
