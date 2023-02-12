package generics

import _ "embed"

//go:embed snapshot.json
var Expected string

// Data should be included.
type Data struct {
	// AllowThis should be included.
	AllowThis string
	// String should be included.
	String DataOfT[string]
	// Int should be included.
	Int DataOfT[int]
}

// DataOfT is included in the output.
type DataOfT[T string | int] struct {
	Field T
}
