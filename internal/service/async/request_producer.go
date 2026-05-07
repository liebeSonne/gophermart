package async

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const DefaultRequestProducerChannelSize = 1000

type RequestProducer interface {
	Produce() <-chan string
	Add(value string)
}

func NewRequestProducer(
	ctx context.Context,
	channelSize uint,
	logger *logrus.Logger,
) RequestProducer {
	p := &requestProducer{
		ctx:    ctx,
		ch:     make(chan string, channelSize),
		logger: logger,
	}

	startTime := time.Now()
	p.logger.Infof("request producer started at %v", startTime)

	go func() {
		<-ctx.Done()
		p.logger.Info("request producer closed: context closed")

		p.mu.Lock()
		close(p.ch)
		p.closed = true
		p.mu.Unlock()

		p.logger.Infof("request producer finished at %v (%v)", time.Now(), time.Since(startTime))
	}()

	return p
}

type requestProducer struct {
	ctx    context.Context
	logger *logrus.Logger
	ch     chan string
	closed bool
	mu     sync.RWMutex
}

func (p *requestProducer) Produce() <-chan string {
	return p.ch
}

func (p *requestProducer) Add(value string) {
	p.mu.RLock()
	isClosed := p.closed
	ch := p.ch
	p.mu.RUnlock()

	if isClosed {
		return
	}

	select {
	case ch <- value:
		p.logger.Debugf("request producer add value: %v", value)
	case <-p.ctx.Done():
		p.logger.Debugf("request producer context closed (%v) on add value (%v)", p.ctx.Err(), value)
		return
	}
}
