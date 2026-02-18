package generator

import (
	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) PrefetchResolverInterface() string {
	return "PrefetchResolver"
}

func (i *nameIndex) PrefetchResolverMethodName(e *schema.Entity, r schema.Relation) string {
	return i.getOrCreate(nKey("PrefetchResolverMethodName", e.Name, "", r.Name), func() string {
		return "Resolve" + sanitizeID(e.Name, sanitizeRawCapitalized) + sanitizeID(r.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) PrefetchResolverVariantMethodName(
	e *schema.Entity,
	v schema.EntityVariant,
	r schema.Relation,
) string {
	return i.getOrCreate(nKey("PrefetchResolverVariantMethodName", e.Name, v.Name, r.Name), func() string {
		return "Resolve" + sanitizeID(e.Name, sanitizeRawCapitalized) +
			"V" + sanitizeID(v.Name, sanitizeRawCapitalized) +
			sanitizeID(r.Name, sanitizeRawCapitalized)
	})
}

type prefetchResolverBuilder struct {
	naming   *nameIndex
	entities []*schema.Entity
	iface    Code
}

func newPrefetchResolverBuilder(naming *nameIndex, entities []*schema.Entity) *prefetchResolverBuilder {
	return &prefetchResolverBuilder{naming: naming, entities: entities}
}

func (b *prefetchResolverBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildIface)}
}

func (b *prefetchResolverBuilder) buildIface() {
	b.iface = Type().Id(b.naming.PrefetchResolverInterface()).InterfaceFunc(func(i *Group) {
		for _, e := range b.entities {
			for _, r := range e.Relations {
				if r.IsReversed() {
					continue
				}

				i.Add(b.methodSignature(b.naming.PrefetchResolverMethodName(e, r), e.Type, r))
			}

			for _, v := range e.Variants {
				for _, r := range v.Relations {
					if r.IsReversed() {
						continue
					}

					i.Add(b.methodSignature(b.naming.PrefetchResolverVariantMethodName(e, v, r), v.Type, r))
				}
			}
		}
	})
}

func (b *prefetchResolverBuilder) Interface() Code {
	return b.iface
}

func (b *prefetchResolverBuilder) methodSignature(name string, t schema.Type, r schema.Relation) Code {
	// <name>(<type>) iter.Seq[<relIdType>]
	return Id(name).
		Params(typeToJen(t)).Add(iterSeq).Types(typeToJen(r.Entity.ID.Type))
}
