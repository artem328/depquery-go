package semantic

import (
	"fmt"
	"go/types"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

type variantKey struct {
	EntityID EntityID
	Name     string
}

type variantEntry struct {
	ID   VariantID
	Type types.Type
}

func (r *Resolver) initVariants() {
	for _, e := range r.schema {
		entityName := e.Name.V
		entityID := r.seenEntities[entityName].ID

		for _, v := range e.Variants {
			// TODO: validate unique types
			name := r.resolveVariantName(entityName, v)

			key := variantKey{
				EntityID: entityID,
				Name:     name,
			}

			if _, ok := r.seenVariants[key]; ok {
				r.eerr(entityName, fmt.Sprintf("variant `%s` was already defined for the entity", v.Name.V), v.Definition)
				continue
			}

			tid, t, valImpl := r.addVariantType(e, v)

			r.seenVariants[key] = variantEntry{ID: VariantID(len(r.model.Variants)), Type: t}
			r.model.Variants = append(r.model.Variants, Variant{
				Name:            name,
				Entity:          entityID,
				Type:            tid,
				ValueAssignable: valImpl,
			})
		}
	}
}

func (r *Resolver) addVariantType(e schema.Entity, v schema.Variant) (_ TypeID, _ types.Type, valImpl bool) {
	var invalid bool
	entityName := e.Name.V

	et, ok := r.types.goTypes[e.Type.Base.V]
	if !ok {
		r.eerr(entityName, "entity type base must be existing non-scalar go type", e.Type.Definition)
		invalid = true
	}

	vt, ok := r.types.goTypes[v.Type.Base.V]
	if !ok {
		r.eerr(entityName, "entity variant type base must be existing non-scalar go type", v.Type.Definition)
		invalid = true
	}

	if invalid {
		return 0, nil, false
	}

	if et.IsBasic {
		r.eerr(entityName, "entity type base cannot be scalar type", e.Type.Definition)
		return 0, nil, false
	}

	if !et.TypeName.Exported() {
		r.eerr(entityName, "entity type base must be exported ("+et.TypeName.Type().String()+" is not)", e.Type.Definition)
		invalid = true
	}

	if !vt.TypeName.Exported() {
		r.eerr(entityName, "entity variant type base must be exported ("+vt.TypeName.Type().String()+" is not)", v.Type.Definition)
		invalid = true
	}

	if invalid {
		return 0, nil, false
	}

	ect, _ := r.composeType(entityName, e.Type)
	if ect == nil {
		return 0, nil, false
	}

	var iface *types.Interface

	switch tt := ect.(type) {
	case *types.Interface:
		iface = tt
	case *types.Named:
		iface, _ = tt.Underlying().(*types.Interface)
	}

	if iface == nil {
		r.eerr(entityName, fmt.Sprintf("entity with variants must be an interface (%s is not)", et.TypeName.Type().String()), e.Type.Definition)
		return 0, nil, false
	}

	vct, vrt := r.composeType(entityName, v.Type)
	if vct == nil {
		return 0, nil, false
	}

	valImpl = types.Implements(vct, iface)
	ptrImpl := types.Implements(types.NewPointer(vct), iface)
	if !valImpl && !ptrImpl {
		r.eerr(
			entityName,
			fmt.Sprintf(
				"variant type %s or its pointer doesn't implement entity interface %s",
				vt.TypeName.Type().String(),
				et.TypeName.Type().String(),
			),
			v.Type.Definition,
		)
		return 0, nil, false
	}

	tid := r.findOrCreateTypeRecursively(vrt)

	return tid, vct, valImpl
}

func (r *Resolver) resolveVariantName(entityName string, v schema.Variant) string {
	name := v.Name.V
	if name != "" {
		return name
	}

	vt, ok := r.types.goTypes[v.Type.Base.V]
	if !ok {
		r.eerr(entityName, "entity variant type base must be existing non-scalar go type", v.Type.Definition)
		return ""
	}

	return vt.TypeName.Name()
}
