package jen

import . "github.com/dave/jennifer/jen"

const libImportPath = "github.com/artem328/depquery-go"

var (
	libID                         = Qual(libImportPath, "ID")
	libCandidate                  = Qual(libImportPath, "Candidate")
	libPlanner                    = Qual(libImportPath, "Planner")
	libTopologicalPlanner         = Qual(libImportPath, "TopologicalPlanner")
	libMaybeFetched               = Qual(libImportPath, "MaybeFetched")
	libFetched                    = Qual(libImportPath, "Fetched")
	libNotInStateErrorConstructor = Qual(libImportPath, "NewNotInStateError")
	iterSeq                       = Qual("iter", "Seq")
	contextContext                = Qual("context", "Context")
)
