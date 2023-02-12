package functiontypes

import _ "embed"

//go:embed snapshot.json
var Expected string

// Data should be documented.
type Data struct {
	// The function here could be ignored.
	A func(test string)
	// As could this.
	B FuncType
	// Non exported fields are not included in the output.
	c func()
}

// No need to document this.
type FuncType func()
