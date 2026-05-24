package depquery

import "context"

type Task func(ctx context.Context) error

type Executor interface {
	Execute(context.Context, []Task) error
}
