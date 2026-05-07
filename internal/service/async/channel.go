package async

import (
	"context"
	"sync"
)

func FanIn[T any](ctx context.Context, chs ...<-chan T) <-chan T {
	outCh := make(chan T)

	var wg sync.WaitGroup

	for _, ch := range chs {
		c := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for n := range c {
				select {
				case <-ctx.Done():
					return
				case outCh <- n:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
