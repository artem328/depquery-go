package schema

type Definition interface {
	DefinedAt() string
}

type (
	Schema []Entity

	Entity struct {
		ID         Value[string]
		Name       Value[string]
		Type       Type
		Relations  []Relation
		Variants   []Variant
		Definition Definition
	}

	Type struct {
		Base       Value[string]
		Params     []Type
		Definition Definition
	}

	Relation struct {
		Name       Value[string]
		Entity     Value[string]
		ReversedBy Value[string]
		Definition Definition
	}

	Variant struct {
		Name       Value[string]
		Type       Type
		Relations  []Relation
		Definition Definition
	}
)
