package pointers

import _ "embed"

//go:embed snapshot.json
var Expected string

// Public should be included.
type Public struct {
	// A pointer to a pointer should be included.
	A **string
}
