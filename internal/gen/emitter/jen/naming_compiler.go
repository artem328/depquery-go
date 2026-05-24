package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type compilerNaming struct {
	Interface     string
	WithMethod    string
	BuilderMethod string
	CompileMethod string
	DefaultConfig string
}

func (n *compilerNaming) warmUp(plan.Plan) {
	n.Interface = "Compiler"
	n.WithMethod = "With"
	n.BuilderMethod = "Builder"
	n.CompileMethod = "Compile"
	n.DefaultConfig = "defaultCompilerConfig"
}
