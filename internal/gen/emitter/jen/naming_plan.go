package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type planNaming struct {
	Interface                    string
	NewInstanceMethod            string
	Struct                       string
	FieldFetchContextConstructor string
	FieldResolvers               string
}

func (n *planNaming) warmUp(plan.Plan) {
	n.Interface = "Plan"
	n.NewInstanceMethod = "NewInstance"
	n.Struct = "plan"
	n.FieldFetchContextConstructor = "ctxConstructor"
	n.FieldResolvers = "resolvers"
}
