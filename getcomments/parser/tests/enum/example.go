package enum

import _ "embed"

//go:embed snapshot.json
var Expected string

type Public struct {
	A StringEnum
	B IntEnum
}

// StringEnum should be included.
type StringEnum string

const (
	// StringEnum comment.
	StringEnumA StringEnum = "A"
	StringEnumB StringEnum = "B"
	StringEnumC StringEnum = "C"
)

// IntEnum should be included.
type IntEnum int

const (
	// IntEnum0 should be included.
	IntEnum0 IntEnum = iota
	IntEnum1
	IntEnum2
)
