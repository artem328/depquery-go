package generator

import "github.com/dave/jennifer/jen"

const libPkg = "github.com/artem328/depquery-go"

var (
	libID                = jen.Qual(libPkg, "ID")
	libMaybeFetched      = jen.Qual(libPkg, "MaybeFetched")
	libFetched           = jen.Qual(libPkg, "Fetched")
	libCandidate         = jen.Qual(libPkg, "Candidate")
	libNotInStateErrCtor = jen.Qual(libPkg, "NewNotInStateError")
	iterSeq              = jen.Qual("iter", "Seq")
	contextCtx           = jen.Qual("context", "Context")
)
