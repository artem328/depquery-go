package schema

type MemberType uint8

const (
	MemberTypeInvalid MemberType = iota
	MemberTypeField
	MemberTypeMethod
)

type Relation struct {
	Name       string
	Entity     *Entity
	ReversedBy *IDMember
}

func (r Relation) IsReversed() bool {
	return r.ReversedBy != nil
}

type IDMember struct {
	Name    string
	RcvType MemberType
	Type    Type
}

type Type struct {
	Package string
	Name    string
	Wrapper string
	Params  []Type
}

type Entity struct {
	Name      string
	Type      Type
	ID        IDMember
	Relations []Relation
	Variants  []EntityVariant
}

type EntityVariant struct {
	Name      string
	Type      Type
	Relations []Relation
}
