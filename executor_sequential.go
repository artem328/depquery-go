package depquery

import "context"

var _ Executor = SequentialExecutor{}

type SequentialExecutor struct{}

func (SequentialExecutor) Execute(ctx context.Context, tasks []Task) error {
	for _, task := range tasks {
		if err := task(ctx); err != nil {
			return err
		}
	}

	return nil
}
