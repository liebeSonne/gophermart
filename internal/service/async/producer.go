package async

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CancelFunc func() bool

type Producer interface {
	Produce() <-chan string
	Add(value string)
	Schedule(value string, delay time.Duration) CancelFunc
	Setup(inputCh <-chan string) CancelFunc
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
		defer func() {
			p.mu.Lock()
			close(p.ch)
			p.closed = true
			p.mu.Unlock()
		}()
		defer func() {
			p.logger.Infof("'%s' producer finished at %v (%v)", p.name, time.Now(), time.Since(startTime))
		}()

		<-ctx.Done()
		p.logger.Infof("'%s' producer closed on context closed", p.name)
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

func (p *producer) Setup(inputCh <-chan string) CancelFunc {
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
