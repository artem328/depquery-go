package generator

import (
	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) EntityPrefetcherInterface() string {
	return "EntityPrefetcher"
}

func (i *nameIndex) EntityPrefetcherMethodName(e *schema.Entity) string {
	return i.getOrCreate(nKey("EntityPrefetcherMethodName", e.Name, "", ""), func() string {
		return "Prefetch" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) EntityPrefetcherReversedMethodName(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("EntityPrefetcherReversedMethodName", e.Name, "", by.Name), func() string {
		return "Prefetch" + sanitizeID(e.Name, sanitizeRawCapitalized) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

type entityPrefetcherBuilder struct {
	naming   *nameIndex
	entities []*schema.Entity
	revRefed []refedReversed
	iface    Code
}

func newEntityPrefetcherBuilder(
	naming *nameIndex,
	entities []*schema.Entity,
	revRefed []refedReversed,
) *entityPrefetcherBuilder {
	return &entityPrefetcherBuilder{naming: naming, entities: entities, revRefed: revRefed}
}

func (b *entityPrefetcherBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildIface)}
}

func (b *entityPrefetcherBuilder) Interface() Code {
	return b.iface
}

func (b *entityPrefetcherBuilder) buildIface() {
	b.iface = Type().Id("EntityPrefetcher").InterfaceFunc(func(i *Group) {
		for _, e := range b.entities {
			i.Add(b.methodSignature(e))
		}

		for _, r := range b.revRefed {
			i.Add(b.methodReversedSignature(r.Entity, r.By))
		}
	})
}

func (b *entityPrefetcherBuilder) methodSignature(e *schema.Entity) Code {
	//	<name>(context.Context, []<idType>) (iter.Seq[<entityType>], error)
	return Id(b.naming.EntityPrefetcherMethodName(e)).
		Params(contextCtx, Index().Add(typeToJen(e.ID.Type))).
		Params(iterSeq.Clone().Types(typeToJen(e.Type)), Error())
}

func (b *entityPrefetcherBuilder) methodReversedSignature(e, by *schema.Entity) Code {
	//	<name>(context.Context, []<byIdType>) (iter.Seq[<entityType>], error)
	return Id(b.naming.EntityPrefetcherReversedMethodName(e, by)).
		Params(contextCtx, Index().Add(typeToJen(by.ID.Type))).
		Params(iterSeq.Clone().Types(typeToJen(e.Type)), Error())
}
