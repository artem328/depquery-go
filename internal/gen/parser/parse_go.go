package parser

import (
	"fmt"
	"go/types"
	"iter"
	"strconv"

	"golang.org/x/tools/go/packages"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

type goParsedType struct {
	typeName *types.TypeName
	named    *types.Named
	basic    *types.Basic
	isBasic  bool
}

type goParsedEntity struct {
	methods  func() iter.Seq[*types.Func]
	str      *types.Struct
	isStruct bool
	typ      types.Type
	idType   types.Type
}

func (p *Parser) parseGo() {
	defer p.goCleanup()

	p.goExpandTypes()

	if _, ok := p.errors(); ok {
		return
	}

	p.goBuildTypesMap()

	if _, ok := p.errors(); ok {
		return
	}

	p.goResolveEntities()

	if _, ok := p.errors(); ok {
		return
	}

	p.goResolveReversedMembers()
}

func (p *Parser) goExpandTypes() {
	for _, e := range p.entities {
		if e.Type == nil {
			p.gerr(e.Name, "no type provided", e.source)

			continue
		}

		if !p.expandTypes(e) {
			continue
		}
	}
}

func (p *Parser) goBuildTypesMap() bool {
	i := newGoTypeIndex()

	for _, e := range p.entities {
		i.add(e.Name, e.Type)

		for _, v := range e.Variants {
			i.add(e.Name, v.Type)
		}
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, i.packages()...)
	if err != nil {
		p.gerr("", fmt.Sprintf("could not load packages: %v", err), nil)

		return false
	}

	ok := true

	for _, pkg := range pkgs {
		pkgName := pkg.Types.Path()
		tt := i.types(pkgName)
		scope := pkg.Types.Scope()

		ok = p.goAddTypes(pkgName, scope, tt) && ok
	}

	if !p.goAddTypes("", types.Universe, i.types("")) || !ok {
		return false
	}

	return p.goValidateEntityTypeParams()
}

//nolint:cyclop // looks fine
func (p *Parser) goAddTypes(pkg string, scope *types.Scope, tt map[string][]goTypeIndexEntry) bool {
	ok := true

	for t, defs := range tt {
		key := (&goType{Package: pkg, Name: t}).String()

		obj := scope.Lookup(t)
		if obj == nil {
			for _, def := range defs {
				p.gerr(def.entity, fmt.Sprintf("type %s couldn't be found", key), def.t.source)
			}

			ok = false

			continue
		}

		tn, isTypeName := obj.(*types.TypeName)
		if !isTypeName {
			for _, def := range defs {
				p.gerr(def.entity, t+" must be a type", def.t.source)
			}

			ok = false

			continue
		}

		named, isNamed := tn.Type().(*types.Named)
		if !isNamed {
			basic, isBasic := tn.Type().(*types.Basic)
			if isBasic {
				p.goTypes[key] = goParsedType{typeName: tn, basic: basic, isBasic: true}

				continue
			}

			for _, def := range defs {
				p.gerr(def.entity, t+" must be a named type", def.t.source)
			}

			ok = false

			continue
		}

		switch named.Underlying().(type) {
		case *types.Struct:
		case *types.Interface:
		case *types.Alias:
		default:
			for _, def := range defs {
				p.gerr(def.entity, t+" must be a struct, an interface, or an alias", def.t.source)
			}

			ok = false

			continue
		}

		p.goTypes[key] = goParsedType{typeName: tn, named: named}
	}

	return ok
}

func (p *Parser) goValidateEntityTypeParams() bool {
	ok := true

	for _, e := range p.entities {
		ok = p.goValidateTypeParams(e.Name, e.Type) && ok

		for _, v := range e.Variants {
			ok = p.goValidateTypeParams(e.Name, v.Type) && ok
		}
	}

	return ok
}

func (p *Parser) goComposeParametrizedType(entity string, gt *goType) (types.Type, bool) {
	t, ok := p.goTypes[gt.String()]
	if !ok {
		p.gerr("", fmt.Sprintf("type %s is not indexed. this is a bug", gt.String()), nil)

		return nil, false
	}

	if t.isBasic {
		return t.basic, true
	}

	if t.named.TypeParams().Len() == 0 && len(gt.Params) == 0 {
		return t.named, true
	}

	if t.named.TypeParams().Len() != len(gt.Params) {
		if t.named.TypeParams().Len() > 0 {
			p.gerr(
				entity,
				fmt.Sprintf(
					"type %s expects %d type parameters, but %d provided",
					t.named.String(),
					t.named.TypeParams().Len(),
					len(gt.Params),
				),
				gt.source,
			)
		} else {
			p.gerr(
				entity,
				fmt.Sprintf("type %s does not have type parameters, but they were provided", t.named.String()),
				gt.source,
			)
		}

		return nil, false
	}

	targs := make([]types.Type, 0, len(gt.Params))

	for _, param := range gt.Params {
		got, ok := p.goComposeParametrizedType(entity, param)
		if !ok {
			return nil, false
		}

		targs = append(targs, got)
	}

	pt, err := types.Instantiate(nil, t.named, targs, true)
	if err != nil {
		p.gerr(entity, fmt.Sprintf("type %s couldn't be instantiated: %s", t.named.String(), err), gt.source)

		return nil, false
	}

	return pt, true
}

func (p *Parser) goValidateTypeParams(entity string, gt *goType) bool {
	_, ok := p.goComposeParametrizedType(entity, gt)

	return ok
}

func (p *Parser) goResolveEntities() {
	for _, e := range p.entities {
		pt, ok := p.goTypes[e.Type.String()]
		if !ok {
			p.gerr(e.Name, "entity type must be existing non-scalar go type", e.Type.source)

			continue
		}

		pe := p.goFindEntity(pt, e)
		if pe.typ == nil {
			continue
		}

		p.goEntityVariants(pe.typ, e)

		e.ID, pe.idType = p.goResolveMember(e.Name, pe, e.ID)

		p.goEntityTypes[e.Name] = pe
	}
}

func (p *Parser) goResolveMember(entity string, e goParsedEntity, i id) (id, types.Type) {
	if t, rt, found := p.goFindMethodMember(entity, e.methods, i); t != nil {
		i.Type = t
		i.MemberType = schema.MemberTypeMethod

		return i, rt
	} else if found {
		return i, nil
	}

	if e.isStruct {
		if t, tt, found := p.goFindFieldMember(entity, e.str, i); t != nil {
			i.Type = t
			i.MemberType = schema.MemberTypeField

			return i, tt
		} else if found {
			return i, nil
		}
	}

	p.gerr(entity, fmt.Sprintf("member %s not found in %s", i.Member, e.typ.String()), i.source)

	return i, nil
}

func (p *Parser) goFindEntity(
	pt goParsedType,
	e *entity,
) (pe goParsedEntity) {
	if pt.isBasic {
		p.gerr(e.Name, "entity type cannot be scalar type", e.Type.source)

		return pe
	}

	pe.typ = pt.named.Underlying()

	for {
		if a, ok := pe.typ.(*types.Alias); ok {
			pe.typ = a.Underlying()

			continue
		}

		break
	}

	switch tt := pe.typ.(type) {
	case *types.Struct:
		pe.str = tt
		pe.isStruct = true
		pe.methods = func() iter.Seq[*types.Func] {
			return structMethodIter(pt.named)
		}
	case *types.Interface:
		pe.methods = tt.Methods
	default:
		p.gerr(e.Name, e.Type.String()+" is neither struct nor interface", e.Type.source)

		return pe
	}

	if !pt.typeName.Exported() {
		p.gerr(e.Name, fmt.Sprintf("type %s is not exported", e.Type.String()), e.Type.source)

		pe = goParsedEntity{} // reset
	}

	return pe
}

func (p *Parser) goFindMethodMember(
	entity string,
	methods func() iter.Seq[*types.Func],
	member id,
) (_ *goType, rt types.Type, found bool) {
	for method := range methods() {
		if method.Name() != member.Member {
			continue
		}

		if !method.Exported() {
			p.gerr(
				entity,
				fmt.Sprintf("method %s of %s is not exported", method.Name(), method.Signature().Recv().Type().String()),
				member.source,
			)

			return nil, nil, true
		}

		res := method.Signature().Results()
		if res.Len() != 1 {
			p.gerr(
				entity,
				fmt.Sprintf(
					"method %s of %s must return exactly one value. got %d",
					method.Name(),
					method.Signature().Recv().Type().String(),
					res.Len(),
				),
				member.source,
			)

			return nil, nil, true
		}

		rt = res.At(0).Type()

		return p.goExtractIDType(entity, rt, member), rt, true
	}

	return nil, nil, false
}

func (p *Parser) goFindFieldMember(
	entity string,
	structType *types.Struct,
	member id,
) (_ *goType, t types.Type, found bool) {
	for field := range structType.Fields() {
		if field.Name() != member.Member {
			continue
		}

		if !field.Exported() {
			p.gerr(entity, fmt.Sprintf("field %s of %s is not exported", field.Name(), structType.String()), member.source)

			return nil, nil, true
		}

		t = field.Type()

		return p.goExtractIDType(entity, t, member), t, true
	}

	return nil, t, false
}

func (p *Parser) goExtractIDType(entity string, t types.Type, id id) *goType {
	if !types.Comparable(t) {
		p.gerr(entity, fmt.Sprintf("type %s is not suitable for ID as its incomparable", t.String()), id.source)

		return nil
	}

	wrapper := make([]byte, 0)

Unwrap:
	for {
		switch w := t.(type) {
		case *types.Pointer:
			wrapper = append(wrapper, '*')
			t = w.Elem()
		case *types.Array:
			wrapper = append(wrapper, "["+strconv.FormatInt(w.Len(), 10)+"]"...)
			t = w.Elem()
		default:
			break Unwrap
		}
	}

	var gt goType

	switch tt := t.(type) {
	case *types.Basic:
		gt.Name = tt.Name()
	case *types.Named:
		gt.Name = tt.Obj().Name()
		gt.Package = tt.Obj().Pkg().Path()
	default:
		p.gerr(entity, "unsupported ID type. likely a bug in generator", id.source)

		return nil
	}

	gt.Wrapper = string(wrapper)

	return &gt
}

func (p *Parser) goEntityVariants(t types.Type, e *entity) {
	if len(e.Variants) == 0 {
		return
	}

	i, ok := t.(*types.Interface)
	if !ok {
		p.gerr(e.Name, "entity with variants must be an interface", e.Type.source)

		return
	}

	for _, v := range e.Variants {
		pt, ok := p.goTypes[v.Type.String()]
		if !ok {
			p.gerr(e.Name, "entity variant type must be existing non scalar type", e.Type.source)

			continue
		}

		ptr := types.NewPointer(pt.named)
		implements := types.Implements(pt.named, i)
		ptrImplements := types.Implements(ptr, i)

		if !implements && !ptrImplements {
			p.gerr(
				e.Name,
				fmt.Sprintf(
					"variant type %s or its pointer don't implement entity interface %s",
					v.Type.String(),
					e.Type.String(),
				),
				v.Type.source,
			)

			continue
		}

		if ptrImplements {
			v.Type.Wrapper = "*" + v.Type.Wrapper
		}
	}
}

func (p *Parser) goResolveReversedMembers() {
	for _, e := range p.entities {
		for i, r := range e.Relations {
			e.Relations[i].Reversed = p.goResolveReversedMemberForRelation(e.Name, r)
		}

		for _, v := range e.Variants {
			for i, r := range v.Relations {
				e.Relations[i].Reversed = p.goResolveReversedMemberForRelation(e.Name, r)
			}
		}
	}
}

func (p *Parser) goResolveReversedMemberForRelation(entity string, r relation) *id {
	if r.Reversed == nil {
		return nil
	}

	ee, ok := p.goEntityTypes[r.Entity]
	if !ok {
		p.gerr(entity, "unresolved relation entity. this is most likely bug", r.source)

		return r.Reversed
	}

	rev, t := p.goResolveMember(entity, ee, *r.Reversed)
	if t == nil {
		// no need to report here, already reported by goResolveMember
		return &rev
	}

	if t != ee.idType {
		p.gerr(
			entity,
			fmt.Sprintf("reversed member type %s does not match with referenced entity ID type %s", t.String(), ee.idType),
			rev.source,
		)
	}

	return &rev
}

func (p *Parser) expandTypes(e *entity) bool {
	ok := p.expandType(e.Type, e.Name)

	for _, v := range e.Variants {
		ok = p.expandType(v.Type, e.Name) && ok
	}

	return ok
}

func (p *Parser) expandType(t *goType, entityName string) bool {
	ok := true

	for _, param := range t.Params {
		ok = p.expandType(param, entityName) && ok
	}

	if t.IsSet() {
		return ok
	}

	if !t.IsRaw() {
		p.gerr(entityName, "no type provided", t.source)

		return false
	}

	if err := t.Parse(); err != nil {
		p.gerr(entityName, "failed to parse type: "+err.Error(), t.source)

		return false
	}

	return ok
}

func (p *Parser) goCleanup() {
	clear(p.goTypes)
	clear(p.goEntityTypes)
}

func structMethodIter(named *types.Named) iter.Seq[*types.Func] {
	iters := [...]func() iter.Seq[*types.Selection]{
		func() iter.Seq[*types.Selection] {
			return types.NewMethodSet(named).Methods()
		},
		func() iter.Seq[*types.Selection] {
			return types.NewMethodSet(types.NewPointer(named)).Methods()
		},
	}

	return func(yield func(*types.Func) bool) {
		for _, it := range iters {
			for method := range it() {
				if f, ok := method.Obj().(*types.Func); ok && !yield(f) {
					return
				}
			}
		}
	}
}

type goTypeIndexEntry struct {
	entity string
	t      *goType
}

type goTypeIndex struct {
	byPkg map[string]map[string][]goTypeIndexEntry
	pkgs  []string
}

func newGoTypeIndex() *goTypeIndex {
	return &goTypeIndex{byPkg: make(map[string]map[string][]goTypeIndexEntry)}
}

func (i *goTypeIndex) packages() []string {
	return i.pkgs
}

func (i *goTypeIndex) add(entity string, t *goType) []string {
	var newPkgs []string

	if _, ok := i.byPkg[t.Package]; !ok {
		i.byPkg[t.Package] = make(map[string][]goTypeIndexEntry)
		if t.Package != "" {
			i.pkgs = append(i.pkgs, t.Package)
		}
	}

	i.byPkg[t.Package][t.Name] = append(i.byPkg[t.Package][t.Name], goTypeIndexEntry{entity, t})

	for _, param := range t.Params {
		i.add(entity, param)
	}

	return newPkgs
}

func (i *goTypeIndex) types(pkg string) map[string][]goTypeIndexEntry {
	return i.byPkg[pkg]
}
