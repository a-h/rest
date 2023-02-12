package functions

import _ "embed"

//go:embed snapshot.json
var Expected string

// ThisShouldBeIgnored because a function can't be part of schema.
func ThisShouldBeIgnored() {
}

// asShouldThis should be ignored, because it's not exported too.
func asShouldThis() {
}

// Data should be included.
type Data struct {
	// A should also be included.
	A string
}

// IgnoreMe because I'm a method on a type.
func (d Data) IgnoreMe() {
}

func (d Data) andMeToo() {
}

func (d *Data) Same() {
}

// DontIgnoreMe just because I'm further down the file.
type DontIgnoreMe struct {
	Please string
}
