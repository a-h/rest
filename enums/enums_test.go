package enums

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type stringEnum string

const (
	stringEnum1   stringEnum = ""
	stringEnum2   stringEnum = "1"
	stringEnum3   stringEnum = "2"
	incorrectType            = "invalid"
	stringEnum4   stringEnum = "3"
)

var notAConstEnum stringEnum = "invalid"

const stringEnum5 stringEnum = "4"

type intEnum int

const (
	intEnum1         intEnum = 0
	intEnum2         intEnum = 1
	intEnum3         intEnum = 2
	incorrectIntType         = "invalid"
	intEnum4         intEnum = 3
)

var notAConstIntEnum intEnum = 4

const intEnum5 intEnum = 5

type iotaIntEnum int

const (
	iotaIntEnum1 iotaIntEnum = iota
	iotaIntEnum2
	iotaIntEnum3
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		ty       reflect.Type
		expected []any
	}{
		{
			name: "string enums",
			ty:   reflect.TypeOf(stringEnum1),
			expected: []any{
				string(stringEnum1),
				string(stringEnum2),
				string(stringEnum3),
				string(stringEnum4),
				string(stringEnum5),
			},
		},
		{
			name: "int enums",
			ty:   reflect.TypeOf(intEnum1),
			expected: []any{
				int(intEnum1),
				int(intEnum2),
				int(intEnum3),
				int(intEnum4),
				int(intEnum5),
			},
		},
		{
			name: "iota int enums",
			ty:   reflect.TypeOf(iotaIntEnum1),
			expected: []any{
				int(iotaIntEnum1),
				int(iotaIntEnum2),
				int(iotaIntEnum3),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Get(tt.ty)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.expected, vals); diff != "" {
				t.Error(diff)
			}
		})
	}
}
