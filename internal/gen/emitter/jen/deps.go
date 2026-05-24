package jen

import . "github.com/dave/jennifer/jen"

const libImportPath = "github.com/artem328/depquery-go"

var (
	libID                         = Qual(libImportPath, "ID")
	libCandidate                  = Qual(libImportPath, "Candidate")
	libPlanner                    = Qual(libImportPath, "Planner")
	libBFSPlanner                 = Qual(libImportPath, "BFSPlanner")
	libMaybeFetched               = Qual(libImportPath, "MaybeFetched")
	libFetched                    = Qual(libImportPath, "Fetched")
	libNotInStateErrorConstructor = Qual(libImportPath, "NewNotInStateError")
	libSyntheticID                = Qual(libImportPath, "SyntheticID")
	libIBytes                     = Qual(libImportPath, "IBytes")
	libI64Bytes                   = Qual(libImportPath, "I64Bytes")
	libI32Bytes                   = Qual(libImportPath, "I32Bytes")
	libI16Bytes                   = Qual(libImportPath, "I16Bytes")
	libI8Bytes                    = Qual(libImportPath, "I8Bytes")
	libF64Bytes                   = Qual(libImportPath, "F64Bytes")
	libF32Bytes                   = Qual(libImportPath, "F32Bytes")
	libSBytes                     = Qual(libImportPath, "SBytes")
	libInstanceConfig             = Qual(libImportPath, "InstanceConfig")
	libInstanceOption             = Qual(libImportPath, "InstanceOption")
	libConcurrentExecutor         = Qual(libImportPath, "ConcurrentExecutor")
	libCompilerConfig             = Qual(libImportPath, "CompilerConfig")
	libCompilerOption             = Qual(libImportPath, "CompilerOption")
	libTask                       = Qual(libImportPath, "Task")
	iterSeq                       = Qual("iter", "Seq")
	iterSeq2                      = Qual("iter", "Seq2")
	contextContext                = Qual("context", "Context")
)
