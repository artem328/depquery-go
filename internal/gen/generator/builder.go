package generator

type builder interface {
	Build()
}

type builderFunc func()

func (f builderFunc) Build() {
	if f == nil {
		return
	}

	f()
}

type builderResolver interface {
	Builders() []builder
}
