package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/aviva-verde/rest/getcomments/parser"
	"github.com/aviva-verde/rest/getcomments/parser/tests/anonymous"
	"github.com/aviva-verde/rest/getcomments/parser/tests/chans"
	"github.com/aviva-verde/rest/getcomments/parser/tests/docs"
	"github.com/aviva-verde/rest/getcomments/parser/tests/enum"
	"github.com/aviva-verde/rest/getcomments/parser/tests/functions"
	"github.com/aviva-verde/rest/getcomments/parser/tests/functiontypes"
	"github.com/aviva-verde/rest/getcomments/parser/tests/pointers"
	"github.com/aviva-verde/rest/getcomments/parser/tests/privatetypes"
	"github.com/aviva-verde/rest/getcomments/parser/tests/publictypes"
	"github.com/google/go-cmp/cmp"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		expected string
	}{
		{
			name:     "private structs are ignored",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/privatetypes",
			expected: privatetypes.Expected,
		},
		{
			name:     "public structs are included",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/publictypes",
			expected: publictypes.Expected,
		},
		{
			name:     "string and integer enums are supported",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/enum",
			expected: enum.Expected,
		},
		{
			name:     "pointers to pointers become a single pointer",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/pointers",
			expected: pointers.Expected,
		},
		{
			name:     "functions and method receivers are ignored",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/functions",
			expected: functions.Expected,
		},
		{
			name:     "fields of type channel are ignored",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/chans",
			expected: chans.Expected,
		},
		{
			name:     "anonymous structs are ignored",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/anonymous",
			expected: anonymous.Expected,
		},
		{
			name:     "function fields and function types are ignored",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/functiontypes",
			expected: functiontypes.Expected,
		},
		{
			name:     "stuct, field and constant comments are extracted",
			pkg:      "github.com/aviva-verde/rest/getcomments/parser/tests/docs",
			expected: docs.Expected,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			m, err := parser.Get(test.pkg)
			if err != nil {
				t.Fatalf("failed to get model %q: %v", test.pkg, err)
			}

			var expected map[string]string
			err = json.Unmarshal([]byte(test.expected), &expected)
			if err != nil {
				t.Fatalf("snapshot load failed: %v", err)
			}

			if diff := cmp.Diff(expected, m); diff != "" {
				t.Error(diff)
			}
		})
	}
}
