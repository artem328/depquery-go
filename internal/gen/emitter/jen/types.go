package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
	"github.com/artem328/depquery-go/internal/gen/semantic"
	. "github.com/dave/jennifer/jen"
)

type types []Code

func warmUpTypes(p plan.Plan) types {
	tt := make(types, len(p.Model.Types))

	for id := range p.Model.Types {
		tid := semantic.TypeID(id)

		tt[tid] = resolveGoType(tid, p.Model.Types, tt)
	}

	return tt
}

func resolveGoType(id semantic.TypeID, st []semantic.Type, tt types) Code {
	if tt[id] != nil {
		return tt[id]
	}

	t := st[id]

	switch stt := t.(type) {
	case semantic.NamedType:
		return Qual(stt.Pkg, stt.Name)
	case semantic.PointerType:
		return Op("*").Add(resolveGoType(stt.Elem, st, tt))
	case semantic.GenericType:
		params := make([]Code, len(stt.Params))
		for i, p := range stt.Params {
			params[i] = resolveGoType(p, st, tt)
		}
		base := resolveGoType(stt.Base, st, tt)
		return Add(base).Types(params...)
	case semantic.SliceType:
		return Index().Add(resolveGoType(stt.Elem, st, tt))
	case semantic.ArrayType:
		return Index(Lit(stt.Size)).Add(resolveGoType(stt.Elem, st, tt))
	case semantic.MapType:
		return Map(resolveGoType(stt.Key, st, tt)).Add(resolveGoType(stt.Elem, st, tt))
	default:
		panic(fmt.Errorf("unknown type %T", t))
	}
}
