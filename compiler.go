package depquery

type CompilerConfig struct {
	Planner Planner
}

type CompilerOption interface {
	Apply(*CompilerConfig)
	isCompilerOption()
}

func WithPlanner(p Planner) CompilerOption {
	return compilerOptionFunc(func(c *CompilerConfig) {
		c.Planner = p
	})
}

type compilerOptionFunc func(*CompilerConfig)

func (f compilerOptionFunc) Apply(c *CompilerConfig) {
	if f != nil {
		f(c)
	}
}

func (compilerOptionFunc) isCompilerOption() {}
