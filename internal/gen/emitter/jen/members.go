package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
	"github.com/artem328/depquery-go/internal/gen/semantic"
	. "github.com/dave/jennifer/jen"
)

type members []Code

func warmUpMembers(p plan.Plan) members {
	m := make(members, len(p.Model.Members))

	for i, mm := range p.Model.Members {
		m[i] = resolveMember(mm)
	}

	return m
}

func (m members) Member(rcv Code, memberID semantic.MemberID) Code {
	return Add(rcv).Add(m[memberID])
}

func resolveMember(m semantic.Member) Code {
	switch m.Kind {
	case semantic.MemberKindField:
		return Dot(m.Name)
	case semantic.MemberKindMethod:
		return Dot(m.Name).Call()
	default:
		panic(fmt.Errorf("unknown member kind: %v", m.Kind))
	}
}
