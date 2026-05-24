package semantic

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"go/types"
	"strings"

	"github.com/artem328/depquery-go/internal/gen/schema"
	"golang.org/x/tools/go/packages"
)

type recursiveTypeKind uint8

const (
	recursiveTypeKindNamed recursiveTypeKind = iota + 1
	recursiveTypeKindPointer
	recursiveTypeKindGeneric
	recursiveTypeKindSlice
	recursiveTypeKindArray
	recursiveTypeKindMap
)

type recursiveType struct {
	Kind   recursiveTypeKind
	Pkg    string
	Name   string
	Size   int64
	Base   *recursiveType
	Params []recursiveType
	Key    *recursiveType
	Elem   *recursiveType
}

func newRTNamed(pkg, name string) recursiveType {
	return recursiveType{Kind: recursiveTypeKindNamed, Pkg: pkg, Name: name}
}

func newRTPointer(elem recursiveType) recursiveType {
	return recursiveType{Kind: recursiveTypeKindPointer, Elem: &elem}
}

func newRTGeneric(base recursiveType, params []recursiveType) recursiveType {
	return recursiveType{Kind: recursiveTypeKindGeneric, Base: &base, Params: params}
}

func newRTSlice(elem recursiveType) recursiveType {
	return recursiveType{Kind: recursiveTypeKindSlice, Elem: &elem}
}

func newRTArray(size int64, elem recursiveType) recursiveType {
	return recursiveType{Kind: recursiveTypeKindArray, Elem: &elem, Size: size}
}

func newRTMap(key, elem recursiveType) recursiveType {
	return recursiveType{Kind: recursiveTypeKindMap, Key: &key, Elem: &elem}
}

type goType struct {
	TypeName *types.TypeName
	Named    *types.Named
	Basic    *types.Basic
	IsBasic  bool
}

type unresolvedGoType struct {
	Raw  string
	Defs []unresolvedGoTypeDefinition
}

type unresolvedGoTypeDefinition struct {
	Entity string
	Def    schema.Definition
}

type typeIndex struct {
	unresolvedGoTypesByPkg map[string]map[string]unresolvedGoType
	packagesToResolve      []string

	goTypes map[string]goType
}

func newTypeIndex() typeIndex {
	return typeIndex{
		unresolvedGoTypesByPkg: make(map[string]map[string]unresolvedGoType),
		goTypes:                make(map[string]goType),
	}
}

func (i *typeIndex) unresolved(pkg string) map[string]unresolvedGoType {
	return i.unresolvedGoTypesByPkg[pkg]
}

func (i *typeIndex) queueRawType(raw, entity string, def schema.Definition) error {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return errors.New("the type is empty")
	}

	var pkg, name string

	period := strings.LastIndex(raw, ".")
	switch {
	case period < 0:
		name = raw
	case period == 0 || period == len(raw)-1:
		return fmt.Errorf("invalid type: `%s`", raw)
	default:
		pkg = raw[:period]
		name = raw[period+1:]
	}

	if _, ok := i.unresolvedGoTypesByPkg[pkg]; !ok {
		i.unresolvedGoTypesByPkg[pkg] = make(map[string]unresolvedGoType)
		if pkg != "" {
			i.packagesToResolve = append(i.packagesToResolve, pkg)
		}
	}

	u := i.unresolvedGoTypesByPkg[pkg][name]
	u.Raw = raw
	u.Defs = append(u.Defs, unresolvedGoTypeDefinition{
		Entity: entity,
		Def:    def,
	})
	i.unresolvedGoTypesByPkg[pkg][name] = u

	return nil
}

func (i *typeIndex) addBasicType(raw string, tn *types.TypeName, b *types.Basic) {
	i.goTypes[raw] = goType{
		TypeName: tn,
		Basic:    b,
		IsBasic:  true,
	}
}

func (i *typeIndex) addNamedType(raw string, tn *types.TypeName, n *types.Named) {
	i.goTypes[raw] = goType{
		TypeName: tn,
		Named:    n,
	}
}

func (r *Resolver) initTypeIndex() {
	for _, e := range r.schema {
		r.addRawTypeToIndex(e.Type, e.Name.V)

		for _, v := range e.Variants {
			r.addRawTypeToIndex(v.Type, e.Name.V)
		}
	}

	if len(r.errors) > 0 {
		return
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, r.types.packagesToResolve...)
	if err != nil {
		r.err(fmt.Errorf("could not load packages: %w", err))

		return
	}

	if !r.validateLoadedPackages(pkgs) {
		return
	}

	for _, pkg := range pkgs {
		tt := r.types.unresolved(pkg.Types.Path())
		scope := pkg.Types.Scope()

		r.populateTypeIndex(scope, tt)
	}

	r.populateTypeIndex(types.Universe, r.types.unresolved(""))
}

func (r *Resolver) addRawTypeToIndex(t schema.Type, entity string) {
	if err := r.types.queueRawType(t.Base.V, entity, t.Definition); err != nil {
		r.werr(err, t.Definition)
	}

	for _, p := range t.Params {
		r.addRawTypeToIndex(p, entity)
	}
}

// the code is stolen from [packages.PrintErrors] and adapted to the error model of the resolver
func (r *Resolver) validateLoadedPackages(pkgs []*packages.Package) bool {
	var n int
	errModules := make(map[*packages.Module]bool)

	for pkg := range packages.Postorder(pkgs) {
		for _, err := range pkg.Errors {
			r.err(err)
			n++
		}

		// Print pkg.Module.Error once if present.
		mod := pkg.Module
		if mod != nil && mod.Error != nil && !errModules[mod] {
			errModules[mod] = true
			r.err(errors.New(mod.Error.Err))
			n++
		}
	}

	return n == 0
}

func (r *Resolver) populateTypeIndex(scope *types.Scope, tt map[string]unresolvedGoType) {
	for t, unresolved := range tt {
		obj := scope.Lookup(t)
		if obj == nil {
			for _, def := range unresolved.Defs {
				r.eerr(def.Entity, "type "+unresolved.Raw+" couldn't be found", def.Def)
			}

			continue
		}

		tn, isTypeName := obj.(*types.TypeName)
		if !isTypeName {
			for _, def := range unresolved.Defs {
				r.eerr(def.Entity, t+" is not a type", def.Def)
			}

			continue
		}

		named, isNamed := tn.Type().(*types.Named)
		if !isNamed {
			if basic, isBasic := tn.Type().(*types.Basic); isBasic {
				r.types.addBasicType(unresolved.Raw, tn, basic)

				continue
			}

			for _, def := range unresolved.Defs {
				r.eerr(def.Entity, t+" is not a named type", def.Def)
			}

			continue
		}

		switch named.Underlying().(type) {
		case *types.Struct:
		case *types.Interface:
		case *types.Alias:
		default:
			for _, def := range unresolved.Defs {
				r.eerr(def.Entity, t+" is not either a struct or an interface or an alias", def.Def)
			}

			continue
		}

		r.types.addNamedType(unresolved.Raw, tn, named)
	}
}

func (r *Resolver) composeType(entity string, st schema.Type) (types.Type, recursiveType) {
	t, ok := r.types.goTypes[st.Base.V]
	if !ok {
		panic("could not find type `" + st.Base.V + "` in index. this is a generator bug")
	}

	if t.IsBasic {
		return t.Basic, newRTNamed("", t.Basic.Name())
	}

	accepts := t.Named.TypeParams().Len()
	provided := len(st.Params)

	rtNamed := newRTNamed(t.TypeName.Pkg().Path(), t.TypeName.Name())

	if accepts == 0 && provided == 0 {
		return t.Named, rtNamed
	}

	if accepts != provided {
		r.eerr(entity, fmt.Sprintf("type `%s` accepts %d type parameters, but %d were provided", t.Named.String(), accepts, provided), st.Definition)
		return nil, recursiveType{}
	}

	targs := make([]types.Type, 0, len(st.Params))
	rtargs := make([]recursiveType, 0, len(st.Params))

	for _, param := range st.Params {
		tparam, rtparam := r.composeType(entity, param)
		if tparam == nil {
			return nil, recursiveType{} // return early since we cannot analyze more at this point
		}

		targs = append(targs, tparam)
		rtargs = append(rtargs, rtparam)
	}

	pt, err := types.Instantiate(nil, t.Named, targs, true)
	if err != nil {
		r.eerr(entity, fmt.Sprintf("type %s couldn't be instantiated: %s", t.Named.String(), err), st.Definition)

		return nil, recursiveType{}
	}

	return pt, newRTGeneric(rtNamed, rtargs)
}

func (r *Resolver) decomposeType(t types.Type) recursiveType {
	switch tt := t.(type) {
	case *types.Basic:
		return newRTNamed("", tt.Name())
	case *types.Named:
		base := newRTNamed(
			tt.Obj().Pkg().Path(),
			tt.Obj().Name(),
		)

		targs := tt.TypeArgs()
		if targs == nil || targs.Len() == 0 {
			return base
		}

		args := make([]recursiveType, 0, targs.Len())

		for i := 0; i < targs.Len(); i++ {
			args = append(args,
				r.decomposeType(targs.At(i)),
			)
		}

		return newRTGeneric(base, args)
	case *types.Pointer:
		return newRTPointer(
			r.decomposeType(tt.Elem()),
		)
	case *types.Slice:
		return newRTSlice(
			r.decomposeType(tt.Elem()),
		)
	case *types.Array:
		return newRTArray(
			tt.Len(),
			r.decomposeType(tt.Elem()),
		)
	case *types.Map:
		return newRTMap(
			r.decomposeType(tt.Key()),
			r.decomposeType(tt.Elem()),
		)
	default:
		panic(fmt.Sprintf("unsupported type %T", t))
	}
}

func (r *Resolver) findOrCreateTypeRecursively(rt recursiveType) TypeID {
	var t Type

	switch rt.Kind {
	case recursiveTypeKindNamed:
		t = NamedType{
			Pkg:  rt.Pkg,
			Name: rt.Name,
		}
	case recursiveTypeKindPointer:
		t = PointerType{
			Elem: r.findOrCreateTypeRecursively(*rt.Elem),
		}
	case recursiveTypeKindGeneric:
		g := GenericType{
			Base:   r.findOrCreateTypeRecursively(*rt.Base),
			Params: make([]TypeID, 0, len(rt.Params)),
		}

		for _, p := range rt.Params {
			g.Params = append(g.Params, r.findOrCreateTypeRecursively(p))
		}

		t = g
	case recursiveTypeKindSlice:
		t = SliceType{Elem: r.findOrCreateTypeRecursively(*rt.Elem)}
	case recursiveTypeKindArray:
		t = ArrayType{Size: rt.Size, Elem: r.findOrCreateTypeRecursively(*rt.Elem)}
	case recursiveTypeKindMap:
		t = MapType{Key: r.findOrCreateTypeRecursively(*rt.Key), Elem: r.findOrCreateTypeRecursively(*rt.Elem)}
	default:
		panic(fmt.Errorf("unknown recursive type kind %d", rt.Kind))
	}

	return r.findOrCreateType(t)
}

func (r *Resolver) findOrCreateType(t Type) TypeID {
	h := r.hashType(t)
	id, ok := r.seenTypes[h]
	if !ok {
		id = TypeID(len(r.model.Types))
		r.seenTypes[h] = id
		r.model.Types = append(r.model.Types, t)
	}

	return id
}

func (r *Resolver) hashType(t Type) [16]byte {
	seed := make([]byte, 0, 64)

	const sep = '\x00'

	switch tt := t.(type) {
	case NamedType:
		seed = append(seed, "named"...)
		seed = append(seed, sep)
		seed = append(seed, tt.Pkg...)
		seed = append(seed, sep)
		seed = append(seed, tt.Name...)
	case PointerType:
		seed = append(seed, "pointer"...)
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint32(seed, uint32(tt.Elem))
	case GenericType:
		seed = append(seed, "generic"...)
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint32(seed, uint32(tt.Base))
		for _, p := range tt.Params {
			seed = append(seed, sep)
			seed = binary.BigEndian.AppendUint32(seed, uint32(p))
		}
	case SliceType:
		seed = append(seed, "slice"...)
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint32(seed, uint32(tt.Elem))
	case ArrayType:
		seed = append(seed, "array"...)
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint64(seed, uint64(tt.Size))
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint32(seed, uint32(tt.Elem))
	case MapType:
		seed = append(seed, "map"...)
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint32(seed, uint32(tt.Key))
		seed = append(seed, sep)
		seed = binary.BigEndian.AppendUint32(seed, uint32(tt.Elem))
	default:
		panic(fmt.Errorf("unknown semantic type %T", t))
	}

	return md5.Sum(seed)
}

func (r *Resolver) resolveUnderlyingTypeKind(t types.Type) UnderlyingTypeKind {
	for {
		if t == nil {
			return 0
		}

		switch tt := t.(type) {
		case *types.Basic:
			switch tt.Kind() {
			case types.Int:
				return UnderlyingTypeInt
			case types.Int8:
				return UnderlyingTypeInt8
			case types.Int16:
				return UnderlyingTypeInt16
			case types.Int32:
				return UnderlyingTypeInt32
			case types.Int64:
				return UnderlyingTypeInt64
			case types.Uint:
				return UnderlyingTypeUint
			case types.Uint8:
				return UnderlyingTypeUint8
			case types.Uint16:
				return UnderlyingTypeUint16
			case types.Uint32:
				return UnderlyingTypeUint32
			case types.Uint64:
				return UnderlyingTypeUint64
			case types.Float32:
				return UnderlyingTypeFloat32
			case types.Float64:
				return UnderlyingTypeFloat64
			case types.String:
				return UnderlyingTypeString
			case types.UntypedInt:
				return UnderlyingTypeInt
			case types.UntypedRune:
				return UnderlyingTypeInt32
			case types.UntypedFloat:
				return UnderlyingTypeFloat32
			case types.UntypedString:
				return UnderlyingTypeString
			default:
				return 0
			}
		case *types.Array:
			b, ok := tt.Elem().(*types.Basic)
			if !ok || b.Kind() != types.Byte {
				return 0
			}

			return UnderlyingTypeByteArray
		default:
			ut := t.Underlying()
			if t == ut {
				return 0
			}
			t = ut
		}
	}
}
