package depquery

import (
	"context"
	"sync"
)

var _ Executor = ConcurrentExecutor{}

type ConcurrentExecutor struct{}

func (ConcurrentExecutor) Execute(ctx context.Context, tasks []Task) error {
	if len(tasks) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, 1)
	defer close(errChan)

	var (
		wg      sync.WaitGroup
		errOnce sync.Once
	)

	wg.Add(len(tasks))
	for _, task := range tasks {
		go func() {
			defer wg.Done()

			if err := task(ctx); err != nil {
				errOnce.Do(func() {
					errChan <- err
				})
			}
		}()
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
}
