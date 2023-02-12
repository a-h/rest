package chans

import _ "embed"

//go:embed snapshot.json
var Expected string

// Data should be included.
type Data struct {
	// AndIgnoreThisToo could be ignored, but isn't.
	AndIgnoreThisToo chan string
	// AndThisArrayOfChannels could also be ignored.
	AndThisArrayOfChannels []chan string
	// AndThisAlias could also be ignored.
	AndThisAlias ChanType
}

// IgnoreThisChannel could be ignored, since channels can't be part of schema.
var IgnoreThisChannel chan string

// ChanType could be ignored, since chanels can't be part of schema.
type ChanType chan string
