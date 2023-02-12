package unsupported

// Type should be included.
type Type struct {
	// MapOfStringToString should be included.
	MapOfStringToString map[string]string
	// MapOfMapsToMaps should be included.
	MapOfMapsToMaps map[string]map[string]string
	// MapOfMapValue should be included.
	MapOfMapValue map[string]MapValue
}

// MapValue should be included.
type MapValue struct {
	// FieldA should be included.
	FieldA string
}
