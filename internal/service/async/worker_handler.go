package async

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Worker - интерфейс обработчика единицы данных I, возвращающего результат O
type Worker[I, O any] interface {
	Handle(value I) O
	SleepingHandle(result O) (bool, time.Duration)
}

// WorkerHandleFunc - интерфейс метода обработки единицы данных I, возвращающего результат O
type WorkerHandleFunc[I, O any] func(value I) O

// WorkerSleepingHandleFunc - метод определяющий по результату обработки одного обработчика, нужно ли заснуть всем обработчикам
type WorkerSleepingHandleFunc[O any] func(result O) (bool, time.Duration)

// WorkerHandler - интерфейс обработчика канала I, возвращающего результаты в канал O
type WorkerHandler[I, O any] interface {
	Handle(inCh <-chan I, outCh chan<- O, countWorkers uint)
}

// NewWorkerHandler - обработчик, запускает обработку данных канала I и возвращает результаты в канал O
func NewWorkerHandler[I, O any](
	ctx context.Context,
	name string,
	handler WorkerHandleFunc[I, O],
	sleepingHandler WorkerSleepingHandleFunc[O],
	logger *logrus.Logger,
) WorkerHandler[I, O] {
	return &workerHandler[I, O]{
		ctx:             ctx,
		name:            name,
		handler:         handler,
		sleepingHandler: sleepingHandler,
		logger:          logger,
	}
}

type workerHandler[I, O any] struct {
	ctx             context.Context
	name            string
	handler         WorkerHandleFunc[I, O]
	sleepingHandler WorkerSleepingHandleFunc[O]
	logger          *logrus.Logger
	mu              sync.RWMutex
	isSleeping      bool
	sleepingTime    time.Duration
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
				return
			}

			h.mu.RLock()
			sleeping := h.isSleeping
			sleepingTime := h.sleepingTime
			h.mu.RUnlock()

			if sleeping {
				h.logger.Infof("'%s' worker handler worker sleeping (%v) by sleeping status", h.name, sleepingTime)
				time.Sleep(sleepingTime)
			}

			h.logger.Debugf("'%s' worker handler handle value (%+v) ", h.name, value)

			result := h.handler(value)
			outCh <- result

			// Каждый из обработчиков может вернуть результат, который приведёт к засыпанию всех обработчиков
			needSleep, sleepingTime := h.sleepingHandler(result)

			if needSleep {
				h.mu.Lock()
				if !h.isSleeping {
					h.isSleeping = true
					h.sleepingTime = sleepingTime
					go func() {
						h.logger.Infof("'%s' worker handler start sleeping (%v) for all workers", h.name, h.sleepingTime)
						time.Sleep(sleepingTime)
						h.mu.Lock()
						h.isSleeping = false
						h.sleepingTime = 0
						h.mu.Unlock()
						h.logger.Infof("'%s' worker handler end sleeping for all workers", h.name)
					}()
				}
				h.mu.Unlock()

				h.logger.Infof("'%s' worker handler worker sleeping (%v) by result", h.name, sleepingTime)
				time.Sleep(sleepingTime)
			}
		}
	}
}
