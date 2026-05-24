package depquery

type InstanceConfig struct {
	Executor Executor
}

type InstanceOption interface {
	Apply(*InstanceConfig)
	isInstanceOption()
}

func WithExecutor(executor Executor) InstanceOption {
	return instanceOptionFunc(func(c *InstanceConfig) {
		c.Executor = executor
	})
}

type instanceOptionFunc func(*InstanceConfig)

func (f instanceOptionFunc) Apply(c *InstanceConfig) {
	if f != nil {
		f(c)
	}
}

func (instanceOptionFunc) isInstanceOption() {}
