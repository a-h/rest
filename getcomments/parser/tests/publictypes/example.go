package publictypes

import _ "embed"

//go:embed snapshot.json
var Expected string

// Public types should be included.
type Public struct {
	// A public field on a public type should be included.
	A string
	B string
	// c should be skipped.
	c string
}
