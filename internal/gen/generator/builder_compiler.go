package generator

import . "github.com/dave/jennifer/jen"

func (i *nameIndex) CompilerInterface() string {
	return "Compiler"
}

func (i *nameIndex) CompilerWithMethod() string {
	return "With"
}

func (i *nameIndex) CompilerBuilderMethod() string {
	return "Builder"
}

func (i *nameIndex) CompilerCompileMethod() string {
	return "Compile"
}

type compilerBuilder struct {
	naming *nameIndex
	iface  Code
}

func newCompilerBuilder(naming *nameIndex) *compilerBuilder {
	return &compilerBuilder{naming: naming}
}

func (b *compilerBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildIface)}
}

func (b *compilerBuilder) Interface() Code {
	return b.iface
}

func (b *compilerBuilder) buildIface() {
	//	type <iface>[B, R any] interface {
	//		<with>(func(B)) <iface>[B, R]
	//		<builder>() B
	//		<compile>() <planIface>[R]
	//	}
	b.iface = Type().Id(b.naming.CompilerInterface()).Types(Id("B"), Id("R").Any()).Interface(
		Id(b.naming.CompilerWithMethod()).
			Params(Func().Params(Id("B"))).
			Id(b.naming.CompilerInterface()).Types(Id("B"), Id("R")),
		Id(b.naming.CompilerBuilderMethod()).Params().Id("B"),
		Id(b.naming.CompilerCompileMethod()).Params().Id(b.naming.PlanInterface()).Types(Id("R")),
	).Line()
}
