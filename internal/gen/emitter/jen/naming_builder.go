package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
)

type builderNaming struct {
	FieldContext   string
	FieldID        string
	FieldParentID  string
	FieldRelations string
	FieldNested    string
	Interface      []string
	Struct         []string
	Constructor    []string
	Entrypoint     []string
	ChildBuilder   []string
	Method         map[plan.BuilderMethodID]string
}

func (n *builderNaming) warmUp(p plan.Plan) {
	n.FieldContext = "_ctx"
	n.FieldID = "_id"
	n.FieldParentID = "_parentID"
	n.FieldRelations = "_relations"
	n.FieldNested = "_nested"

	n.Interface = make([]string, len(p.Builders))
	n.Struct = make([]string, len(p.Builders))
	n.Constructor = make([]string, len(p.Builders))
	n.Entrypoint = make([]string, len(p.Builders))
	n.ChildBuilder = make([]string, len(p.Builders))
	n.Method = make(map[plan.BuilderMethodID]string)

	for id, b := range p.Builders {
		bid := plan.BuilderID(id)

		n.Interface[bid] = resolveBuilderInterfaceName(p, bid)
		n.Struct[bid] = resolveBuilderStructName(p, bid)
		n.Constructor[bid] = resolveBuilderConstructorName(p, bid)
		n.Entrypoint[bid] = resolveBuilderEntrypointName(p, bid)
		n.ChildBuilder[bid] = resolveBuilderChildBuilderName(p, bid)

		for _, m := range b.GetMethods() {
			n.Method[m.GetID()] = resolveBuilderMethodName(p, m)
		}
	}
}

func resolveBuilderInterfaceName(p plan.Plan, bid plan.BuilderID) string {
	b := p.Builders[bid]

	switch bb := b.(type) {
	case plan.RootBuilder:
		return sanitizeID(p.Model.Entities[bb.Entity].Name, sanitizeExported) + "Builder"
	case plan.VariantBuilder:
		variant := p.Model.Variants[bb.Variant]
		return sanitizeID(p.Model.Entities[variant.Entity].Name, sanitizeExported) + sanitizeID(variant.Name, sanitizeRawCapitalized) + "Builder"
	default:
		panic(fmt.Errorf("unknown builder type: %T", b))
	}
}

func resolveBuilderStructName(p plan.Plan, bid plan.BuilderID) string {
	b := p.Builders[bid]

	switch bb := b.(type) {
	case plan.RootBuilder:
		return sanitizeID(p.Model.Entities[bb.Entity].Name, sanitizeUnexported) + "Builder"
	case plan.VariantBuilder:
		variant := p.Model.Variants[bb.Variant]
		return sanitizeID(p.Model.Entities[variant.Entity].Name, sanitizeUnexported) + "V" + sanitizeID(variant.Name, sanitizeRawCapitalized) + "Builder"
	default:
		panic(fmt.Errorf("unknown builder type: %T", b))
	}
}

func resolveBuilderConstructorName(p plan.Plan, bid plan.BuilderID) string {
	return "new" + sanitizeID(resolveBuilderStructName(p, bid), sanitizeRawCapitalized)
}

func resolveBuilderEntrypointName(p plan.Plan, bid plan.BuilderID) string {
	b := p.Builders[bid]

	switch bb := b.(type) {
	case plan.RootBuilder:
		return sanitizeID(p.Model.Entities[bb.Entity].Name, sanitizeExported)
	case plan.VariantBuilder:
		return ""
	default:
		panic(fmt.Errorf("unknown builder type: %T", b))
	}
}

func resolveBuilderMethodName(p plan.Plan, m plan.BuilderMethod) string {
	switch mm := m.(type) {
	case plan.EnableBuilderMethod:
		return "With" + sanitizeID(p.Model.Relations[mm.Relation].Name, sanitizeRawCapitalized)
	case plan.DeepBuilderMethod:
		return sanitizeID(p.Model.Relations[mm.Relation].Name, sanitizeExported)
	case plan.VariantBuilderMethod:
		return "If" + sanitizeID(p.Model.Variants[mm.Variant].Name, sanitizeRawCapitalized)
	case plan.NestedBuilderMethod:
		return "In" + sanitizeID(p.Model.Nesteds[mm.Nested].Name, sanitizeRawCapitalized)
	default:
		panic(fmt.Errorf("unknown builder method type: %T", m))
	}
}

func resolveBuilderChildBuilderName(p plan.Plan, bid plan.BuilderID) string {
	b := p.Builders[bid]

	switch bb := b.(type) {
	case plan.RootBuilder:
		return sanitizeID(p.Model.Entities[bb.Entity].Name, sanitizeUnexported)
	case plan.VariantBuilder:
		variant := p.Model.Variants[bb.Variant]
		return "v" + sanitizeID(variant.Name, sanitizeRawCapitalized)
	default:
		panic(fmt.Errorf("unknown builder type: %T", b))
	}
}
