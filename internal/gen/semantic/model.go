package semantic

import "strconv"

type Model struct {
	Types     []Type
	Entities  []Entity
	Variants  []Variant
	Members   []Member
	Relations []Relation
}

type TypeID int

func (id TypeID) String() string {
	return "semantic.TypeID(" + strconv.Itoa(int(id)) + ")"
}

type Type interface {
	isType()
}

type NamedType struct {
	Pkg  string
	Name string
}

func (NamedType) isType() {}

type PointerType struct {
	Elem TypeID
}

func (PointerType) isType() {}

type GenericType struct {
	Base   TypeID
	Params []TypeID
}

func (GenericType) isType() {}

type SliceType struct {
	Elem TypeID
}

func (SliceType) isType() {}

type ArrayType struct {
	Size int64
	Elem TypeID
}

func (ArrayType) isType() {}

type MapType struct {
	Key  TypeID
	Elem TypeID
}

func (MapType) isType() {}

type MemberID int

func (id MemberID) String() string {
	return "semantic.MemberID(" + strconv.Itoa(int(id)) + ")"
}

type MemberKind uint8

const (
	MemberKindField MemberKind = iota + 1
	MemberKindMethod
)

type Member struct {
	Name string
	Type TypeID
	Kind MemberKind
}

type EntityID int

func (id EntityID) String() string {
	return "semantic.EntityID(" + strconv.Itoa(int(id)) + ")"
}

type Entity struct {
	Name     string
	Type     TypeID
	IDMember MemberID
}

type VariantID int

func (id VariantID) String() string {
	return "semantic.VariantID(" + strconv.Itoa(int(id)) + ")"
}

type Variant struct {
	Name            string
	Entity          EntityID
	Type            TypeID
	ValueAssignable bool
}

type Relation struct {
	Name       string
	From       EntityID
	Variant    Optional[VariantID]
	To         EntityID
	ReversedBy Optional[MemberID]
}

type Optional[T any] struct {
	V   T
	Set bool
}
