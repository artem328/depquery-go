package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/semantic"
	. "github.com/dave/jennifer/jen"
)

func block(s Code) Code {
	return Line().Add(s).Line()
}

var valuesMultilineOpts = Options{
	Open:      "{",
	Close:     "}",
	Separator: ",",
	Multi:     true,
}

func valuesMultiline(statements ...Code) *Statement {
	return Custom(valuesMultilineOpts, statements...)
}

func appendSlice(slice Code, vals ...Code) Code {
	appendArgs := make([]Code, 0, len(vals)+1)
	appendArgs = append(appendArgs, slice)
	appendArgs = append(appendArgs, vals...)

	return Add(slice).Op("=").Append(appendArgs...)
}

func generateTypeToBytes(k semantic.UnderlyingTypeKind, v Code) Code {
	switch k {
	case semantic.UnderlyingTypeUint, semantic.UnderlyingTypeInt:
		return Add(libIBytes).Call(v)
	case semantic.UnderlyingTypeUint8, semantic.UnderlyingTypeInt8:
		return Add(libI8Bytes).Call(v)
	case semantic.UnderlyingTypeUint16, semantic.UnderlyingTypeInt16:
		return Add(libI16Bytes).Call(v)
	case semantic.UnderlyingTypeUint32, semantic.UnderlyingTypeInt32:
		return Add(libI32Bytes).Call(v)
	case semantic.UnderlyingTypeUint64, semantic.UnderlyingTypeInt64:
		return Add(libI64Bytes).Call(v)
	case semantic.UnderlyingTypeFloat32:
		return Add(libF32Bytes).Call(v)
	case semantic.UnderlyingTypeFloat64:
		return Add(libF64Bytes).Call(v)
	case semantic.UnderlyingTypeString:
		return Add(libSBytes).Call(v)
	case semantic.UnderlyingTypeByteArray:
		return Add(v).Index(Empty(), Empty())
	default:
		panic(fmt.Errorf("unknown underlying type kind %v", k))
	}
}
