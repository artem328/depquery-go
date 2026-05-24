package jen

import . "github.com/dave/jennifer/jen"

func (r *Renderer) renderCompilerInterface() {
	typeParamBuilder := Id("B")
	typeParamRoot := Id("R")

	r.f.Add(block(
		Type().Id(r.naming.Compiler.Interface).Types(typeParamBuilder, Add(typeParamRoot).Any()).Interface(
			r.generateCompilerWithMethodSignature(typeParamBuilder, typeParamRoot, Null()),
			r.generateCompilerBuilderMethodSignature(typeParamBuilder),
			r.generateCompilerCompileMethodSignature(typeParamRoot, Null()),
		),
	))
}

func (r *Renderer) generateCompilerInterfaceType(typeParamBuilder, typeParamRoot Code) Code {
	return Id(r.naming.Compiler.Interface).Types(typeParamBuilder, typeParamRoot)
}

func (r *Renderer) generateCompilerWithMethodSignature(typeParamBuilder, typeParamRoot, funcVar Code) Code {
	return Id(r.naming.Compiler.WithMethod).Params(
		r.generateBuilderCallbackParam(typeParamBuilder, funcVar),
	).Params(
		r.generateCompilerInterfaceType(typeParamBuilder, typeParamRoot),
	)
}

func (r *Renderer) generateCompilerWithMethodBody(rcv, f Code) []Code {
	return []Code{
		If(Add(f).Op("!=").Nil()).Block(
			Add(f).Call(rcv),
		),
		Empty(),
		Return(rcv),
	}
}

func (r *Renderer) generateCompilerWithMethodCall(rcv, f Code) Code {
	return Add(rcv).Dot(r.naming.Compiler.WithMethod).Call(f)
}

func (r *Renderer) generateCompilerBuilderMethodSignature(typeParamBuilder Code) Code {
	return Id(r.naming.Compiler.BuilderMethod).Params().Params(typeParamBuilder)
}

func (r *Renderer) generateCompilerBuilderMethodBody(rcv Code) []Code {
	return []Code{
		Return(rcv),
	}
}

func (r *Renderer) generateCompilerCompileMethodSignature(typeParamRoot, opts Code) Code {
	return Id(r.naming.Compiler.CompileMethod).Params(Add(opts).Op("...").Add(libCompilerOption)).Params(r.generatePlanInterfaceType(typeParamRoot))
}

func (r *Renderer) generateCompilerCompileMethodBody(typeParamRoot, opts, fetchCtxConstructor, buildContext Code) []Code {
	config := Id("c")
	opt := Id("opt")

	return []Code{
		Add(config).Op(":=").Id(r.naming.Compiler.DefaultConfig),
		For(List(Id("_"), opt).Op(":=").Range().Add(opts)).Block(
			Add(opt).Dot("Apply").Call(Op("&").Add(config)),
		),
		Empty(),
		Return(r.generatePlanStructInit(typeParamRoot, fetchCtxConstructor, r.generateBuildContextPlanMethodCall(buildContext, Add(config).Dot("Planner")))),
	}
}

func (r *Renderer) renderCompilerDefaultConfig() {
	r.f.Add(block(
		Var().Id(r.naming.Compiler.DefaultConfig).Op("=").Add(libCompilerConfig).Add(valuesMultiline(
			Id("Planner").Op(":").Add(libBFSPlanner).Values(),
		)),
	))
}
