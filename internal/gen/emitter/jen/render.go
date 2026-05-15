package jen

import (
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
