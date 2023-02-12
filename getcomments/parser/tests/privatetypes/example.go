package privatetypes

import _ "embed"

//go:embed snapshot.json
var Expected string

// private types should not be included.
type private struct {
	// A public field on a private type should not be included.
	A string
	B string
}
