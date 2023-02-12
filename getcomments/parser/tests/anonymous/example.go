package anonymous

import _ "embed"

//go:embed snapshot.json
var Expected string

// Data should be included.
type Data struct {
	// A should be included.
	A struct {
		// B should be included.
		B string
	}
}
