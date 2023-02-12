package docs

import _ "embed"

//go:embed snapshot.json
var Expected string

// Struct documentation.
type Data struct {
	// Field documentation.
	A string

	// Field documentation
	Cats CatType
}

// CatType is a type of cat.
type CatType string

// Looks grumpy, but is warm and caring.
const CatTypePersian CatType = "persian"

// Some say that these are relatively aggressive.
const CatTypeManx CatType = "manx"

const (
	// CatTypeLion is a lion.
	CatTypeLion CatType = "lion"
	// The king of big cats.
	CatTypeTiger CatType = "tiger"
)
