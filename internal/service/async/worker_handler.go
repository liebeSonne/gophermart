package async

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Worker[I, O any] interface {
	Handle(value I, outCh chan<- O)
}

type WorkerHandleFunc[I, O any] func(value I, resulCh chan<- O)

type WorkerHandler[I, O any] interface {
	Handle(inCh <-chan I, outCh chan<- O, countWorkers uint)
}

func NewWorkerHandler[I, O any](
	ctx context.Context,
	name string,
	handler WorkerHandleFunc[I, O],
	logger *logrus.Logger,
) WorkerHandler[I, O] {
	return &workerHandler[I, O]{
		ctx:     ctx,
		name:    name,
		handler: handler,
		logger:  logger,
	}
}

type workerHandler[I, O any] struct {
	ctx     context.Context
	name    string
	handler WorkerHandleFunc[I, O]
	logger  *logrus.Logger
}

func (h *workerHandler[I, O]) Handle(inCh <-chan I, outCh chan<- O, countWorkers uint) {
	h.logger.Infof("'%s' worker handler run %d workers", h.name, countWorkers)
	for i := uint(0); i < countWorkers; i++ {
		go h.runWorker(inCh, outCh)
	}
}

func (h *workerHandler[I, O]) runWorker(inCh <-chan I, outCh chan<- O) {
	for {
		select {
		case <-h.ctx.Done():
			h.logger.Debugf("'%s' worker handler worker finished on contetx closed (%v)", h.name, h.ctx.Err())
			return
		case value, ok := <-inCh:
			if !ok {
				h.logger.Debugf("'%s' worker handler worker finished on input channel closed", h.name)
				continue
			}
			h.handler(value, outCh)
		}
	}
}
